package ipldtau

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	block "github.com/ipfs/go-block-format"
)

/*
  TauBlock
  INPUT
*/

func TestTxInBlockBodyRlpParsing(t *testing.T) {
	fi, err := os.Open("test_data/tau-block-body-rlp-999999")
	checkError(err, t)

	_, output, _, err := FromBlockRLP(fi)
	checkError(err, t)

	if len(output) != 11 {
		t.Fatal("Wrong number of parsed txs")
	}

	// Oh, let's just grab the last element and one from the middle
	testTx05Fields(output[5], t)
	testTx10Fields(output[10], t)
}

func TestTxInBlockHeaderRlpParsing(t *testing.T) {
	fi, err := os.Open("test_data/tau-block-header-rlp-999999")
	checkError(err, t)

	_, output, _, err := FromBlockRLP(fi)
	checkError(err, t)

	if len(output) != 0 {
		t.Fatal("No transactions should have been gotten from here")
	}
}

func TestTxInBlockBodyJsonParsing(t *testing.T) {
	fi, err := os.Open("test_data/tau-block-body-json-999999")
	checkError(err, t)

	_, output, _, err := FromBlockJSON(fi)
	checkError(err, t)

	if len(output) != 11 {
		t.Fatal("Wrong number of parsed txs")
	}

	testTx05Fields(output[5], t)
	testTx10Fields(output[10], t)
}

/*
  OUTPUT
*/

func TestDecodeTransaction(t *testing.T) {
	// Prepare the "fetched transaction".
	// This one is supposed to be in the datastore already,
	// and given away by github.com/ipfs/go-ipfs/merkledag
	rawTransactionString :=
		"f86c34850df84758008252089432be343b94f860124dc4fee278fdcbd38c102d88880f25" +
			"8512af0d4000801ba0e9a25c929c26d1a95232ba75aef419a91b470651eb77614695e16c" +
			"5ba023e383a0679fb2fc0d0b0f3549967c0894ee7d947f07d238a83ef745bc3ced5143a4af36"
	rawTransaction, err := hex.DecodeString(rawTransactionString)
	checkError(err, t)
	c := rawdataToCid(MTauTx, rawTransaction)

	// Just to clarify: This `block` is an IPFS block
	storedTransaction, err := block.NewBlockWithCid(rawTransaction, c)
	checkError(err, t)

	// Now the proper test
	tauTransaction, err := DecodeTauTx(storedTransaction.Cid(), storedTransaction.RawData())
	checkError(err, t)

	testTx05Fields(tauTransaction, t)
}

/*
  Block INTERFACE
*/

func TestTauTxLoggable(t *testing.T) {
	txs := prepareParsedTxs(t)

	l := txs[0].Loggable()
	if _, ok := l["type"]; !ok {
		t.Fatal("Loggable map expected the field 'type'")
	}

	if l["type"] != "tau-tx" {
		t.Fatal("Wrong Loggable 'type' value")
	}
}

/*
  Node INTERFACE
*/

func TestTauTxResolve(t *testing.T) {
	tx := prepareParsedTxs(t)[0]

	// Empty path
	obj, rest, err := tx.Resolve([]string{})
	rtx, ok := obj.(*TauTx)
	if !ok {
		t.Fatal("Wrong type of returned object")
	}
	if rtx.Cid() != tx.Cid() {
		t.Fatal("wrong returned object")
	}
	if rest != nil {
		t.Fatal("rest should be nil")
	}
	if err != nil {
		t.Fatal("err should be nil")
	}

	// len(p) > 1
	badCases := [][]string{
		[]string{"two", "elements"},
		[]string{"here", "three", "elements"},
		[]string{"and", "here", "four", "elements"},
	}

	for _, bc := range badCases {
		obj, rest, err = tx.Resolve(bc)
		if obj != nil {
			t.Fatal("obj should be nil")
		}
		if rest != nil {
			t.Fatal("rest should be nil")
		}
		if err.Error() != fmt.Sprintf("unexpected path elements past %s", bc[0]) {
			t.Fatal("wrong error")
		}
	}

	moreBadCases := []string{
		"i",
		"am",
		"not",
		"a",
		"tx",
		"field",
	}
	for _, mbc := range moreBadCases {
		obj, rest, err = tx.Resolve([]string{mbc})
		if obj != nil {
			t.Fatal("obj should be nil")
		}
		if rest != nil {
			t.Fatal("rest should be nil")
		}
		if err.Error() != fmt.Sprintf("no such link") {
			t.Fatal("wrong error")
		}
	}

	goodCases := []string{
		"gas",
		"gasPrice",
		"input",
		"nonce",
		"r",
		"s",
		"toAddress",
		"v",
		"value",
	}
	for _, gc := range goodCases {
		_, _, err = tx.Resolve([]string{gc})
		if err != nil {
			t.Fatalf("error should be nil %v", gc)
		}
	}

}

func TestTauTxTree(t *testing.T) {
	tx := prepareParsedTxs(t)[0]
	_ = tx

	// Bad cases
	tree := tx.Tree("non-empty-string", 0)
	if tree != nil {
		t.Fatal("Expected nil to be returned")
	}

	tree = tx.Tree("non-empty-string", 1)
	if tree != nil {
		t.Fatal("Expected nil to be returned")
	}

	tree = tx.Tree("", 0)
	if tree != nil {
		t.Fatal("Expected nil to be returned")
	}

	// Good cases
	tree = tx.Tree("", 1)
	lookupElements := map[string]interface{}{
		"gas":       nil,
		"gasPrice":  nil,
		"input":     nil,
		"nonce":     nil,
		"r":         nil,
		"s":         nil,
		"toAddress": nil,
		"v":         nil,
		"value":     nil,
	}

	if len(tree) != len(lookupElements) {
		t.Fatalf("Wrong number of elements. Got %d. Expecting %d", len(tree), len(lookupElements))
	}

	for _, te := range tree {
		if _, ok := lookupElements[te]; !ok {
			t.Fatalf("Unexpected Element: %v", te)
		}
	}
}

func TestTauTxResolveLink(t *testing.T) {
	tx := prepareParsedTxs(t)[0]

	// bad case
	obj, rest, err := tx.ResolveLink([]string{"supercalifragilist"})
	if obj != nil {
		t.Fatalf("Expected obj to be nil")
	}
	if rest != nil {
		t.Fatal("Expected rest to be nil")
	}
	if err.Error() != "no such link" {
		t.Fatal("Wrong error")
	}

	// good case
	obj, rest, err = tx.ResolveLink([]string{"nonce"})
	if obj != nil {
		t.Fatalf("Expected obj to be nil")
	}
	if rest != nil {
		t.Fatal("Expected rest to be nil")
	}
	if err.Error() != "resolved item was not a link" {
		t.Fatal("Wrong error")
	}
}

func TestTauTxCopy(t *testing.T) {
	tx := prepareParsedTxs(t)[0]

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic")
		}
		if r != "dont use this yet" {
			t.Fatal("Expected panic message 'dont use this yet'")
		}
	}()

	_ = tx.Copy()
}

func TestTauTxLinks(t *testing.T) {
	tx := prepareParsedTxs(t)[0]

	if tx.Links() != nil {
		t.Fatal("Links() expected to return nil")
	}
}

func TestTauTxStat(t *testing.T) {
	tx := prepareParsedTxs(t)[0]

	obj, err := tx.Stat()
	if obj == nil {
		t.Fatal("Expected a not null object node.NodeStat")
	}

	if err != nil {
		t.Fatal("Expected a nil error")
	}
}

func TestTauTxSize(t *testing.T) {
	tx := prepareParsedTxs(t)[0]

	size, err := tx.Size()
	if size != uint64(tx.Transaction.Size().Int64()) {
		t.Fatal("Expected a size equal to 0")
	}

	if err != nil {
		t.Fatal("Expected a nil error")
	}
}

/*
  AUXILIARS
*/

// prepareParsedTxs is a convenienve mtauod
func prepareParsedTxs(t *testing.T) []*TauTx {
	fi, err := os.Open("test_data/tau-block-body-rlp-999999")
	checkError(err, t)

	_, output, _, err := FromBlockRLP(fi)
	checkError(err, t)

	return output
}

func testTx05Fields(tauTx *TauTx, t *testing.T) {
	// Was the cid calculated?
	if tauTx.Cid().String() != "z44VCrqacegDLXw385vC4tZi84ifPengFdSqbLveMRmsFBeDdNs" {
		t.Fatal("Wrong cid")
	}

	// Do we have the rawdata available?
	if fmt.Sprintf("%x", tauTx.RawData()[:10]) != "f86c34850df847580082" {
		t.Fatal("Wrong Rawdata")
	}

	// Proper Fields of types.Transaction
	if fmt.Sprintf("%x", tauTx.To()) != "32be343b94f860124dc4fee278fdcbd38c102d88" {
		t.Fatal("Wrong Recipient")
	}
	if len(tauTx.Data()) != 0 {
		t.Fatal("Wrong Data")
	}
	if fmt.Sprintf("%v", tauTx.Gas()) != "21000" {
		t.Fatal("Wrong Gas")
	}
	if fmt.Sprintf("%v", tauTx.Value()) != "1091424800000000000" {
		t.Fatal("Wrong Value")
	}
	if fmt.Sprintf("%v", tauTx.Nonce()) != "52" {
		t.Fatal("Wrong Nonce")
	}
	if fmt.Sprintf("%v", tauTx.GasPrice()) != "60000000000" {
		t.Fatal("Wrong Gas Price")
	}
}

func testTx10Fields(tauTx *TauTx, t *testing.T) {
	// Was the cid calculated?
	if tauTx.Cid().String() != "z44VCrqbjszozB5K5Xqm3tm9YDqrWPE5H9QRpKAZRjCLQFbrctT" {
		t.Fatal("Wrong cid")
	}

	// Do we have the rawdata available?
	if fmt.Sprintf("%x", tauTx.RawData()[:10]) != "f8708302a120850ba43b" {
		t.Fatal("Wrong Rawdata")
	}

	// Proper Fields of types.Transaction
	if fmt.Sprintf("%x", tauTx.To()) != "1c51bf013add0857c5d9cf2f71a7f15ca93d4816" {
		t.Fatal("Wrong Recipient")
	}
	if len(tauTx.Data()) != 0 {
		t.Fatal("Wrong Data")
	}
	if fmt.Sprintf("%v", tauTx.Gas()) != "90000" {
		t.Fatal("Wrong Gas")
	}
	if fmt.Sprintf("%v", tauTx.Value()) != "1049756850000000000" {
		t.Fatal("Wrong Value")
	}
	if fmt.Sprintf("%v", tauTx.Nonce()) != "172320" {
		t.Fatal("Wrong Nonce")
	}
	if fmt.Sprintf("%v", tauTx.GasPrice()) != "50000000000" {
		t.Fatal("Wrong Gas Price")
	}
}
