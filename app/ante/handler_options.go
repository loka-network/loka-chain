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

package ante

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"github.com/loka-network/loka/v1/app/ante/cache"
	"github.com/loka-network/loka/v1/app/ante/cosmos"
	"github.com/loka-network/loka/v1/app/ante/interfaces"
	anteutils "github.com/loka-network/loka/v1/app/ante/utils"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	cosmosante "github.com/loka-network/loka/v1/app/ante/cosmos"
	evmante "github.com/loka-network/loka/v1/app/ante/evm"
	evmtypes "github.com/loka-network/loka/v1/x/evm/types"

	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	vestingtypes "github.com/loka-network/loka/v1/x/vesting/types"
)

const EthSigVerificationResultCacheKey = "ante:EthSigVerificationResult"

// HandlerOptions defines the list of module keepers required to run the Evmos
// AnteHandler decorators.
type HandlerOptions struct {
	Cdc                    codec.BinaryCodec
	AccountKeeper          evmtypes.AccountKeeper
	BankKeeper             evmtypes.BankKeeper
	DistributionKeeper     anteutils.DistributionKeeper
	IBCKeeper              *ibckeeper.Keeper
	StakingKeeper          vestingtypes.StakingKeeper
	FeeMarketKeeper        evmante.FeeMarketKeeper
	EvmKeeper              interfaces.EVMKeeper
	FeegrantKeeper         ante.FeegrantKeeper
	ExtensionOptionChecker ante.ExtensionOptionChecker
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	MaxTxGasWanted         uint64
	TxFeeChecker           anteutils.TxFeeChecker

	ExtraDecorators   []sdk.AnteDecorator
	PendingTxListener PendingTxListener

	// see #494, just for benchmark, don't turn on on production
	UnsafeUnorderedTx bool
	AnteCache         *cache.AnteCache
}

// Validate checks if the keepers are defined
func (options HandlerOptions) Validate() error {
	if options.Cdc == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "codec is required for AnteHandler")
	}
	if options.AccountKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.IBCKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "ibc keeper is required for AnteHandler")
	}
	if options.StakingKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "staking keeper is required for AnteHandler")
	}
	if options.FeeMarketKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "fee market keeper is required for AnteHandler")
	}
	if options.EvmKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "evm keeper is required for AnteHandler")
	}
	if options.SigGasConsumer == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "signature gas consumer is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "sign mode handler is required for AnteHandler")
	}
	if options.DistributionKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "distribution keeper is required for AnteHandler")
	}
	if options.TxFeeChecker == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "tx fee checker is required for AnteHandler")
	}
	if options.AnteCache == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "ante cache is required for AnteHandler")
	}
	return nil
}

func newEthAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		blockCfg, err := options.EvmKeeper.EVMBlockConfig(ctx, options.EvmKeeper.ChainID())
		if err != nil {
			return ctx, errorsmod.Wrap(errortypes.ErrLogic, err.Error())
		}
		evmParams := &blockCfg.Params
		evmDenom := evmParams.EvmDenom
		feemarketParams := &blockCfg.FeeMarketParams
		baseFee := blockCfg.BaseFee
		rules := blockCfg.Rules

		// all transactions must implement FeeTx
		_, ok := tx.(sdk.FeeTx)
		if !ok {
			ctx.Logger().Error("tx must implement sdk.FeeTx", "tx", tx)
			return ctx, errorsmod.Wrapf(errortypes.ErrInvalidType, "invalid transaction type %T, expected sdk.FeeTx", tx)
		}

		// We need to setup an empty gas config so that the gas is consistent with Ethereum.
		ctx, err = interfaces.SetupEthContext(ctx)
		if err != nil {
			ctx.Logger().Error("failed to setup eth context", "error", err)
			return ctx, err
		}

		if err := cosmos.CheckEthMempoolFee(ctx, tx, simulate, baseFee, evmDenom); err != nil {
			ctx.Logger().Error("CheckEthMempoolFee error", "err", err)
			return ctx, err
		}

		if err := cosmos.CheckEthMinGasPrice(tx, feemarketParams.MinGasPrice, baseFee); err != nil {
			ctx.Logger().Error("CheckEthMinGasPrice error", "err", err)
			return ctx, err
		}

		if err := interfaces.ValidateEthBasic(ctx, tx, evmParams, baseFee); err != nil {
			ctx.Logger().Error("ValidateEthBasic error", "err", err)
			return ctx, err
		}

		if v, ok := ctx.GetIncarnationCache(EthSigVerificationResultCacheKey); ok {
			if v != nil {
				err = v.(error)
			}
		} else {
			ethSigner := ethtypes.MakeSigner(blockCfg.ChainConfig, blockCfg.BlockNumber)
			err = VerifyEthSig(tx, ethSigner)
			ctx.SetIncarnationCache(EthSigVerificationResultCacheKey, err)
		}
		if err != nil {
			ctx.Logger().Error("VerifyEthSig error", "err", err)
			return ctx, err
		}

		// AccountGetter cache the account objects during the ante handler execution,
		// it's safe because there's no store branching in the ante handlers.
		accountGetter := NewCachedAccountGetter(ctx, options.AccountKeeper)

		if err := VerifyEthAccount(ctx, tx, options.EvmKeeper, evmDenom, accountGetter); err != nil {
			ctx.Logger().Error("VerifyEthAccount error", "err", err)
			return ctx, err
		}

		if err := CheckEthCanTransfer(ctx, tx, baseFee, rules, options.EvmKeeper, evmParams); err != nil {
			ctx.Logger().Error("CheckEthCanTransfer error", "err", err)
			return ctx, err
		}

		ctx, err = CheckEthGasConsume(
			ctx, tx, rules, options.EvmKeeper,
			baseFee, options.MaxTxGasWanted, evmDenom,
		)
		if err != nil {
			ctx.Logger().Error("CheckEthGasConsume error", "err", err)
			return ctx, err
		}

		if err := CheckAndSetEthSenderNonce(
			ctx, tx, options.AccountKeeper, options.UnsafeUnorderedTx, accountGetter, options.AnteCache); err != nil {
			ctx.Logger().Error("CheckAndSetEthSenderNonce error", "err", err)
			return ctx, err
		}

		extraDecorators := options.ExtraDecorators
		if options.PendingTxListener != nil {
			extraDecorators = append(extraDecorators, newTxListenerDecorator(options.PendingTxListener))
		}
		if len(extraDecorators) > 0 {
			return sdk.ChainAnteDecorators(extraDecorators...)(ctx, tx, simulate)
		}
		return ctx, nil
	}
}

// Deprecated newEVMAnteHandler creates the default ante handler for Ethereum transactions
// func newEVMAnteHandler(options HandlerOptions) sdk.AnteHandler {
// 	return sdk.ChainAnteDecorators(
// 		// outermost AnteDecorator. SetUpContext must be called first
// 		evmante.NewEthSetUpContextDecorator(options.EvmKeeper),
// 		// Check eth effective gas price against the node's minimal-gas-prices config
// 		evmante.NewEthMempoolFeeDecorator(options.EvmKeeper),
// 		// Check eth effective gas price against the global MinGasPrice
// 		evmante.NewEthMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper),
// 		evmante.NewEthValidateBasicDecorator(options.EvmKeeper),
// 		evmante.NewEthSigVerificationDecorator(options.EvmKeeper),
// 		evmante.NewEthAccountVerificationDecorator(options.AccountKeeper, options.EvmKeeper),
// 		evmante.NewCanTransferDecorator(options.EvmKeeper),
// 		evmante.NewEthVestingTransactionDecorator(options.AccountKeeper, options.BankKeeper, options.EvmKeeper),
// 		evmante.NewEthGasConsumeDecorator(options.BankKeeper, options.DistributionKeeper, options.EvmKeeper, options.StakingKeeper, options.MaxTxGasWanted),
// 		evmante.NewEthIncrementSenderSequenceDecorator(options.AccountKeeper),
// 		// move to checks total gas-wanted against block gas limit in process proposal
// 		// evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
// 		// emit eth tx hash and index at the very last ante handler. (repeated in the PatchTxResponses)
// 		// evmante.NewEthEmitEventDecorator(options.EvmKeeper),
// 	)
// }

// newCosmosAnteHandler creates the default ante handler for Cosmos transactions
func newCosmosAnteHandler(ctx sdk.Context, options HandlerOptions) sdk.AnteHandler {
	evmParams := options.EvmKeeper.GetParams(ctx)
	feemarketParams := options.FeeMarketKeeper.GetParams(ctx)
	evmDenom := evmParams.EvmDenom
	return sdk.ChainAnteDecorators(
		cosmosante.RejectMessagesDecorator{}, // reject MsgEthereumTxs
		cosmosante.NewAuthzLimiterDecorator( // disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
		),
		ante.NewSetUpContextDecorator(),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		cosmosante.NewMinGasPriceDecorator(options.FeeMarketKeeper, evmDenom, &feemarketParams),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		cosmosante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.DistributionKeeper, options.FeegrantKeeper, options.StakingKeeper, options.TxFeeChecker),
		cosmosante.NewVestingDelegationDecorator(options.AccountKeeper, options.StakingKeeper, options.Cdc),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		// evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)
}

// newCosmosAnteHandlerEip712 creates the ante handler for transactions signed with EIP712
func newLegacyCosmosAnteHandlerEip712(ctx sdk.Context, options HandlerOptions) sdk.AnteHandler {
	evmParams := options.EvmKeeper.GetParams(ctx)
	feemarketParams := options.FeeMarketKeeper.GetParams(ctx)
	evmDenom := evmParams.EvmDenom

	return sdk.ChainAnteDecorators(
		cosmosante.RejectMessagesDecorator{}, // reject MsgEthereumTxs
		cosmosante.NewAuthzLimiterDecorator( // disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
		),
		ante.NewSetUpContextDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		cosmosante.NewMinGasPriceDecorator(options.FeeMarketKeeper, evmDenom, &feemarketParams),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		cosmosante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.DistributionKeeper, options.FeegrantKeeper, options.StakingKeeper, options.TxFeeChecker),
		cosmosante.NewVestingDelegationDecorator(options.AccountKeeper, options.StakingKeeper, options.Cdc),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		// Note: signature verification uses EIP instead of the cosmos signature validator
		//nolint: staticcheck
		cosmosante.NewLegacyEip712SigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		// evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)
}
