// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package ante

import (
	"fmt"
	"math"
	"math/big"

	"github.com/loka-network/loka/v1/app/ante/cache"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/loka-network/loka/v1/app/ante/interfaces"
	ethermint "github.com/loka-network/loka/v1/types"
	"github.com/loka-network/loka/v1/x/evm/keeper"
	"github.com/loka-network/loka/v1/x/evm/statedb"
	evmtypes "github.com/loka-network/loka/v1/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type AccountGetter func(sdk.AccAddress) sdk.AccountI

// NewCachedAccountGetter cache the account objects during the ante handler execution,
// it's safe because there's no store branching in the ante handlers,
// it also creates new account in memory if it doesn't exist in the store.
func NewCachedAccountGetter(ctx sdk.Context, ak evmtypes.AccountKeeper) AccountGetter {
	accounts := make(map[string]sdk.AccountI, 1)
	return func(addr sdk.AccAddress) sdk.AccountI {
		acc := accounts[string(addr)]
		if acc == nil {
			acc = ak.GetAccount(ctx, addr)
			if acc == nil {
				// we create a new account in memory if it doesn't exist,
				// which is only set to store when updated.
				acc = ak.NewAccountWithAddress(ctx, addr)
			}
			accounts[string(addr)] = acc
		}
		return acc
	}
}

// VerifyEthAccount validates checks that the sender balance is greater than the total transaction cost.
// The account will be created in memory if it doesn't exist, i.e cannot be found on store, which will eventually set to
// store when increasing nonce.
// This AnteHandler decorator will fail if:
// - any of the msgs is not a MsgEthereumTx
// - from address is empty
// - account balance is lower than the transaction cost
func VerifyEthAccount(
	ctx sdk.Context, tx sdk.Tx,
	evmKeeper interfaces.EVMKeeper, evmDenom string,
	accountGetter AccountGetter,
) error {
	if !ctx.IsCheckTx() {
		return nil
	}

	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
		}

		ethTx := msgEthTx.AsTransaction()

		// sender address should be in the tx cache from the previous AnteHandle call
		from := msgEthTx.GetFrom()
		if from.Empty() {
			return errorsmod.Wrap(errortypes.ErrInvalidAddress, "from address cannot be empty")
		}

		// check whether the sender address is EOA
		acct := statedb.NewAccountFromSdkAccount(accountGetter(from))
		if acct.IsContract() {
			fromAddr := common.BytesToAddress(from)
			return errorsmod.Wrapf(errortypes.ErrInvalidType,
				"the sender is not EOA: address %s, codeHash <%s>", fromAddr, acct.CodeHash)
		}

		balance := evmKeeper.GetBalance(ctx, from, evmDenom)
		if err := keeper.CheckSenderBalanceFromTx(sdkmath.NewIntFromBigIntMut(balance), ethTx); err != nil {
			return errorsmod.Wrap(err, "failed to check sender balance")
		}
	}
	return nil
}

// CheckEthGasConsume validates that the Ethereum tx message has enough to cover intrinsic gas
// (during CheckTx only) and that the sender has enough balance to pay for the gas cost.
//
// Intrinsic gas for a transaction is the amount of gas that the transaction uses before the
// transaction is executed. The gas is a constant value plus any cost incurred by additional bytes
// of data supplied with the transaction.
//
// This AnteHandler decorator will fail if:
// - the message is not a MsgEthereumTx
// - sender account cannot be found
// - transaction's gas limit is lower than the intrinsic gas
// - user doesn't have enough balance to deduct the transaction fees (gas_limit * gas_price)
// - transaction or block gas meter runs out of gas
// - sets the gas meter limit
// - gas limit is greater than the block gas meter limit
func CheckEthGasConsume(
	ctx sdk.Context, tx sdk.Tx,
	rules params.Rules,
	evmKeeper interfaces.EVMKeeper,
	baseFee *big.Int,
	maxGasWanted uint64,
	evmDenom string,
) (sdk.Context, error) {
	gasWanted := uint64(0)
	var events sdk.Events

	// Use the lowest priority of all the messages as the final one.
	minPriority := int64(math.MaxInt64)

	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
		}

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, errorsmod.Wrap(err, "failed to unpack tx data")
		}

		priority := evmtypes.GetTxPriority(txData, baseFee)

		if priority < minPriority {
			minPriority = priority
		}

		// We can't trust the tx gas limit, because we'll refund the unused gas.
		gasLimit := msgEthTx.GetGas()
		if ctx.IsCheckTx() && maxGasWanted != 0 {
			gasLimit = min(gasLimit, maxGasWanted)
		}
		if gasWanted > math.MaxInt64-gasLimit {
			return ctx, fmt.Errorf("gasWanted(%d) + gasLimit(%d) overflow", gasWanted, gasLimit)
		}
		gasWanted += gasLimit
		// user balance is already checked during CheckTx so there's no need to
		// verify it again during ReCheckTx
		if ctx.IsReCheckTx() {
			continue
		}

		fees, err := keeper.VerifyFee(txData, evmDenom, baseFee, rules.IsHomestead, rules.IsIstanbul, rules.IsShanghai)
		if err != nil {
			return ctx, errorsmod.Wrapf(err, "failed to verify the fees")
		}

		fromBytes := common.FromHex(msgEthTx.From)
		err = evmKeeper.DeductTxCostsFromUserBalance(ctx, fees, common.BytesToAddress(fromBytes))
		if err != nil {
			return ctx, errorsmod.Wrapf(err, "failed to deduct transaction costs from user balance")
		}

		events = append(events,
			sdk.NewEvent(
				sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeyFee, fees.String()),
			),
		)
	}

	ctx.EventManager().EmitEvents(events)

	blockGasLimit := ethermint.BlockGasLimit(ctx)

	// return error if the tx gas is greater than the block limit (max gas)

	// NOTE: it's important here to use the gas wanted instead of the gas consumed
	// from the tx gas pool. The later only has the value so far since the
	// EthSetupContextDecorator so it will never exceed the block gas limit.
	if gasWanted > blockGasLimit {
		return ctx, errorsmod.Wrapf(
			errortypes.ErrOutOfGas,
			"tx gas (%d) exceeds block gas limit (%d)",
			gasWanted,
			blockGasLimit,
		)
	}

	// Set tx GasMeter with a limit of GasWanted (i.e gas limit from the Ethereum tx).
	// The gas consumed will be then reset to the gas used by the state transition
	// in the EVM.

	// FIXME: use a custom gas configuration that doesn't add any additional gas and only
	// takes into account the gas consumed at the end of the EVM transaction.
	newCtx := ctx.
		WithGasMeter(ethermint.NewInfiniteGasMeterWithLimit(gasWanted)).
		WithPriority(minPriority)

	// we know that we have enough gas on the pool to cover the intrinsic gas
	return newCtx, nil
}

// CheckEthCanTransfer creates an EVM from the message and calls the BlockContext CanTransfer function to
// see if the address can execute the transaction.
func CheckEthCanTransfer(
	ctx sdk.Context, tx sdk.Tx,
	baseFee *big.Int,
	rules params.Rules,
	evmKeeper interfaces.EVMKeeper,
	evmParams *evmtypes.Params,
) error {
	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
		}

		tx := msgEthTx.AsTransaction()
		if rules.IsLondon {
			if baseFee == nil {
				return errorsmod.Wrap(
					evmtypes.ErrInvalidBaseFee,
					"base fee is supported but evm block context value is nil",
				)
			}
			if tx.GasFeeCap().Cmp(baseFee) < 0 {
				return errorsmod.Wrapf(
					errortypes.ErrInsufficientFee,
					"max fee per gas less than block base fee (%s < %s)",
					tx.GasFeeCap(), baseFee,
				)
			}
		}
		value := tx.Value()
		if value == nil || value.Sign() == -1 {
			return fmt.Errorf("value (%s) must be positive", value)
		}
		fromBytes := common.FromHex(msgEthTx.From)
		from := common.BytesToAddress(fromBytes)
		// check that caller has enough balance to cover asset transfer for **topmost** call
		// NOTE: here the gas consumed is from the context with the infinite gas meter
		if value.Sign() > 0 && !canTransfer(ctx, evmKeeper, evmParams.EvmDenom, from, value) {
			return errorsmod.Wrapf(
				errortypes.ErrInsufficientFunds,
				"failed to transfer %s from address %s using the EVM block context transfer function",
				value,
				from,
			)
		}
	}

	return nil
}

// canTransfer adapted the core.CanTransfer from go-ethereum
func canTransfer(ctx sdk.Context, evmKeeper interfaces.EVMKeeper, denom string, from common.Address, amount *big.Int) bool {
	balance := evmKeeper.GetBalance(ctx, sdk.AccAddress(from.Bytes()), denom)
	return balance.Cmp(amount) >= 0
}

// CheckAndSetEthSenderNonce handles incrementing the sequence of the signer (i.e sender). If the transaction is a
// contract creation, the nonce will be incremented during the transaction execution and not within
// this AnteHandler decorator.
func CheckAndSetEthSenderNonce(
	ctx sdk.Context, tx sdk.Tx, ak evmtypes.AccountKeeper, unsafeUnOrderedTx bool, accountGetter AccountGetter, cache *cache.AnteCache,
) error {
	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
		}

		tx := msgEthTx.AsTransaction()

		from := msgEthTx.GetFrom()
		acc := accountGetter(from)
		if acc == nil {
			return errorsmod.Wrapf(
				errortypes.ErrUnknownAddress,
				"account %s is nil", common.BytesToAddress(from.Bytes()),
			)
		}
		expectedNonce := acc.GetSequence()
		txNonce := tx.Nonce()
		fromStr := from.String()

		// if flag is set, we bypass nonce all check verification
		if !unsafeUnOrderedTx {
			ex := cache.Exists(fromStr, txNonce)
			// to support tx replacement, we check if the transaction nonce exists in the cache and if yes we skip
			// nonce verification, and we don't set the sequence
			// We allow skip verification only during CheckTx to keep sequence safe during the execution.
			if ctx.IsCheckTx() && !ctx.IsReCheckTx() && ex {
				continue
			}

			// nonce verification, the sequence needs to be in order
			if txNonce != expectedNonce {
				return errorsmod.Wrapf(
					errortypes.ErrInvalidSequence,
					"invalid nonce; got %d, expected %d", txNonce, expectedNonce,
				)
			}

			if ctx.IsCheckTx() {
				cache.Set(fromStr, txNonce)
			} else if ex {
				// delete in case of deliver tx
				cache.Delete(fromStr, txNonce)
			}
		}

		// increase sequence of sender
		if err := acc.SetSequence(expectedNonce + 1); err != nil {
			return errorsmod.Wrapf(err, "failed to set sequence to %d", acc.GetSequence()+1)
		}

		ak.SetAccount(ctx, acc)
	}

	return nil
}
