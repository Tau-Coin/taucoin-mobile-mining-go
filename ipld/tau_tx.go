package ipldtau

import (
	"bytes"
	"encoding/json"
	"fmt"

	cid "github.com/ipfs/go-cid"
	node "github.com/ipfs/go-ipld-format"

	hexutil "github.com/Tau-Coin/taucoin-mobile-mining-go/common/hexutil"
	types "github.com/Tau-Coin/taucoin-mobile-mining-go/core/types"
	rlp "github.com/Tau-Coin/taucoin-mobile-mining-go/rlp"
)

// TauTx (tau-tx codec 0x93) represents an tauereum transaction
type TauTx struct {
	types.Transaction
	cid     cid.Cid
	rawdata []byte
}

// Static (compile time) check that TauTx satisfies the node.Node interface.
var _ node.Node = (*TauTx)(nil)

/*
  INPUT
*/

// NewTx computes the cid and rlp-encodes a types.Transaction object
// returning a proper TauTx node
func NewTx(t types.Transaction) *TauTx {
	buf := new(bytes.Buffer)
	if err := t.EncodeRLP(buf); err != nil {
		panic(err)
	}
	rawdata := buf.Bytes()

	return &TauTx{
		Transaction: t,
		cid:         rawdataToCid(MTauTx, rawdata),
		rawdata:     rawdata,
	}
}

/*
 OUTPUT
*/

// DecodeTauTx takes a cid and its raw binary data
// from IPFS and returns an TauTx object for further processing.
func DecodeTauTx(c *cid.Cid, b []byte) (*TauTx, error) {
	var t types.Transaction
	err := rlp.DecodeBytes(b, &t)
	if err != nil {
		return nil, err
	}

	return &TauTx{
		Transaction: t,
		cid:         *c,
		rawdata:     b,
	}, nil
}

/*
  Block INTERFACE
*/

// RawData returns the binary of the RLP encode of the transaction.
func (t *TauTx) RawData() []byte {
	return t.rawdata
}

// Cid returns the cid of the transaction.
func (t *TauTx) Cid() cid.Cid {
	return t.cid
}

// String is a helper for output
func (t *TauTx) String() string {
	return fmt.Sprintf("<TauereumTx %s>", t.cid)
}

// Loggable returns in a map the type of IPLD Link.
func (t *TauTx) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"type": "tau-tx",
	}
}

/*
  Node INTERFACE
*/

// Resolve resolves a path through this node, stopping at any link boundary
// and returning the object found as well as the remaining path to traverse
func (t *TauTx) Resolve(p []string) (interface{}, []string, error) {
	if len(p) == 0 {
		return t, nil, nil
	}

	if len(p) > 1 {
		return nil, nil, fmt.Errorf("unexpected path elements past %s", p[0])
	}

	switch p[0] {

	case "nonce":
		return t.Nonce(), nil, nil
	case "r":
		_, r, _ := t.RawSignatureValues()
		return hexutil.EncodeBig(r), nil, nil
	case "s":
		_, _, s := t.RawSignatureValues()
		return hexutil.EncodeBig(s), nil, nil
	case "v":
		v, _, _ := t.RawSignatureValues()
		return hexutil.EncodeBig(v), nil, nil
	default:
		return nil, nil, fmt.Errorf("no such link")
	}
}

// Tree lists all paths within the object under 'path', and up to the given depth.
// To list the entire object (similar to `find .`) pass "" and -1
func (t *TauTx) Tree(p string, depth int) []string {
	if p != "" || depth == 0 {
		return nil
	}
	return []string{"gas", "gasPrice", "input", "nonce", "r", "s", "toAddress", "v", "value"}
}

// ResolveLink is a helper function that calls resolve and asserts the
// output is a link
func (t *TauTx) ResolveLink(p []string) (*node.Link, []string, error) {
	obj, rest, err := t.Resolve(p)
	if err != nil {
		return nil, nil, err
	}

	if lnk, ok := obj.(*node.Link); ok {
		return lnk, rest, nil
	}

	return nil, nil, fmt.Errorf("resolved item was not a link")
}

// Copy will go away. It is here to comply with the interface.
func (t *TauTx) Copy() node.Node {
	panic("dont use this yet")
}

// Links is a helper function that returns all links within this object
func (t *TauTx) Links() []*node.Link {
	return nil
}

// Stat will go away. It is here to comply with the interface.
func (t *TauTx) Stat() (*node.NodeStat, error) {
	return &node.NodeStat{}, nil
}

// Size will go away. It is here to comply with the interface.
func (t *TauTx) Size() (uint64, error) {
	return uint64(0), nil
}

/*
  TauTx functions
*/

// MarshalJSON processes the transaction into readable JSON format.
func (t *TauTx) MarshalJSON() ([]byte, error) {
	v, r, s := t.RawSignatureValues()

	out := map[string]interface{}{
		"nonce":     t.Nonce(),
		"r":         hexutil.EncodeBig(r),
		"s":         hexutil.EncodeBig(s),
		"v":         hexutil.EncodeBig(v),
	}
	return json.Marshal(out)
}
