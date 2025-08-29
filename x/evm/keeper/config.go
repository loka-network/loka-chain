// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package keeper

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	evmostypes "github.com/loka-network/loka/v1/types"
	"github.com/loka-network/loka/v1/x/evm/statedb"
	"github.com/loka-network/loka/v1/x/evm/types"
	feemarkettypes "github.com/loka-network/loka/v1/x/feemarket/types"
)

// EVMBlockConfig encapsulates the common parameters needed to execute an EVM message,
// it's cached in object store during the block execution.
type EVMBlockConfig struct {
	Params          types.Params
	FeeMarketParams feemarkettypes.Params
	ChainConfig     *params.ChainConfig
	CoinBase        common.Address
	BaseFee         *big.Int
	// not supported, always zero
	Random *common.Hash
	// unused, always zero
	Difficulty *big.Int
	// cache the big.Int version of block number, avoid repeated allocation
	BlockNumber *big.Int
	BlockTime   uint64
	Rules       params.Rules
}

// EVMConfig encapsulates common parameters needed to create an EVM to execute a message
// It's mainly to reduce the number of method parameters
type EVMConfig struct {
	*EVMBlockConfig
	TxConfig   statedb.TxConfig
	Tracer     vm.EVMLogger
	DebugTrace bool
	// Overrides      *rpctypes.StateOverride
	// BlockOverrides *rpctypes.BlockOverrides
}

// EVMBlockConfig creates the EVMBlockConfig based on current state
func (k *Keeper) EVMBlockConfig(ctx sdk.Context, chainID *big.Int) (*EVMBlockConfig, error) {
	objStore := ctx.ObjectStore(k.objectKey)
	v := objStore.Get(types.KeyPrefixObjectParams)
	if v != nil {
		return v.(*EVMBlockConfig), nil
	}

	params := k.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(chainID)

	feemarketParams := k.feeMarketKeeper.GetParams(ctx)

	// get the coinbase address from the block proposer
	proposerAddress := sdk.ConsAddress(ctx.BlockHeader().ProposerAddress)
	if len(proposerAddress) == 0 {
		// it's ok that proposer address don't exsits in some contexts like CheckTx.
		proposerAddress = sdk.ConsAddress(nil)
	}
	coinbase, err := k.GetCoinbaseAddress(ctx, proposerAddress)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to obtain coinbase address")
	}

	var baseFee *big.Int
	if types.IsLondon(ethCfg, ctx.BlockHeight()) {
		baseFee = feemarketParams.BaseFee.BigInt()
		// should not be nil if london hardfork enabled
		if baseFee == nil {
			baseFee = new(big.Int)
		}
	}
	time := ctx.BlockHeader().Time
	var blockTime uint64
	if !time.IsZero() {
		blockTime, err = evmostypes.SafeUint64(time.Unix())
		if err != nil {
			return nil, err
		}
	}
	blockNumber := big.NewInt(ctx.BlockHeight())
	rules := ethCfg.Rules(blockNumber, ethCfg.MergeNetsplitBlock != nil)

	var zero common.Hash
	cfg := &EVMBlockConfig{
		Params:          params,
		FeeMarketParams: feemarketParams,
		ChainConfig:     ethCfg,
		CoinBase:        coinbase,
		BaseFee:         baseFee,
		Difficulty:      big.NewInt(0),
		Random:          &zero,
		BlockNumber:     blockNumber,
		BlockTime:       blockTime,
		Rules:           rules,
	}
	objStore.Set(types.KeyPrefixObjectParams, cfg)
	return cfg, nil
}

func (k *Keeper) RemoveParamsCache(ctx sdk.Context) {
	ctx.ObjectStore(k.objectKey).Delete(types.KeyPrefixObjectParams)
}

func (cfg EVMConfig) GetTracer() vm.EVMLogger {
	if _, ok := cfg.Tracer.(*types.NoOpTracer); ok {
		return nil
	}
	return cfg.Tracer
}

// EVMConfig creates the EVMConfig based on current state
func (k *Keeper) EVMConfig(ctx sdk.Context, chainID *big.Int, txHash common.Hash) (*EVMConfig, error) {
	blockCfg, err := k.EVMBlockConfig(ctx, chainID)
	if err != nil {
		return nil, err
	}
	var txConfig statedb.TxConfig
	if txHash == (common.Hash{}) {
		txConfig = statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash()))
	} else {
		txConfig = k.TxConfig(ctx, txHash)
	}

	return &EVMConfig{
		EVMBlockConfig: blockCfg,
		TxConfig:       txConfig,
	}, nil
}

// TxConfig loads `TxConfig` from current transient storage
func (k *Keeper) TxConfig(ctx sdk.Context, txHash common.Hash) statedb.TxConfig {
	return statedb.NewTxConfig(
		common.BytesToHash(ctx.HeaderHash()), // BlockHash
		txHash,                               // TxHash
		0,                                    // TxIndex
		0,                                    // LogIndex
	)
}

// VMConfig creates an EVM configuration from the debug setting and the extra EIPs enabled on the
// module parameters. The config generated uses the default JumpTable from the EVM.
func (k Keeper) VMConfig(ctx sdk.Context, cfg *EVMConfig) vm.Config {
	noBaseFee := true
	if types.IsLondon(cfg.ChainConfig, ctx.BlockHeight()) {
		noBaseFee = cfg.FeeMarketParams.NoBaseFee
	}

	return vm.Config{
		Tracer:    cfg.GetTracer(),
		NoBaseFee: noBaseFee,
		ExtraEips: cfg.Params.EIPs(),
	}
}
