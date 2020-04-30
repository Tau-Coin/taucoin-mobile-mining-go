package userdb

import (

	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/types"
)

type ChainConfig struct{
	follow uint8    // 0- unfollow, 1- followed
}

type RangeConfig struct{
	height uint64   // BlockNum
	time uint32		// For pruning block 
}

type PeerConfig struct{
	chainid common.ChainID   // Added with chainid
	blocknum uint64		// Added with blocknum
}

type RelayConfig struct{
	chainid common.ChainID   // Added with chainid
	blocknum uint64		// Added with blocknum
	time uint32			// Added with time
}

type RepoConfig struct{
	txs TxConfig
	files FileConfig
}

type TxConfig struct{
	Type	uint8
	Sender	common.Address
	Nonce	uint64
	Fee		uint32
	TxJson	types.Transaction
}

type FileConfig struct{
	//TBD
	Type uint8 //download or shared
	IpldPeers []common.IPLDPeerID
}
