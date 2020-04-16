package ipldtau

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	cid "github.com/ipfs/go-cid"
	node "github.com/ipfs/go-ipld-format"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	types "github.com/Tau-Coin/taucoin-mobile-mining-go/core/types"
	rlp "github.com/Tau-Coin/taucoin-mobile-mining-go/rlp"
)

// TauBlock (tau-block, codec 0x90), represents an tauereum block header
type TauBlock struct {
	*types.Header

	cid     cid.Cid
	rawdata []byte
}

// Static (compile time) check that TauBlock satisfies the node.Node interface.
var _ node.Node = (*TauBlock)(nil)

/*
  INPUT
*/

// FromBlockRLP takes an RLP message representing
// an tauereum block header or body (header, ommers and txs)
// to return it as a set of IPLD nodes for further processing.
func FromBlockRLP(r io.Reader) (*TauBlock, []*TauTx, []*TauTxTrie, error) {
	// We may want to use this stream several times
	rawdata, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, nil, err
	}

	// Let's try to decode the received element as a block body
	var decodedBlock types.Block
	err = rlp.Decode(bytes.NewBuffer(rawdata), &decodedBlock)
	if err != nil {
		if err.Error()[:41] != "rlp: expected input list for types.Header" {
			return nil, nil, nil, err
		}

		// Maybe it is just a header... (body sans ommers and txs)
		var decodedHeader types.Header
		err := rlp.Decode(bytes.NewBuffer(rawdata), &decodedHeader)
		if err != nil {
			return nil, nil, nil, err
		}

		// It was a header
		return &TauBlock{
			Header:  &decodedHeader,
			cid:     rawdataToCid(MTauBlock, rawdata),
			rawdata: rawdata,
		}, nil, nil, nil
	}

	// This is a block body (header + ommers + txs)
	// We'll extract the header bits here
	headerRawData := getRLP(decodedBlock.Header())
	tauBlock := &TauBlock{
		Header:  decodedBlock.Header(),
		cid:     rawdataToCid(MTauBlock, headerRawData),
		rawdata: headerRawData,
	}

	// Process the found tau-tx objects
	tauTxNodes, tauTxTrieNodes, err := processTransactions(decodedBlock.Transactions(),
		decodedBlock.Header().TxHash[:])
	if err != nil {
		return nil, nil, nil, err
	}

	return tauBlock, tauTxNodes, tauTxTrieNodes, nil
}

// FromBlockJSON takes the output of an tauereum client JSON API
// (i.e. parity or gtau) and returns a set of IPLD nodes.
func FromBlockJSON(r io.Reader) (*TauBlock, []*TauTx, []*TauTxTrie, error) {
	var obj objJSONBlock
	dec := json.NewDecoder(r)
	err := dec.Decode(&obj)
	if err != nil {
		return nil, nil, nil, err
	}

	headerRawData := getRLP(obj.Result.Header)
	tauBlock := &TauBlock{
		Header:  &obj.Result.Header,
		cid:     rawdataToCid(MTauBlock, headerRawData),
		rawdata: headerRawData,
	}

	// Process the found tau-tx objects
	tauTxNodes, tauTxTrieNodes, err := processTransactions(obj.Result.Transactions,
		obj.Result.Header.TxHash[:])
	if err != nil {
		return nil, nil, nil, err
	}

	return tauBlock, tauTxNodes, tauTxTrieNodes, nil
}

// processTransactions will take the found transactions in a parsed block body
// to return IPLD node slices for tau-tx and tau-tx-trie
func processTransactions(txs []*types.Transaction, expectedTxRoot []byte) ([]*TauTx, []*TauTxTrie, error) {
	var tauTxNodes []*TauTx
	//transactionTrie := newTxTrie()
	var tauTxTrieNodes []*TauTxTrie

	return tauTxNodes, tauTxTrieNodes, nil
}

/*
  OUTPUT
*/

// DecodeTauBlock takes a cid and its raw binary data
// from IPFS and returns an TauBlock object for further processing.
func DecodeTauBlock(c *cid.Cid, b []byte) (*TauBlock, error) {
	var h types.Header
	err := rlp.Decode(bytes.NewReader(b), &h)
	if err != nil {
		return nil, err
	}

	return &TauBlock{
		Header:  &h,
		cid:     *c,
		rawdata: b,
	}, nil
}

/*
  Block INTERFACE
*/

// RawData returns the binary of the RLP encode of the block header.
func (b *TauBlock) RawData() []byte {
	return b.rawdata
}

// Cid returns the cid of the block header.
func (b *TauBlock) Cid() cid.Cid {
	return b.cid
}

// String is a helper for output
func (b *TauBlock) String() string {
	return fmt.Sprintf("<TauBlock %s>", b.cid)
}

// Loggable returns a map the type of IPLD Link.
func (b *TauBlock) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"type": "tau-block",
	}
}

/*
  Node INTERFACE
*/

// Resolve resolves a path through this node, stopping at any link boundary
// and returning the object found as well as the remaining path to traverse
func (b *TauBlock) Resolve(p []string) (interface{}, []string, error) {
	if len(p) == 0 {
		return b, nil, nil
	}

	first, rest := p[0], p[1:]

	switch first {
	case "tx":
		return &node.Link{Cid: commonHashToCid(MTauTxTrie, b.TxHash)}, rest, nil
	}

	if len(p) != 1 {
		return nil, nil, fmt.Errorf("unexpected path elements past %s", first)
	}

	switch first {
	case "difficulty":
		return b.Difficulty, nil, nil
	case "number":
		return b.Number, nil, nil
	case "time":
		return b.Time, nil, nil
	default:
		return nil, nil, fmt.Errorf("no such link")
	}
}

// Tree lists all paths within the object under 'path', and up to the given depth.
// To list the entire object (similar to `find .`) pass "" and -1
func (b *TauBlock) Tree(p string, depth int) []string {
	if p != "" || depth == 0 {
		return nil
	}

	return []string{
		"time",
		"bloom",
		"coinbase",
		"difficulty",
		"extra",
		"gaslimit",
		"gasused",
		"mixdigest",
		"nonce",
		"number",
		"parent",
		"receipts",
		"root",
		"tx",
		"uncles",
	}
}

// ResolveLink is a helper function that allows easier traversal of links through blocks
func (b *TauBlock) ResolveLink(p []string) (*node.Link, []string, error) {
	obj, rest, err := b.Resolve(p)
	if err != nil {
		return nil, nil, err
	}

	if lnk, ok := obj.(*node.Link); ok {
		return lnk, rest, nil
	}

	return nil, nil, fmt.Errorf("resolved item was not a link")
}

// Copy will go away. It is here to comply with the Node interface.
func (b *TauBlock) Copy() node.Node {
	panic("dont use this yet")
}

// Links is a helper function that returns all links within this object
// HINT: Use `ipfs refs <cid>`
func (b *TauBlock) Links() []*node.Link {
	return []*node.Link{
		&node.Link{Cid: commonHashToCid(MTauBlock, b.ParentHash)},
		&node.Link{Cid: commonHashToCid(MTauTxTrie, b.TxHash)},
	}
}

// Stat will go away. It is here to comply with the Node interface.
func (b *TauBlock) Stat() (*node.NodeStat, error) {
	return &node.NodeStat{}, nil
}

// Size will go away. It is here to comply with the Node interface.
func (b *TauBlock) Size() (uint64, error) {
	return 0, nil
}

/*
  TauBlock functions
*/

// MarshalJSON processes the block header into readable JSON format,
// converting the right links into their cids, and keeping the original
// hex hash, allowing the user to simplify external queries.
func (b *TauBlock) MarshalJSON() ([]byte, error) {
	out := map[string]interface{}{
		"time":       b.Time,
		"difficulty": b.Difficulty,
		"number":     b.Number,
		"parent":     commonHashToCid(MTauBlock, b.ParentHash),
	}
	return json.Marshal(out)
}

// objJSONBlock defines the output of the JSON RPC API for either
// "tau_BlockByHash" or "tau_BlockByHeader".
type objJSONBlock struct {
	Result objJSONBlockResult `json:"result"`
}

// objJSONBLockResult is the  nested struct that takes
// the contents of the JSON field "result".
type objJSONBlockResult struct {
	types.Header           // Use its fields and unmarshaler
	*objJSONBlockResultExt // Add these fields to the parsing
}

// objJSONBLockResultExt facilitates the composition
// of the field "result", adding to the
// `types.Header` fields, both ommers (their hashes) and transactions.
type objJSONBlockResultExt struct {
	OmmerHashes  []common.Hash        `json:"uncles"`
	Transactions []*types.Transaction `json:"transactions"`
}

// UnmarshalJSON overrides the function types.Header.UnmarshalJSON, allowing us
// to parse the fields of Header, plus ommer hashes and transactions.
// (yes, ommer hashes. You will need to "tau_getUncleCountByBlockHash" per each ommer)
func (o *objJSONBlockResult) UnmarshalJSON(input []byte) error {
	err := o.Header.UnmarshalJSON(input)
	if err != nil {
		return err
	}

	o.objJSONBlockResultExt = &objJSONBlockResultExt{}
	err = json.Unmarshal(input, o.objJSONBlockResultExt)
	return err
}
