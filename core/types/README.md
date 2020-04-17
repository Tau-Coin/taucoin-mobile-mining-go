# transaction.go
TXsJSON, flexible bytes;
= {
         version:  uint8,          //"0x1" as default;
        chainid: [32+ 4+ 32]byte // `Nickname`+ `blocktime` + hash(signature(timestampinRelaySwitchTimeUnit))
sender: [20]byte        //TAU address, for IPLD codec
nounce: uint64
timestamp:  uint32,    //unix timestamp
expiredTime: uint16,   //unit-minute-default 1440
txfee: uint32,             //unit-tcent: [0, 2^32-1 ]
`ChainIDsenderAddress`IPFSsig; [65]byte, //IPFS signature on `ChainIDsenderAddress` to proof association. Verifier decodes siganture to derive IPFSaddress QM..; 
ChainIDsenderOtherInfo: [128]rune,  // nickname, telegramid, etal
msg:  {
    msg: [2048]rune,      //Descriptions of the share file
    FileAMTRoot; [128]byte,

    receiver: [20]byte,  //TAU address
    amount: uint32,     //unit-tcent: [0, 2^32-1 ]

    RelayNounceAddress: [128]byte, //Multiaddress id in IPFS
}
signature: [65]byte,  //r: 32 bytes, s: 32 bytes, v: 1 byte
}
* hamt_update(`Tminer`Balance,`Tminer`Balance + txfee); // update balance 
* hamt_update(`Tminer`TXnounce,`Tminer`TXounce + 1); // for the coinbase tx nounce increase
* hamt_update(`Tsender`Balance,`Tsender`Balance - txfee);
* hamt_update(`Tsender`TXnounce,`Tsender`TXounce + 1);
* stateroot.hamt_add(`Tsender`nounceTXJSON, msg); // when user follow tsender, can traver its files.
1. File tx:
{
// the File processing
// 1. tgz then use ipfs block standard size e.g. 250k to chop the data to m pieceis
// 2. newNode.amt(1,piece(1)); loop to newNode.hamt(m,piece(m));
// 3. FileAMTroot=AMT_flush_put()
// 4. return FileAMTroot to contract Json. 
}
* hamt_upate(`fileAMTroot``ChainID`SeedingNounce, `fileAMTroot``ChainID`SeedingNounce+1);
* hamt_add  (`fileAMTroot``ChainID`Seeding`Nounce`IPFSpeer, `ChainID``Tsender`IPFSaddr) // seeding peer ipfs id, the first seeder is the creator of the file.

2. Wiring coins tx:
* hamt_update(`Tsender`Balance,`Tsender`Balance - amount);
* hamt_update(`Ttxreceiver`Balance,`Ttxreceiver`Balance + amount);
* hamt_update(`Treceiver`TXnounce,`Treceiver`TXnounce++);
* stateroot.hamt_add(`Treceiver`nounceTXJSON, msg); // when user follow tsender, can traver its files.

3. Relay annoucement operation
* stateroot.hamt_update(RelayNounce , ++)
* stateroot.hamt_add(RelayNounceAddress, msg)
