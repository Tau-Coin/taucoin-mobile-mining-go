package userdb

import (

	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/taudb"

	cid "github.com/ipfs/go-cid"
)

type Userdb struct{

	ldb taudb.KeyValueStore

	chainInfo		map[common.ChainID]ChainConfig
	blockRoots		map[common.ChainID]cid.Cid

	mutableRange	map[common.ChainID]RangeConfig
	pruneRange		map[common.ChainID]RangeConfig

	ipldPeers		map[common.ChainID]map[common.IPLDPeerID]PeerConfig
	relayList		map[common.RelayType]map[common.RelayMultiAdd]RelayConfig

	followsRepoList	map[common.IPLDPeerID]RepoConfig

	txsPool			map[common.ChainID]map[common.Hash]TxConfig
	filesPool		map[common.ChainID]map[common.Hash]FileConfig

	immutablePoints			map[common.ChainID]cid.Cid
	votesCountingPoints		map[common.ChainID]cid.Cid

	//Lock will be added later
}

func NewUserdb(db taudb.KeyValueStore) *Userdb {
	//read chainInfo from leveldb
	return &Userdb{
		ldb: db,
		chainInfo: make(map[common.ChainID]ChainConfig),
		blockRoots: make(map[common.ChainID]cid.Cid),

		mutableRange: make(map[common.ChainID]RangeConfig),
		pruneRange: make(map[common.ChainID]RangeConfig),
	}
}

func (udb *Userdb) AddNewChain(chainid common.ChainID){
	udb.chainInfo[chainid]= ChainConfig{
		follow: 0,
	}
}

func (udb *Userdb) FollowNewChain(chainid common.ChainID){
	udb.chainInfo[chainid] = ChainConfig{
		follow: 1,
	}
}

func (udb *Userdb) UnfollowChain(chainid common.ChainID){
	udb.chainInfo[chainid] = ChainConfig{
		follow: 0,
	}
}
