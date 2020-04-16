// Copyright 2017 The go-tau Authors
// This file is part of the go-tau library.
//
// The go-tau library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-tau library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-tau library. If not, see <http://www.gnu.org/licenses/>.

package tau

import (
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/consensus/tauhash"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/miner"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/params"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/tau/downloader"
)

// DefaultConfig contains default settings for use on the Tau main net.
var DefaultConfig = Config{
	SyncMode: downloader.FastSync,
	Tauash: tauhash.Config{
		CacheDir:       "tauhash",
		CachesInMem:    2,
		CachesOnDisk:   3,
		DatasetsInMem:  1,
		DatasetsOnDisk: 2,
	},
	NetworkId:      1,
	DatabaseCache:  512,
	TrieCleanCache: 256,
	TrieDirtyCache: 256,
	TrieTimeout:    60 * time.Minute,
	Miner: miner.Config{
		FeeFloor: big.NewInt(params.GWei),
		Recommit: 3 * time.Second,
	},
	TxPool: core.DefaultTxPoolConfig,
}

func init() {
	home := os.Getenv("HOME")
	if home == "" {
		if user, err := user.Current(); err == nil {
			home = user.HomeDir
		}
	}
	if runtime.GOOS == "darwin" {
		DefaultConfig.Tauash.DatasetDir = filepath.Join(home, "Library", "Tauash")
	} else if runtime.GOOS == "windows" {
		localappdata := os.Getenv("LOCALAPPDATA")
		if localappdata != "" {
			DefaultConfig.Tauash.DatasetDir = filepath.Join(localappdata, "Tauash")
		} else {
			DefaultConfig.Tauash.DatasetDir = filepath.Join(home, "AppData", "Local", "Tauash")
		}
	} else {
		DefaultConfig.Tauash.DatasetDir = filepath.Join(home, ".tauhash")
	}
}

//go:generate gencodec -type Config -formats toml -out gen_config.go

type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the Tau main net block is used.
	Genesis *core.Genesis `toml:",omitempty"`

	// Protocol options
	NetworkId uint64 // Network ID to use for selecting peers to connect to
	SyncMode  downloader.SyncMode

	NoPruning  bool // Whtauer to disable pruning and flush everything to disk
	NoPrefetch bool // Whtauer to disable prefetching and only load state on demand

	// Whitelist of required block number -> hash values to accept
	Whitelist map[uint64]common.Hash `toml:"-"`

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	DatabaseCache      int
	DatabaseFreezer    string

	TrieCleanCache int
	TrieDirtyCache int
	TrieTimeout    time.Duration

	// Mining options
	Miner miner.Config

	// Tauash options
	Tauash tauhash.Config

	// Transaction pool options
	TxPool core.TxPoolConfig

	// Miscellaneous options
	DocRoot string `toml:"-"`

	// Checkpoint is a hardcoded checkpoint which can be nil.
	Checkpoint *params.TrustedCheckpoint `toml:",omitempty"`
}
