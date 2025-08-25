//go:build !rocksdb
// +build !rocksdb

package app

import (
	"errors"

	storetypes "cosmossdk.io/store/types"
)

func (app *Evmos) setupVersionDB(
	homePath string,
	keys map[string]*storetypes.KVStoreKey,
	tkeys map[string]*storetypes.TransientStoreKey,
	okeys map[string]*storetypes.ObjectStoreKey,
) (storetypes.RootMultiStore, error) {
	return nil, errors.New("versiondb is not supported in this binary")
}
