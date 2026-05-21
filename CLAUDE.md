# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project identity

This repository is **Loka Chain** — an EVM-compatible Cosmos SDK L1 forked from Evmos. Despite the directory name `evmos/`, the import paths, binary name, and bech32 prefix have been rebranded to `loka`:

- Binary: `lokad` (built into `$GOPATH/bin` via `make install` or into `./build/` via `make build`)
- Module / version path: `github.com/loka/loka` (see `Makefile` `PACKAGE_NAME` and `ldflags`)
- Native denom: `aloka` (18 decimals)
- Bech32 prefix: `loka` (account), `lokavaloper`, etc.
- Default local chain-id: `loka_567002-1` (EIP-155 chain id `567002`)
- Key algorithm: `eth_secp256k1` (Ethereum-style keys)

Many in-tree paths and Go files still reference `evmos` (e.g. `evmtypes`, `evmoscmd`, the parent `app/app.go` package, `cmd/evmosd` historical content). Treat `lokad` as the user-facing name and `loka` as the chain identity, but **do not blanket-rename `evmos` symbols in upstream-mirrored code** — much of `x/`, `rpc/`, `server/`, `types/`, `store/`, `memiavl/`, `versiondb/` is forked SDK code where renaming the package would break compilation against external Cosmos SDK consumers.

## Build, run, test

```sh
make install          # build lokad into $GOPATH/bin (default for local dev)
make build            # build into ./build/lokad
make build-rocksdb    # CGO build with rocksdb backend (uses scripts/install_librocksdb.sh)
make test             # unit tests, -race, 15m timeout, excludes ./tests/e2e
make test-race        # race detector across non-simulation packages
make test-unit-cover  # coverage report
make test-e2e         # docker-based e2e (builds image unless TARGET_VERSION set)
make test-memiavl    # tests inside ./memiavl (separate go module)
make test-store      # tests inside ./store
make test-versiondb  # tests inside ./versiondb (requires rocksdb build tag)
make lint             # golangci-lint + solhint on contracts/
make lint-fix         # golangci-lint --fix
make release-check    # validate .goreleaser.yml against goreleaser v2 schema (fast, CI)
make release-dry-run  # full cross-build via pinned goreleaser-cross image (slow)
```

Build tags: `netgo objstore` always, plus `ledger` (auto-detects gcc), plus optional `rocksdb grocksdb_clean_link`, `pebbledb`, `boltdb`, `cleveldb` selected via `COSMOS_BUILD_OPTIONS=rocksdb` etc. All Go test invocations in the Makefile pass `-tags=objstore` — running `go test ./...` directly without this tag will skip or break object-store backed code paths.

Run a single test:
```sh
go test -tags=objstore -run TestFoo ./path/to/pkg
```

### Running a local node

```sh
./local_node.sh           # interactive: prompts before wiping the existing config dir
./local_node.sh nohup     # background, logs to nohup.out
./local_node.sh pending   # slow consensus timeouts for debugging pending-tx behavior
```

`local_node.sh` hardcodes `HOMEDIR=/Users/blake/work/nagara/code/evmos/.vscode/lokad-config` — edit before running on a fresh machine. It re-runs `make install` (or `make build-rocksdb` if `COSMOS_BUILD_OPTIONS=rocksdb`) at the top of every run, then rewrites genesis to use `aloka`, enables `block-executor = "block-stm"` and `memiavl.enable = true` in `app.toml`, and starts JSON-RPC on `:8545` / WS on `:8546`.

`init_validators.sh <ip1> <ip2> <ip3> <ip4>` builds a 4-validator production-style network; `start_node_archive.sh` runs an archive-mode full node. `show_tps.sh` polls the local node for TPS.

## High-level architecture

This is a Cosmos SDK chain wired together in `app/app.go`. The big-picture pieces that span multiple files:

### EVM on Cosmos SDK

The chain runs the EVM as an SDK module, not as a separate L2. The relevant in-tree modules live under `x/`:

- `x/evm` — the EVM state machine, statedb, precompiles, keeper. EVM txs are wrapped as `MsgEthereumTx` Cosmos messages; the keeper applies them against an SDK-managed statedb.
- `x/feemarket` — EIP-1559-style base fee for EVM txs.
- `x/erc20` — bridges native Cosmos coins to ERC-20 contracts and vice versa (token pair registry).
- `x/inflation`, `x/epochs`, `x/vesting`, `x/ibc` — Evmos-specific token economics, epoch scheduling, vesting accounts, and IBC middleware.

`app/app.go` constructs the keepers in a deliberate order (params → capability → auth → bank → staking → distribution → … → evm → feemarket → erc20 → ibc …). Don't reorder keeper construction without understanding the dependency graph (e.g. evm depends on bank, feemarket; erc20 depends on evm).

### Two-server frontend (Cosmos gRPC + Ethereum JSON-RPC)

A single `lokad start` process runs both:

- The standard Cosmos SDK gRPC / REST / Tendermint RPC stack.
- An Ethereum-compatible JSON-RPC server under `server/` and `rpc/`. `server/json_rpc.go` mounts the namespace handlers from `rpc/namespaces/ethereum/` (eth, txpool, personal, net, debug, web3), backed by `rpc/backend/` which translates `eth_*` calls into reads against the SDK statedb / `x/evm` keeper. `rpc/stream/` powers `eth_subscribe` over websockets.

When debugging "why does `eth_call` return X" or "why doesn't `eth_getLogs` return event Y", trace from `rpc/namespaces/ethereum/eth/api.go` → `rpc/backend/*.go` → `x/evm/keeper/...`. The JSON-RPC layer is a thin translator over the SDK keeper, not a separate execution engine.

### Parallel execution (Block-STM)

The chain uses Block-STM parallel execution, enabled by setting `block-executor = "block-stm"` in `app.toml` (done automatically by `local_node.sh`). Relevant entry points:

- `app/executor.go` — wires the block-STM executor into the SDK baseapp.
- `app/tps_counter.go` — telemetry hook used by `show_tps.sh`.

Block-STM is upstream from `crypto-org-chain/cronos`-style forks; it runs txs optimistically in parallel and re-executes on storage-slot conflicts. Anything that mutates per-tx state outside the SDK store (global vars, singletons, ad-hoc caches) will break determinism — keep state changes inside `sdk.Context`.

### Forked storage stack: memiavl, versiondb, store

Three sibling Go modules ship in-repo and replace pieces of the upstream Cosmos SDK store:

- `memiavl/` — in-memory IAVL with disk snapshots; faster commit, lower I/O than stock IAVL. Enabled via `[memiavl] enable = true` in `app.toml`.
- `versiondb/` — versioned KV store backed by rocksdb (`-tags rocksdb` required).
- `store/` — wrapper / extensions over the SDK store package.

Each has its own `go.mod`. Run `make test-memiavl` / `make test-store` / `make test-versiondb` instead of expecting the root `go test ./...` to cover them. The root module pulls them in via `replace` directives in `go.mod`.

### CLI entry point

`cmd/lokad/main.go` → `cmd/lokad/root.go` builds the cobra command tree. `txCommand` accepts the temp-app's `BasicManager` so that custom tx subcommands resolve correctly (see commit `43b9cdb79`). `cmd/lokad/testnet.go` builds the multi-validator testnet files (`lokad testnet init-files`). The `localnet-*` Makefile targets shell out to this via Docker.

## Conventions worth knowing

- **Don't blanket-rename `evmos` → `loka`** in vendored / forked SDK code; many `*types`, `*keeper`, and proto package paths intentionally stay `evmos` for binary compatibility with the SDK ecosystem. The rebrand is at the application surface (`lokad` binary, `aloka` denom, `loka` bech32) and in user-facing branding (README, docs, scripts). Audit each rename against `git grep` before committing.
- **goreleaser is pinned by digest** (`Makefile` `GORELEASER_CROSS_DIGEST`). If a release build fails after an apparently unrelated dependency bump, suspect a silent goreleaser-cross image update and refresh the digest deliberately rather than re-pulling `:latest`.
- **Genesis edits live in `local_node.sh`**. When adding a new module that needs custom genesis params for local dev, edit `local_node.sh`'s `jq` block rather than scattering setup across scripts.
- **Proto generation runs in Docker** (`make proto-gen`, pinned to `ghcr.io/cosmos/proto-builder:0.14.0`). Don't expect local `protoc` to match — regenerate inside the container.
- **`tests/e2e` is not in the default `make test`** because it builds a docker image and runs upgrade scenarios. CI runs it separately; locally it needs `make build-docker` first (or `TARGET_VERSION=<tag>` to skip the local build).
