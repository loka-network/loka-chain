package statedb_test

import (
	"bytes"
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/loka-network/loka/v1/x/evm/statedb"
	evmtypes "github.com/loka-network/loka/v1/x/evm/types"
)

var (
	_             statedb.Keeper = &MockKeeper{}
	errAddress    common.Address = common.BigToAddress(big.NewInt(100))
	emptyCodeHash                = crypto.Keccak256(nil)
)

type MockAccount struct {
	account statedb.Account
	states  statedb.Storage
}

type MockKeeper struct {
	accounts map[common.Address]MockAccount
	codes    map[common.Hash][]byte
}

func NewMockKeeper() *MockKeeper {
	return &MockKeeper{
		accounts: make(map[common.Address]MockAccount),
		codes:    make(map[common.Hash][]byte),
	}
}

func (k *MockKeeper) AddBalance(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) error {
	if common.Address(addr.Bytes()) == errAddress {
		return errors.New("mock db error")
	}
	acct, exists := k.accounts[common.BytesToAddress(addr)]
	if !exists {
		return errors.New("account not found")
	}
	acct.account.Balance.Add(acct.account.Balance, coins.AmountOf(coins.Denoms()[0]).BigInt())
	k.accounts[common.BytesToAddress(addr)] = acct
	return nil
}

func (k *MockKeeper) SubBalance(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) error {
	if common.Address(addr.Bytes()) == errAddress {
		return errors.New("mock db error")
	}
	acct, exists := k.accounts[common.BytesToAddress(addr)]
	if !exists {
		return errors.New("account not found")
	}
	acct.account.Balance.Sub(acct.account.Balance, coins.AmountOf(coins.Denoms()[0]).BigInt())
	k.accounts[common.BytesToAddress(addr)] = acct
	return nil
}

func (k *MockKeeper) Transfer(ctx sdk.Context, sender, recipient sdk.AccAddress, coins sdk.Coins) error {
	if common.Address(sender.Bytes()) == errAddress || common.Address(recipient.Bytes()) == errAddress {
		return errors.New("mock db error")
	}
	if err := k.SubBalance(ctx, sender, coins); err != nil {
		return err
	}
	return k.AddBalance(ctx, recipient, coins)
}

func (k *MockKeeper) SetBalance(ctx sdk.Context, addr common.Address, amount *big.Int, evmDenom string) error {
	acct, exists := k.accounts[addr]
	if !exists {
		return errors.New("account not found")
	}
	acct.account.Balance = new(big.Int).Set(amount)
	k.accounts[addr] = acct
	return nil
}

func (k *MockKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, evmDenom string) *big.Int {
	acct, exists := k.accounts[common.Address(addr.Bytes())]
	if !exists {
		return big.NewInt(0)
	}
	return acct.account.Balance
}

func (k MockKeeper) GetAccount(_ sdk.Context, addr common.Address) *statedb.Account {
	acct, ok := k.accounts[addr]
	if !ok {
		return nil
	}
	return &acct.account
}

func (k MockKeeper) GetState(_ sdk.Context, addr common.Address, key common.Hash) common.Hash {
	return k.accounts[addr].states[key]
}

func (k MockKeeper) GetCode(_ sdk.Context, codeHash common.Hash) []byte {
	return k.codes[codeHash]
}

func (k MockKeeper) ForEachStorage(_ sdk.Context, addr common.Address, cb func(key, value common.Hash) bool) {
	if acct, ok := k.accounts[addr]; ok {
		for k, v := range acct.states {
			if !cb(k, v) {
				return
			}
		}
	}
}

func (k MockKeeper) SetAccount(_ sdk.Context, addr common.Address, account statedb.Account) error {
	if addr == errAddress {
		return errors.New("mock db error")
	}
	acct, exists := k.accounts[addr]
	if exists {
		// update
		acct.account = account
		k.accounts[addr] = acct
	} else {
		k.accounts[addr] = MockAccount{account: account, states: make(statedb.Storage)}
	}
	return nil
}

func (k MockKeeper) SetState(_ sdk.Context, addr common.Address, key common.Hash, value []byte) {
	if acct, ok := k.accounts[addr]; ok {
		if len(value) == 0 {
			delete(acct.states, key)
		} else {
			acct.states[key] = common.BytesToHash(value)
		}
	}
}

func (k MockKeeper) SetCode(_ sdk.Context, codeHash []byte, code []byte) {
	k.codes[common.BytesToHash(codeHash)] = code
}

func (k MockKeeper) DeleteAccount(_ sdk.Context, addr common.Address) error {
	if addr == errAddress {
		return errors.New("mock db error")
	}
	old := k.accounts[addr]
	delete(k.accounts, addr)
	if !bytes.Equal(old.account.CodeHash, emptyCodeHash) {
		delete(k.codes, common.BytesToHash(old.account.CodeHash))
	}
	return nil
}

func (k MockKeeper) Clone() *MockKeeper {
	accounts := make(map[common.Address]MockAccount, len(k.accounts))
	for k, v := range k.accounts {
		accounts[k] = v
	}
	codes := make(map[common.Hash][]byte, len(k.codes))
	for k, v := range k.codes {
		codes[k] = v
	}
	return &MockKeeper{accounts, codes}
}

func (k *MockKeeper) GetParams(ctx sdk.Context) evmtypes.Params {
	return evmtypes.Params{
		EvmDenom:            "aloka",
		EnableCreate:        true,
		EnableCall:          true,
		ExtraEIPs:           []int64{},
		ChainConfig:         evmtypes.ChainConfig{},
		AllowUnprotectedTxs: false,
	}
}
