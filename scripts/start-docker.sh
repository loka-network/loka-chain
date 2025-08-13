#!/bin/bash

KEY="dev0"
CHAINID="loka_567000-1"
MONIKER="mymoniker"
DATA_DIR=$(mktemp -d -t loka-datadir.XXXXX)

echo "create and add new keys"
./lokad keys add $KEY --home $DATA_DIR --no-backup --chain-id $CHAINID --algo "eth_secp256k1" --keyring-backend test
echo "init Loka with moniker=$MONIKER and chain-id=$CHAINID"
./lokad init $MONIKER --chain-id $CHAINID --home $DATA_DIR
echo "prepare genesis: Allocate genesis accounts"
./lokad add-genesis-account \
    "$(./lokad keys show $KEY -a --home $DATA_DIR --keyring-backend test)" 1000000000000000000aloka,1000000000000000000stake \
    --home $DATA_DIR --keyring-backend test
echo "prepare genesis: Sign genesis transaction"
./lokad gentx $KEY 1000000000000000000stake --keyring-backend test --home $DATA_DIR --keyring-backend test --chain-id $CHAINID
echo "prepare genesis: Collect genesis tx"
./lokad collect-gentxs --home $DATA_DIR
echo "prepare genesis: Run validate-genesis to ensure everything worked and that the genesis file is setup correctly"
./lokad validate-genesis --home $DATA_DIR

echo "starting loka node $i in background ..."
./lokad start --pruning=nothing --rpc.unsafe \
    --keyring-backend test --home $DATA_DIR \
    >$DATA_DIR/node.log 2>&1 &
disown

echo "started loka node"
tail -f /dev/null
