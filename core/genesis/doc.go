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
//1.0 how to construct a genesis contract
//2.0 if genesis contract was created, other variable will be populated
//*********************************************************//

// 1.0 how to construct a genesis contract Y
// 1.1 check aplication current runing context (background)
// 1.2 background will tell you wheather there are some contract chains in app.
// 1.3 if there are chains in background, using tau address private key pair existing
// 	else new private key pairs.
// 1.4 if there are chains in background, some other informations have also been known.
// 	like tauRelay[tau], succeedRelay[ipfsAddr]config
// 	config contains time that this was connected successfully in unit relayswithtimeunit=15s
// 1.5 please specify below value a config file ?
// 	version uint64
// 	timestampinrelaytimeunit uint64 (example current unix time 1500000 to 1500015 will have same birthday unix/15)
// 	contractnum int32
// 	(
// 	nickname [32]byte, blocktime uint8 default 5min, sign timestampinrelaytimeunit with tau private key
// 	marked sig
// 	hash(sig)
// 	)
// 	chainid [32 + 1 + 32]byte
// 	safetycontractresultroot nil
// 	basetarget [32]byte default ?
// 	cummulativedifficulty int64 / [32]byte default = 0 at here.
// 	generationSignature [32]byte random
// 	txfee uint64 default = 10 000 000 total suply
// 	nonce uint64 default = 0 minerpower, if initial is 0 lead num 0 contract will mined when selfminingtime=60min timeout
// 	ipfssigon [65]byte node ipfs private key sign tauaddr to show relationship ipfs and tauaddress locally.
// 	msg []byte contains optional msg like your vision
// 	signature tau private key sign this contract ,others can recover genesis tau address
// 1.6 construct genesis contract Y if before have no error
// 1.7 execuate genesis will be update hamt,temporary vaiable and db locally.
// 	hamt:
// 	1 contract per hamt here will be create , and update from now on.
// 	hamt := new hamt()
// 	if (!isTau()){
// 		hamt.add(contractJSON,Y)
// 		hamt.add(genesistauaddressbalance, above specify value)
// 		hamt.add(genesistauaddressnone,0)
//      hamt.add(relaynonce,count of relay from relay announce msg if exist)
//      hamt.add(relaynonceaddress, relay msg mutiaddr if exist)
// 		contractresultstateroot = hamt.add(genesistauaddressNoncetxJSON,nil) //empty tx in genesis contract
// 	}else{
// 		hamt.add(contractJSON,Y)
// 		for i:=0;i < 13w;i++{
// 			hamt.add(genesistauaddress[i]balance, above specify value)
// 			hamt.add(genesistauaddress[i]nonce,0)
// 		}
//      hamt.add(relaynonce,count of relay from relay announce msg tau default)
//      hamt.add(relaynonceaddress, relay msg mutiaddr tau default)
// 		contractresultstateroot = hamt.add(genesistauaddressNoncetxJSON,nil) //empty tx in genesis contract
// 	}

// 	temporary vaiable:
// 	currentchainid = above got.

// 	db:
// 	mychainsdb.put(chainid,...)
// 	mycontractresultstaterootdb.put(chainid,contractresultstateroot)
// 	mysafetycontractresultstaterootdb.put(chainid,nil)
// 	mysafetycontractresultstaterootminerdb.put(chainid,genesistauaddress)
// 	myprevioussafetycontractresultstaterootminerdb.put(chainid,nil)
//  mypeersdb.put(chianid,addres included taugenesis and ipfs besides config)
//  myrelaysdb.put(chaind,relay multiaddr and config) //keep ratio 2:1:7 to get network alive healthy,from 3 aspect
//  taurealy(hamt), ownrelay(hamt), succeedrelay(active process)
package genesis
