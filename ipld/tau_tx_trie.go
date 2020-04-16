package ipldtau

import (
	"fmt"

	cid "github.com/ipfs/go-cid"
	//node "github.com/ipfs/go-ipld-format"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/types"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/rlp"
)

// TauTxTrie (tau-tx-trie codec 0x92) represents
// a node from the transaction trie in tauereum.
type TauTxTrie struct {
	*TrieNode
}

// Static (compile time) check that TauTxTrie satisfies the node.Node interface.
//var _ node.Node = (*TauTxTrie)(nil)

/*
 INPUT
*/

// To create a proper trie of the tau-tx-trie objects, it is required
// to input all transactions belonging to a forest in a single step.
// We are adding the transactions, and creating its trie on
// block body parsing time.

/*
  OUTPUT
*/

// DecodeTauTxTrie returns an TauTxTrie object from its cid and rawdata.
func DecodeTauTxTrie(c *cid.Cid, b []byte) (*TauTxTrie, error) {
	tn, err := decodeTrieNode(c, b, decodeTauTxTrieLeaf)
	if err != nil {
		return nil, err
	}
	return &TauTxTrie{TrieNode: tn}, nil
}

// decodeTauTxTrieLeaf parses a tau-tx-trie leaf
//from decoded RLP elements
func decodeTauTxTrieLeaf(i []interface{}) ([]interface{}, error) {
	var t types.Transaction
	err := rlp.DecodeBytes(i[1].([]byte), &t)
	if err != nil {
		return nil, err
	}
	return []interface{}{
		i[0].([]byte),
		&TauTx{
			Transaction: t,
			cid:         rawdataToCid(MTauTx, i[1].([]byte)),
			rawdata:     i[1].([]byte),
		},
	}, nil
}

/*
  Block INTERFACE
*/

// RawData returns the binary of the RLP encode of the transaction.
func (t *TauTxTrie) RawData() []byte {
	return t.rawdata
}

// Cid returns the cid of the transaction.
func (t *TauTxTrie) Cid() *cid.Cid {
	return t.cid
}

// String is a helper for output
func (t *TauTxTrie) String() string {
	return fmt.Sprintf("<TauereumTxTrie %s>", t.cid)
}

// Loggable returns in a map the type of IPLD Link.
func (t *TauTxTrie) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"type": "tau-tx-trie",
	}
}

/*
  TauTxTrie functions

// txTrie wraps a localTrie for use on the transaction trie.
type txTrie struct {
	*localTrie
}

// newTxTrie initializes and returns a txTrie.
func newTxTrie() *txTrie {
	return &txTrie{
		localTrie: newLocalTrie(),
	}
}

// getNodes invokes the localTrie, which computes the root hash of the
// transaction trie and returns its database keys, to return a slice
// of TauTxTrie nodes.
func (tt *txTrie) getNodes() []*TauTxTrie {
	keys := tt.getKeys()
	var out []*TauTxTrie

	for _, k := range keys {
		rawdata, err := tt.db.Get(k)
		if err != nil {
			panic(err)
		}

		tn := &TrieNode{
			cid:     rawdataToCid(MTauTxTrie, rawdata),
			rawdata: rawdata,
		}
		out = append(out, &TauTxTrie{TrieNode: tn})
	}

	return out
}
*/
