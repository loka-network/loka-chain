#!/bin/bash

KEYS[0]="dev0"
KEYS[1]="dev1"
KEYS[2]="dev2"
CHAINID="loka_567002-1"
MONIKER="localtestnet"
# Remember to change to other types of keyring like 'file' in-case exposing to outside world,
# otherwise your balance will be wiped quickly
# The keyring test does not require private key to steal tokens from you
KEYRING="test"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
# Set dedicated home directory for the lokad instance
HOMEDIR="$HOME/.tmp-lokad"
# to trace evm
#TRACE="--trace"
TRACE=""
RUN_IN_BACKGROUND=false

# Path variables
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e


# Reinstall daemon
if [[ "$COSMOS_BUILD_OPTIONS" == *"rocksdb"* ]]; then
	export DYLD_LIBRARY_PATH=/opt/homebrew/lib:/usr/lib:$DYLD_LIBRARY_PATH
	export COSMOS_BUILD_OPTIONS=rocksdb
	make build-rocksdb
else
	make install
fi

# User prompt if an existing local node configuration is found.
if [ -d "$HOMEDIR" ]; then
	printf "\nAn existing folder at '%s' was found. You can choose to delete this folder and start a new local node with new keys from genesis. When declined, the existing local node is started. \n" "$HOMEDIR"
	echo "Overwrite the existing configuration and start a new local node? [y/n]"
	read -r overwrite
else
	overwrite="Y"
fi

# Setup local node if overwrite is set to Yes, otherwise skip setup
if [[ $overwrite == "y" || $overwrite == "Y" ]]; then
	# Remove the previous folder
	rm -rf "$HOMEDIR"

	# Set client config
	lokad config set client chain-id "$CHAINID" --home "$HOMEDIR"
	lokad config set client keyring-backend "$KEYRING" --home "$HOMEDIR"

	# If keys exist they should be deleted
	for KEY in "${KEYS[@]}"; do
		lokad keys add "$KEY" --keyring-backend $KEYRING --algo $KEYALGO --home "$HOMEDIR"
	done

	# Set moniker and chain-id for Loka (Moniker can be anything, chain-id must be an integer)
	lokad init $MONIKER -o --chain-id $CHAINID --home "$HOMEDIR"

	# Change parameter token denominations to aloka
	jq '.app_state["staking"]["params"]["bond_denom"]="aloka"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["crisis"]["constant_fee"]["denom"]="aloka"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="aloka"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["gov"]["params"]["expedited_min_deposit"][0]["denom"]="aloka"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["evm"]["params"]["evm_denom"]="aloka"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["inflation"]["params"]["mint_denom"]="aloka"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set gas limit in genesis
	jq '.consensus["params"]["block"]["max_gas"]="363000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set claims start time
	current_date=$(date -u +"%Y-%m-%dT%TZ")
	jq -r --arg current_date "$current_date" '.app_state["claims"]["params"]["airdrop_start_time"]=$current_date' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set claims records for validator account
	amount_to_claim=10000
	claims_key=${KEYS[0]}
	node_address=$(lokad keys show "$claims_key" --keyring-backend $KEYRING --home "$HOMEDIR" | grep "address" | cut -c12-)
	jq -r --arg node_address "$node_address" --arg amount_to_claim "$amount_to_claim" '.app_state["claims"]["claims_records"]=[{"initial_claimable_amount":$amount_to_claim, "actions_completed":[false, false, false, false],"address":$node_address}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set claims decay
	jq '.app_state["claims"]["params"]["duration_of_decay"]="1000000s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["claims"]["params"]["duration_until_decay"]="100000s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Claim module account:
	# 0xA61808Fe40fEb8B3433778BBC2ecECCAA47c8c47 || loka15cvq3ljql6utxseh0zau9m8ve2j8erz8xrvsjp
	jq -r --arg amount_to_claim "$amount_to_claim" '.app_state["bank"]["balances"] += [{"address":"loka15cvq3ljql6utxseh0zau9m8ve2j8erz8xrvsjp","coins":[{"denom":"aloka", "amount":$amount_to_claim}]}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	if [[ $1 == "pending" ]]; then
		if [[ "$OSTYPE" == "darwin"* ]]; then
			sed -i '' 's/timeout_propose = "3s"/timeout_propose = "30s"/g' "$CONFIG"
			sed -i '' 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' "$CONFIG"
			sed -i '' 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' "$CONFIG"
			sed -i '' 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_commit = "5s"/timeout_commit = "150s"/g' "$CONFIG"
			sed -i '' 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' "$CONFIG"
		else
			sed -i 's/timeout_propose = "3s"/timeout_propose = "30s"/g' "$CONFIG"
			sed -i 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' "$CONFIG"
			sed -i 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' "$CONFIG"
			sed -i 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_commit = "5s"/timeout_commit = "150s"/g' "$CONFIG"
			sed -i 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' "$CONFIG"
		fi
	fi

	# enable prometheus metrics
	if [[ "$OSTYPE" == "darwin"* ]]; then
		# sed -i '' 's/prometheus = false/prometheus = true/' "$CONFIG"
		sed -i '' 's/prometheus-retention-time = 0/prometheus-retention-time  = 1000000000000/g' "$APP_TOML"
		sed -i '' 's/enabled = false/enabled = false/g' "$APP_TOML"
	else
		# sed -i 's/prometheus = false/prometheus = true/' "$CONFIG"
		sed -i 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$APP_TOML"
		sed -i 's/enabled = false/enabled = false/g' "$APP_TOML"
	fi

	# Set consensus timeouts
    sed -i.bak 's/timeout_propose = ".*"/timeout_propose = "200ms"/g' "$CONFIG"
    sed -i.bak 's/timeout_propose_delta = ".*"/timeout_propose_delta = "100ms"/g' "$CONFIG"
    sed -i.bak 's/timeout_prevote = ".*"/timeout_prevote = "200ms"/g' "$CONFIG"
    sed -i.bak 's/timeout_prevote_delta = ".*"/timeout_prevote_delta = "100ms"/g' "$CONFIG"
    sed -i.bak 's/timeout_precommit = ".*"/timeout_precommit = "200ms"/g' "$CONFIG"
    sed -i.bak 's/timeout_precommit_delta = ".*"/timeout_precommit_delta = "100ms"/g' "$CONFIG"
    sed -i.bak 's/timeout_commit = ".*"/timeout_commit = "1s"/g' "$CONFIG"
    sed -i.bak 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' "$CONFIG"

	# Set mempool size
    sed -i.bak 's/^size = .*/size = 200000/g' "$CONFIG"
	if [[ "$COSMOS_BUILD_OPTIONS" == *"rocksdb"* ]]; then
    	sed -i.bak 's/db_backend = "goleveldb"/db_backend = "rocksdb"/g' "$CONFIG"
	fi 

	# Change proposal periods to pass within a reasonable time for local testing
	sed -i.bak 's/"max_deposit_period": "172800s"/"max_deposit_period": "30s"/g' "$HOMEDIR"/config/genesis.json
	sed -i.bak 's/"voting_period": "172800s"/"voting_period": "30s"/g' "$HOMEDIR"/config/genesis.json
	sed -i.bak 's/"expedited_voting_period": "86400s"/"expedited_voting_period": "15s"/g' "$GENESIS"

	# set custom pruning settings, archiving node, pruning = nothing
	sed -i.bak 's/pruning = "default"/pruning = "nothing"/g' "$APP_TOML"
	sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
	sed -i.bak 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"
	sed -i.bak 's/block-executor = ".*"/block-executor = "block-stm"/g' "$APP_TOML"

	# Allocate genesis accounts (cosmos formatted addresses)
	for KEY in "${KEYS[@]}"; do
		lokad add-genesis-account "$KEY" 100000000000000000000000000aloka --keyring-backend $KEYRING --home "$HOMEDIR"
	done

	# bc is required to add these big numbers
	total_supply=$(echo "${#KEYS[@]} * 100000000000000000000000000 + $amount_to_claim" | bc)
	jq -r --arg total_supply "$total_supply" '.app_state["bank"]["supply"][0]["amount"]=$total_supply' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	
	# Sign genesis transaction
	echo "gentx: Sign genesis transaction"
	lokad gentx "${KEYS[0]}" 1000000000000000000000aloka --keyring-backend $KEYRING --chain-id $CHAINID --home "$HOMEDIR"
	## In case you want to create multiple validators at genesis
	## 1. Back to `lokad keys add` step, init more keys
	## 2. Back to `lokad add-genesis-account` step, add balance for those
	## 3. Clone this ~/.lokad home directory into some others, let's say `~/.clonedLokad`
	## 4. Run `gentx` in each of those folders
	## 5. Copy the `gentx-*` folders under `~/.clonedLokad/config/gentx/` folders into the original `~/.lokad/config/gentx`

	# Collect genesis tx
	echo "collect-gentxs: genesis transaction"
	lokad collect-gentxs --home "$HOMEDIR"

	# Run this to ensure everything worked and that the genesis file is setup correctly
	echo "validate-genesis: genesis transaction"
	lokad validate-genesis --home "$HOMEDIR"

	if [[ $1 == "pending" ]]; then
		echo "pending mode is on, please wait for the first block committed."
	fi
fi

# check if the first argument is "nohup" to run in background
if [[ "$1" == "nohup" ]]; then
    RUN_IN_BACKGROUND=true
fi

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
echo "start: localnode"
if [[ "$RUN_IN_BACKGROUND" == true ]]; then
    echo "Starting node in background (logs: nohup.out)..."
    nohup lokad start --log_level $LOGLEVEL --minimum-gas-prices=0.0001aloka --json-rpc.api eth,txpool,personal,net,debug,web3 --json-rpc.address 0.0.0.0:8545 --json-rpc.ws-address 0.0.0.0:8546 --api.enable --home "$HOMEDIR" 2>&1 &
    echo "Node started in background. PID: $!"
else
    echo "Starting node in foreground (logs to console)..."
	lokad start --log_level $LOGLEVEL --minimum-gas-prices=0.0001aloka --json-rpc.api eth,txpool,personal,net,debug,web3 --json-rpc.address 0.0.0.0:8545 --json-rpc.ws-address 0.0.0.0:8546 --api.enable --home "$HOMEDIR"
fi