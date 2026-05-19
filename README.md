<div align="center">

<img src="lokachain-logo.svg" alt="Loka Chain — Parallel EVM · NeoBFT + Block-STM · 300K TPS" width="640" />

# Loka Chain

**Parallel EVM L1 for high-throughput on-chain finance.**
*NeoBFT consensus. Block-STM parallel execution. Full EVM compatibility.*

<a href="https://github.com/loka-network/loka-chain/releases"><img src="https://img.shields.io/github/v/release/loka-network/loka-chain?style=for-the-badge&color=FFD86B&label=release" alt="Release" /></a>
<a href="https://github.com/loka-network/loka-chain/actions"><img src="https://img.shields.io/github/actions/workflow/status/loka-network/loka-chain/goreleaser.yml?style=for-the-badge&label=build" alt="Build" /></a>
<a href="https://github.com/loka-network/loka-chain/stargazers"><img src="https://img.shields.io/github/stars/loka-network/loka-chain?style=for-the-badge&logo=github&color=FFD93D" alt="Stars" /></a>
<a href="https://discord.gg/hetu"><img src="https://img.shields.io/badge/discord-join-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="Discord" /></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=for-the-badge" alt="License" /></a>
<a href="https://golang.org"><img src="https://img.shields.io/badge/go-1.23.3%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" /></a>
<a href="https://soliditylang.org"><img src="https://img.shields.io/badge/EVM-compatible-627EEA?style=for-the-badge&logo=ethereum&logoColor=white" alt="EVM" /></a>
<a href="https://x.com/lokachain"><img src="https://img.shields.io/badge/twitter-@lokachain-1DA1F2?style=for-the-badge&logo=twitter&logoColor=white" alt="Twitter" /></a>
<a href="https://lokachain.org/"><img src="https://img.shields.io/badge/website-lokachain.org-blue?style=for-the-badge" alt="Website" /></a>

[Quick Start](#-quick-start) · [Why Loka Chain](#-why-loka-chain) · [Architecture](#-architecture-three-layer-parachain-model) · [Docs](#-documentation) · [Community](#-community)

</div>

---

## What is Loka Chain

Loka Chain is a **high-performance EVM-compatible Layer 1** built for enterprise-grade on-chain finance. Where general-purpose L1s force every contract through a single sequential execution pipeline, Loka pairs **NeoBFT consensus** (Byzantine-fault-tolerant, deterministic finality) with **Block-STM parallel execution** (optimistic concurrency with fine-grained conflict detection), then composes the network as a **three-layer parachain model** so independent applications run on dedicated lanes instead of stepping on each other.

Existing Solidity contracts, Hardhat / Foundry pipelines, and MetaMask wallets work unchanged.

> **What this means in practice**
>
> - **300,000 TPS** on the sequencer, **100,000 TPS per parachain** with sub-second finality
> - **Block-STM** runs transactions in parallel, detects conflicts at the storage-slot level, and commits deterministically
> - **One viral app does not tank the rest** — parachains have isolated gas, governance, and execution
> - Drop-in **EVM compatibility**: Solidity, JSON-RPC, MetaMask, Hardhat, Foundry — no rewrite, no new tooling

---

## ⚡ Quick Start

### Local node (single-machine)

```sh
git clone https://github.com/loka-network/loka-chain
cd loka-chain
make install          # produces lokad in $GOPATH/bin (Go 1.23.3+, macOS 15+)
./local_node.sh       # spins up a local validator with sensible defaults
```

### Connect with standard Ethereum tooling

```sh
# JSON-RPC endpoint exposed at :8545 by default
export LOKA_RPC=http://127.0.0.1:8545

# MetaMask / Hardhat / Foundry — works unchanged
cast block-number --rpc-url $LOKA_RPC
cast chain-id      --rpc-url $LOKA_RPC
```

### Production-style 4-validator network

```sh
# Run on each validator host with lokad in $PATH
./init_validators.sh <ip1> <ip2> <ip3> <ip4>
./start_node_archive.sh   # archive-mode full node
```

> [!NOTE]
> Loka Chain is in **beta**. NeoBFT consensus and the Block-STM execution engine must be configured per the [`loka-consensus`](https://github.com/loka-network/loka-consensus) deployment guide before mainnet.

---

## 🎯 Why Loka Chain

Four structural advantages over both monolithic L1s and stock Cosmos SDK chains.

### 1. Parallel execution — Block-STM, not a single-threaded VM

Most EVM-compatible chains execute transactions one at a time. Loka uses **Block-STM** (Software Transactional Memory at the block level): every tx in a block runs **optimistically in parallel**, the runtime watches for storage-slot read/write conflicts, and any conflicting tx is re-executed deterministically until commit order matches a serial schedule.

| Chain | Execution model | Realistic TPS |
|---|---|---|
| Ethereum L1 | Sequential EVM | ~15 |
| Polygon PoS | Sequential EVM | ~7,000 (with reorg risk) |
| Stock Cosmos SDK | Sequential | ~2,000 |
| **Loka Chain (sequencer)** | **Block-STM parallel** | **~300,000** |
| **Loka Chain (per parachain)** | **Block-STM parallel** | **~100,000 with sub-second finality** |

Block-STM is the same primitive Aptos uses — except **Loka keeps full EVM bytecode compatibility** so you don't rewrite contracts in Move.

### 2. Drop-in EVM, zero migration

| What works unchanged | How |
|---|---|
| Solidity contracts | Deploy the existing bytecode — no recompile, no auditor re-review |
| MetaMask | Add the JSON-RPC URL, that's it |
| Hardhat / Foundry / Remix | Standard Ethereum JSON-RPC endpoints |
| ERC-20 / ERC-721 / ERC-4337 | Same standards, same audit trail |
| Subgraph / The Graph indexers | Same event log format |

Alternative high-TPS chains (Solana, Aptos, Sui) make you rewrite in Rust / Move / something else. Loka **keeps the Solidity tooling moat intact**.

### 3. Modular parachains — apps don't share gas with each other

Most L1s are **monolithic**: a viral game, NFT mint, or memecoin spam crashes the gas market for every other app on the chain. Loka isolates applications onto **dedicated parachains**:

- **Per-parachain gas market** — payments doesn't pay for a game's congestion
- **Per-parachain governance** — compliance-sensitive deployments can set their own rules
- **Per-parachain VM tuning** — high-frequency trading parachain ≠ social parachain
- Shared **Chain-S settlement** layer for cross-parachain state anchoring (IBC-style)
- Optional **TEE-secured validators** (Intel SGX, AWS Nitro Enclaves) per parachain for verifiable execution

### 4. Enterprise finance stack out of the box

You don't have to assemble a stablecoin + DEX + KYC + AA stack from third parties:

| Component | What it gives you |
|---|---|
| **LOKA Stable** | Compliant stablecoin issuance with real-time on-chain auditability |
| **LOKA DEX** | On-chain order book with **1 ms matching latency** + unified liquidity |
| **Account Abstraction (AA)** | Sponsored transactions, social recovery, programmable accounts |
| **Gas Abstraction** | Pay gas in any whitelisted asset (USDC, your stablecoin, etc.) |
| **KYC whitelisting + audit logs** | Built into the chain runtime, not bolted on |

---

## 🌐 Supported VMs & Networks

| Layer | Component | Status | When to use it |
|------|---------|--------|----------------|
| **Execution** | EVM (Solidity / Vyper) | ✅ Live | Existing Ethereum dApps, DeFi, NFT, RWA |
| **Execution** | WASM | 🚧 Optional | Compute-heavy / non-Solidity workloads |
| **Execution** | TEE (SGX / Nitro) | 🚧 Optional | Verifiable off-chain computation, MEV-resistant ordering |
| **Settlement** | Chain-S main chain | ✅ Live | State anchoring, cross-parachain transfers |
| **Parachains** | Per-application L2 | ✅ Live | Payments · trading · gaming · compliance |
| **Bridge** | IBC-style cross-parachain | ✅ Live | Native interop without trusted relayers |
| **Bridge** | Ethereum L1 bridge | 🔜 Upcoming | USDC / stETH / blue-chip asset on-ramp |

---

## 🏗 Architecture: Three-Layer Parachain Model

Loka separates **what** is being computed from **where** and **how** it's settled. Apps live in parachains; state anchors back to a single main settlement chain; consensus and parallel execution are a shared substrate underneath everything.

```text
┌─────────────────────────────────────────────────────────────────────────────────┐
│                       Application Layer                                         │
│       Solidity contracts · ERC-20/721/4337 · LOKA Stable · LOKA DEX             │
└────┬──────────────────┬──────────────────┬──────────────────┬───────────────────┘
     │ parachain=pay    │ parachain=trade  │ parachain=game   │ parachain=...
┌────▼────────────┐ ┌───▼──────────────┐ ┌─▼────────────────┐ ┌─▼──────────────┐
│ Parachain L2    │ │ Parachain L2     │ │ Parachain L2     │ │ Parachain L2   │
│ • own gas mkt   │ │ • own gas mkt    │ │ • own gas mkt    │ │ • own gas mkt  │
│ • own gov       │ │ • own gov        │ │ • own gov        │ │ • own gov      │
│ • EVM / WASM    │ │ • on-chain CLOB  │ │ • EVM            │ │ • EVM / TEE    │
│ • 100K TPS      │ │ • 1ms matching   │ │ • 100K TPS       │ │ • 100K TPS     │
└────────┬────────┘ └─────────┬────────┘ └────────┬─────────┘ └───────┬────────┘
         │                    │                   │                   │
         └────────────────────┴───────────────────┴───────────────────┘
                                       │
┌──────────────────────────────────────▼──────────────────────────────────────────┐
│                Chain-S — Main Settlement Layer                                  │
│       State anchoring · IBC-style cross-parachain · Optional core settlement    │
└──────────────────────────────────────┬──────────────────────────────────────────┘
                                       │
┌──────────────────────────────────────▼──────────────────────────────────────────┐
│           Consensus & Execution Substrate (shared by all parachains)            │
│       NeoBFT (deterministic BFT finality)  ·  Block-STM (parallel EVM)          │
│                       Optional TEE validators (SGX / Nitro)                     │
└─────────────────────────────────────────────────────────────────────────────────┘
```

The substrate is uniform; the application layer is plural. Every parachain inherits NeoBFT's finality guarantees and Block-STM's throughput, but each one decides its own gas model, governance, and (optionally) VM. Adding a parachain does **not** cost throughput on the others — it adds throughput overall.

---

## 🔬 Technical Implementation

<details>
<summary><b>1. NeoBFT consensus</b></summary>

Byzantine-fault-tolerant consensus with **deterministic finality** (1 block = final, no probabilistic settlement). Sub-second block times. Tolerates up to ⌊(n−1)/3⌋ Byzantine validators. Pluggable proposer selection (round-robin / weighted) per parachain.

</details>

<details>
<summary><b>2. Block-STM parallel execution</b></summary>

Inspired by Aptos's Block-STM paper. Every tx in a block is scheduled to run in parallel; the runtime tracks the read/write set per tx at storage-slot granularity. On conflict, the later tx is **re-executed**, not aborted — eventually the commit order converges with a serial schedule, so commits are deterministic regardless of how many cores executed in parallel.

Combined with EVM bytecode means: existing Solidity contracts get parallel execution for free; no programming-model changes for developers.

</details>

<details>
<summary><b>3. Full EVM compatibility</b></summary>

Standard `eth_*` and `net_*` JSON-RPC endpoints. EIP-155 chain IDs. Eth_subscribe / WebSocket. ERC-20/721/1155/4337 native. The Graph subgraphs work unchanged.

</details>

<details>
<summary><b>4. Parachain isolation</b></summary>

Each parachain is a sovereign execution environment with its own state tree, gas market, and governance, anchored back to Chain-S. Cross-parachain transfers go through an IBC-style atomic packet-relay — no trusted bridge custodian.

</details>

<details>
<summary><b>5. Optional TEE-secured validators</b></summary>

Validators can opt into running inside Intel SGX or AWS Nitro Enclaves. The enclave attestation lets users verify, at quorum time, that a specific code commit (not an arbitrary modified binary) produced the block. Useful for regulated stablecoin issuance, verifiable order matching, and MEV-resistance.

</details>

---

## 📋 Core Features

- **Consensus:** NeoBFT, BFT finality, sub-second blocks
- **Execution:** Block-STM parallel EVM, 300K TPS sequencer, 100K TPS per parachain
- **VM:** EVM (default) · WASM (optional) · TEE-enclave (optional)
- **Compatibility:** Solidity / JSON-RPC / MetaMask / Hardhat / Foundry / The Graph
- **Networking:** IBC-style cross-parachain, native atomic packet relay
- **Finance primitives:** LOKA Stable, LOKA DEX (1 ms matching), Account Abstraction, Gas Abstraction
- **Compliance:** KYC whitelisting, audit logs, opt-in TEE attestation
- **Operator UX:** `local_node.sh` for dev, `init_validators.sh` for production

---

## 🗺 Roadmap

- ✅ NeoBFT consensus on Chain-S main settlement
- ✅ Block-STM parallel EVM execution
- ✅ Parachain isolation with per-chain gas / governance
- ✅ EVM JSON-RPC + ERC-20/721/4337 support
- ✅ LOKA Stable issuance module
- 🚧 LOKA DEX 1 ms on-chain CLOB
- 🚧 Account Abstraction (EIP-4337) end-to-end UX
- 🚧 TEE-secured validator opt-in (Intel SGX + AWS Nitro)
- 🔜 Ethereum L1 bridge for USDC / stETH on-ramp
- 🔜 WASM execution lane for non-Solidity workloads
- 🔜 Mainnet launch

---

## 🛠 Build & Test

```sh
make install          # build and install lokad to $GOPATH/bin
make build            # produce build/lokad locally
make test             # unit tests
make test-race        # race detector
make lint             # golangci-lint
make release-dry-run  # verify the goreleaser pipeline locally
```

> Required Go version: **1.23.3+** (macOS 15 or later for the Apple Silicon build path)

### Docker

```sh
docker-compose up     # spin up a local network with the bundled compose file
```

---

## 📚 Documentation

| Topic | File |
|------|------|
| Parallel parachain architecture | [docs/parallel-sub-chain.md](docs/parallel-sub-chain.md) |
| Parallel-chain economic model | [docs/parallel-chain-economic.md](docs/parallel-chain-economic.md) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |
| Docker compose | [docker-compose.yml](docker-compose.yml) |
| Release pipeline | [.github/workflows/goreleaser.yml](.github/workflows/goreleaser.yml) |
| Validator init script | [init_validators.sh](init_validators.sh) |
| Local dev script | [local_node.sh](local_node.sh) |

---

## 🔒 Security

Loka Chain is in **beta**. Do not custody mainnet-equivalent value on it yet.

> [!IMPORTANT]
> If you discover a security vulnerability, please open an issue at <https://github.com/loka-network/loka-chain/issues> tagged `security`, or reach out privately first for sensitive findings.

---

## 🤝 Community

- **Open an Issue:** <https://github.com/loka-network/loka-chain/issues>
- **Website:** <https://www.lokachain.org/>
- **Twitter / X:** <https://x.com/lokachain>
- **Discord:** <https://discord.gg/hetu>

We welcome all contributions:

- Open a PR — bug fixes, new features, performance work
- File issues for bugs / feature requests
- Improve documentation
- Share use cases that exercise the parallel-execution model

---

## 📄 License

See [LICENSE](LICENSE).

<div align="center">
<sub>built by <a href="https://lokachain.org">Loka Network</a> · part of the <a href="https://hetu.io">Hetu Project</a></sub>
</div>
