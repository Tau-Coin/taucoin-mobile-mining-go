package ipldtau

import (
	"bytes"

	cid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"

	common "github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/rlp"
)

// IPLD Codecs for Tauereum
// See the authoritative document:
// https://github.com/multiformats/multicodec/blob/master/table.csv
const (
	RawBinary           = 0x55
	MTauBlock           = 0xa0
	MTauBlockList       = 0xa1
	MTauTx              = 0xa2
	MTauTxTrie          = 0xa3
)

// rawdataToCid takes the desired codec and a slice of bytes
// and returns the proper cid of the object.
func rawdataToCid(codec uint64, rawdata []byte) cid.Cid {
	c, err := cid.Prefix{
		Codec:    codec,
		Version:  1,
		MhType:   mh.KECCAK_256,
		MhLength: -1,
	}.Sum(rawdata)
	if err != nil {
		panic(err)
	}
	return c
}

// keccak256ToCid takes a keccak256 hash and returns its cid based on
// the codec given.
func keccak256ToCid(codec uint64, h []byte) cid.Cid {
	buf, err := mh.Encode(h, mh.KECCAK_256)
	if err != nil {
		panic(err)
	}

	return cid.NewCidV1(codec, mh.Multihash(buf))
}

// commonHashToCid takes a go-tau common.Hash and returns its
// cid based on the codec given,
func commonHashToCid(codec uint64, h common.Hash) cid.Cid {
	mhash, err := mh.Encode(h[:], mh.KECCAK_256)
	if err != nil {
		panic(err)
	}

	return cid.NewCidV1(codec, mhash)
}

func commonHashToCidString(codec uint64, h common.Hash) string {
	mhash, err := mh.Encode(h[:], mh.KECCAK_256)
	if err != nil {
		panic(err)
	}

	return cid.NewCidV1(codec, mhash).String()
}

// getRLP encodes the given object to RLP returning its bytes.
func getRLP(object interface{}) []byte {
	buf := new(bytes.Buffer)
	if err := rlp.Encode(buf, object); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

/*
  LOCAL TRIE

// localTrie wraps a go-tau trie and its underlying memory db.
// It contributes to the creation of the trie node objects.
type localTrie struct {
	db   *taudb.Database
	trie *trie.Trie
}

// newLocalTrie initializes and returns a localTrie object
func newLocalTrie() *localTrie {
	var err error
	lt := &localTrie{}

	lt.db, err = taudb.NewMemDatabase()
	if err != nil {
		panic(err)
	}

	lt.trie, err = trie.New(common.Hash{}, lt.db)
	if err != nil {
		panic(err)
	}

	return lt
}

// add receives the index of an object and its rawdata value
// and includes it into the localTrie
func (lt *localTrie) add(idx int, rawdata []byte) {
	key, err := rlp.EncodeToBytes(uint(idx))
	if err != nil {
		panic(err)
	}

	lt.trie.Update(key, rawdata)
}

// rootHash returns the computed trie root.
// Useful for sanity checks on parsed data.
func (lt *localTrie) rootHash() []byte {
	return lt.trie.Hash().Bytes()
}

// getKeys returns the stored keys of the memory database
// of the localTrie for further processing.
func (lt *localTrie) getKeys() [][]byte {
	var err error

	_, err = lt.trie.Commit()
	if err != nil {
		panic(err)
	}

	return lt.db.Keys()
}
*/
