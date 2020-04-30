package userdb

/*
 
1. dbChains                                      map[ChainID] config; //Chains to follow, string is for planned config info
key: []byte("dbchains")
value: json.Marshal(map[ChainID]config)
var ChainID string
type config struct{
   Account Address
   Followed uint8
}
 
2. dbBlockRoots                    map[ChainID] cbor.cid; // the new contract state
key: []byte("dbblockroots")
value: json.Marshal(map[ChainID]cbor.cid)
var ChainID string
type cbor.cid string
 
3. dbMutableRange                    map[ChainID]config;
key: []byte("dbmutablerange")
value: json.Marshal(map[ChainID]string)
var ChainID string
type config struct{
	HeightNum uint64
}
 
4. dbPruneRange                    map[ChainID]config;
key: []byte("dbprunerange")
value: json.Marshal(map[ChainID]string)
var ChainID string
type config struct{
	HeightNum uint64
}
 
5. dbIPLDPeers              map[ChainID]map[IPLDPeerID]config
key: []byte("dbipldpeers")
value: json.Marshal(map[ChainID]map[IPLDPeerID]config)
var ChainID string
 
type IPLDPeerID string
type config struct{
    NickName [32]byte
}
 
6. dbRelays             map[ChainID]map[RelaysMultipleAddr]config;// incllude timestampInRelaySwitchTimeUnit; timestamp is to selelct relays in the mutable ranges. 
key: []byte("dbrelays")
value: json.Marshal(map[ChainID]map[RelaysMultipleAddr]config)
var ChainID string
type relaysMultipleAddr string
type config struct{
    Time uint32 //unixtimestamp/RelaySwitchTime
    ...
}
 
7. dbFollowedIPLDPeersRepo  map[IPLDpeerID]Repo // for each followed IPLDpeer, this keeps the repo info.
 
type IPLDPeerID string
type Repo struct{
	dbTXsPool
	dbSelfFilesPool
}
 
8. dbTXsPool            map[ChainID]map[hash(txjson)]TxJsonConfig;
key: []byte("dbtxspool")
value: json.Marshal(map[ChainID]map[hash(txjson)]TxJsonConfig;)
var ChainID string
hash(txjson) -> string
 
type TxJsonConfig struct{
   Type uint8
   Sender Address
   Nonce uint64   
   Fee uint32
   Tx TxJson
}
 
9. dbSelfFilesPool        map[ChainID]map[FileHash]config; // config describes type(shared, downloading), range progress and parameters. 
key: []byte("dbselffilespool")
value: json.Marshal(map[ChainID]map[FileHash]config)
var ChainID string
type FileHash string
type config struct{
   FileType uint8
   FileSize  uint32  //uint-KB
   FileTime uint32
   Progress uint8
}
 
uint-KB
10. dbtotalFilesDownloadedData	string
key: []byte("dbtotalfilesdownloadeddata")
value: string
11. dbtotalFilesUploadedData	string
key: []byte("dbtotalfilesuploadeddata")
value: string
12. dbtotalTMDownloadedData; TM: transaction and mining string
key: []byte("dbtotaltmdownloadeddata")
value: string
 
13. dbImmutablePoints    map[ChainID] root 
key: []byte("dbimmutablepoints")
value: json.Marshal(map[ChainID]root)
14. dbVotesCountingPoints   map[ChainID] root 
*/
