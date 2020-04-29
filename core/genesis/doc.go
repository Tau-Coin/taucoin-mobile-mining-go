// Copyright 2020 The TauCoin Authors
// This file is part of the TauCoin library.
//
// The TauCoin library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The TauCoin library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// for more details of LGPL, see <http://www.gnu.org/licenses/>.
// maintained by likeopen
//*********************************************************//
//1.0 how to construct a genesis block
//2.0 if genesis block was created, other variable will be populated
//*********************************************************//

// 1.0 how to construct a genesis block g
// 1.1 check aplication current runing context (background)
// 1.2 background will tell you wheather there are some contract chains in app.
// 1.3 if there are chains in background, using tau address private key pair existing
// 	else new private key pairs.
// 1.4 if there are chains in background, some other informations have also been known.
// 	like tauRelay[tau], succeedRelay[ipfsAddr]config
// 	config contains time that this was connected successfully in unit relayswithtimeunit=15s
// 1.5 please specify below value a config file ?
// 	version uint64
// 	timestampinrelaytimeunit uint32 (example current unix time 1500000 to 1500015 will have same birthday unix/15)
// 	blocknum int32 default 0
// 	previousblockroot nil
// 	basetarget [32]byte default ?
// 	cummulativedifficulty int64 / [32]byte default = 0 at here.
// 	generationSignature [32]byte random
// 	IPLDsigon [65]byte node ipfs private key sign tauaddr to show relationship ipfs and tauaddress locally.
//msg []byte contains initalK-V that is initial state. nonce is previous value in tau and 0 in others, 1th block is allowed to be mined with 0
// 	chainid [32 + 32]byte (nickname [32]byte,hash(timestampinrelaytimeunit and tau private key))
// 	signature tau private key sign this contract ,others can recover genesis tau address
// 1.6 construct genesis block g if before have no error
// 2.0 execuate genesis will update IPLD block store temporary variable and leveldb locally.
// 	IPLD blockchain(blockstore):
// 	1 contract per hamt here will be create , and update from now on.
// 	bs := new blockstore()
//      rootcid = bs.put(cbor.encode(g))

// 	temporary vaiable:
// 	currentchainid = above got.

// 	db:
// 	dbChains.put(chainid,...)
// 	dbBlockRoot.put(chainid,rootcid)
//      dbState.put(chainid,state associated to tauaddress)//fix me
//      dbPeers.put(chianid,addres included taugenesis and ipfs besides config)
//      dbRelay.put(chaind,relay multiaddr and config) //keep ratio 2:1:7 to get network alive healthy,from 3 aspect
package genesis
