package ipldtau

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"testing"

	block "github.com/ipfs/go-block-format"
	node "github.com/ipfs/go-ipld-format"

	"github.com/tauereum/go-tauereum/core/types"
)

func TestBlockBodyRlpParsing(t *testing.T) {
	fi, err := os.Open("test_data/tau-block-body-rlp-999999")
	checkError(err, t)

	output, _, _, err := FromBlockRLP(fi)
	checkError(err, t)

	testTauBlockFields(output, t)
}

func TestBlockHeaderRlpParsing(t *testing.T) {
	fi, err := os.Open("test_data/tau-block-header-rlp-999999")
	checkError(err, t)

	output, _, _, err := FromBlockRLP(fi)
	checkError(err, t)

	testTauBlockFields(output, t)
}

func TestBlockBodyJsonParsing(t *testing.T) {
	fi, err := os.Open("test_data/tau-block-body-json-999999")
	checkError(err, t)

	output, _, _, err := FromBlockJSON(fi)
	checkError(err, t)

	testTauBlockFields(output, t)
}

func TestTauBlockProcessTransactionsError(t *testing.T) {
	// Let's just change one byte in a field of one of these transactions.
	fi, err := os.Open("test_data/error-tx-tau-block-body-json-999999")
	checkError(err, t)

	_, _, _, err = FromBlockJSON(fi)
	if err == nil {
		t.Fatal("Expected an error")
	}
}

// TestDecodeBlockHeader should work for both inputs (block header and block body)
// as what we are storing is just the block header
func TestDecodeBlockHeader(t *testing.T) {
	storedTauBlock := prepareStoredTauBlock("test_data/tau-block-header-rlp-999999", t)

	tauBlock, err := DecodeTauBlock(storedTauBlock.Cid(), storedTauBlock.RawData())
	checkError(err, t)

	testTauBlockFields(tauBlock, t)
}

func TestTauBlockString(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	if tauBlock.String() != "<TauBlock z43AaGF4uHSY4waU68L3DLUKHZP7yfZoo6QbLmid5HomZ4WtbWw>" {
		t.Fatalf("Wrong String()")
	}
}

func TestTauBlockLoggable(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	l := tauBlock.Loggable()
	if _, ok := l["type"]; !ok {
		t.Fatal("Loggable map expected the field 'type'")
	}

	if l["type"] != "tau-block" {
		t.Fatal("Wrong Loggable 'type' value")
	}
}

func TestTauBlockJSONMarshal(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	jsonOutput, err := tauBlock.MarshalJSON()
	checkError(err, t)

	var data map[string]interface{}
	err = json.Unmarshal(jsonOutput, &data)
	checkError(err, t)

	// Testing all fields is boring, but can help us to avoid
	// that dreaded regression
	if data["bloom"].(string)[:10] != "0x00000000" {
		t.Fatal("Wrong Bloom")
	}
	if data["coinbase"] != "0x52bc44d5378309ee2abf1539bf71de1b7d7be3b5" {
		t.Fatal("Wrong Coinbase")
	}
	if parseFloat(data["difficulty"]) != "12555463106190" {
		t.Fatal("Wrong Difficulty")
	}
	if data["extra"] != "0xd783010303844765746887676f312e342e32856c696e7578" {
		t.Fatal("Wrong Extra")
	}
	if parseFloat(data["gaslimit"]) != "3141592" {
		t.Fatal("Wrong Gas limit")
	}
	if parseFloat(data["gasused"]) != "231000" {
		t.Fatal("Wrong Gas used")
	}
	if data["mixdigest"] != "0x5b10f4a08a6c209d426f6158bd24b574f4f7b7aa0099c67c14a1f693b4dd04d0" {
		t.Fatal("Wrong Mix digest")
	}
	if data["nonce"] != "0xf491f46b60fe04b3" {
		t.Fatal("Wrong nonce")
	}
	if parseFloat(data["number"]) != "999999" {
		t.Fatal("Wrong block number")
	}
	if parseMapElement(data["parent"]) != "z43AaGF6wP6uoLFEauru5oLK5JS5MGfNuGDK1xWEpQK4BqkJkL3" {
		t.Fatal("Wrong Parent cid")
	}
	if parseMapElement(data["receipts"]) != "z44vkPhhDSTXPAswvC1rdDunzkgZ7FgAAnhGQtNDNDk9m9N2BZA" {
		t.Fatal("Wrong Receipt root cid")
	}
	if parseMapElement(data["root"]) != "z45oqTSAZvPiiPV8hMZDH5fi4NkaAkMYTJC6PmaeWBmYUpbMpoh" {
		t.Fatal("Wrong root hash cid")
	}
	if parseFloat(data["time"]) != "1455404037" {
		t.Fatal("Wrong Time")
	}
	if parseMapElement(data["tx"]) != "z443fKyHHMwVy13VXtD4fdRcUXSqkr79Q5E8hcmEravVBq3Dc51" {
		t.Fatal("Wrong Tx root cid")
	}
	if parseMapElement(data["uncles"]) != "z43c7o74hjCAqnyneWetkyXU2i5KuGQLbYfVWZMvJMG4VTYABtz" {
		t.Fatal("Wrong Uncle hash cid")
	}
}

func TestTauBlockLinks(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	links := tauBlock.Links()
	if links[0].Cid.String() != "z43AaGF6wP6uoLFEauru5oLK5JS5MGfNuGDK1xWEpQK4BqkJkL3" {
		t.Fatal("Wrong cid for parent link")
	}
	if links[1].Cid.String() != "z44vkPhhDSTXPAswvC1rdDunzkgZ7FgAAnhGQtNDNDk9m9N2BZA" {
		t.Fatal("Wrong cid for receipt root link")
	}
	if links[2].Cid.String() != "z45oqTSAZvPiiPV8hMZDH5fi4NkaAkMYTJC6PmaeWBmYUpbMpoh" {
		t.Fatal("Wrong cid for state root link")
	}
	if links[3].Cid.String() != "z443fKyHHMwVy13VXtD4fdRcUXSqkr79Q5E8hcmEravVBq3Dc51" {
		t.Fatal("Wrong cid for transaction root link")
	}
	if links[4].Cid.String() != "z43c7o74hjCAqnyneWetkyXU2i5KuGQLbYfVWZMvJMG4VTYABtz" {
		t.Fatal("Wrong cid for uncle root link")
	}
}

func TestTauBlockResolveEmptyPath(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	obj, rest, err := tauBlock.Resolve([]string{})
	checkError(err, t)

	if tauBlock != obj.(*TauBlock) {
		t.Fatal("Should have returned the same tau-block object")
	}

	if len(rest) != 0 {
		t.Fatal("Wrong rest of the path returned")
	}
}

func TestTauBlockResolveNoSuchLink(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	_, _, err := tauBlock.Resolve([]string{"wewonthavtauisfieldever"})
	if err == nil {
		t.Fatal("Should have failed with unknown field")
	}

	if err.Error() != "no such link" {
		t.Fatal("Wrong error message")
	}
}

func TestTauBlockResolveBloom(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	obj, rest, err := tauBlock.Resolve([]string{"bloom"})
	checkError(err, t)

	// The marshaler of types.Bloom should output it as 0x
	bloomInText := fmt.Sprintf("%x", obj.(types.Bloom))
	if bloomInText[:10] != "0000000000" {
		t.Fatal("Wrong Bloom")
	}

	if len(rest) != 0 {
		t.Fatal("Wrong rest of the path returned")
	}
}

func TestTauBlockResolveBloomExtraPathElements(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	obj, rest, err := tauBlock.Resolve([]string{"bloom", "unexpected", "extra", "elements"})
	if obj != nil {
		t.Fatal("Returned obj should be nil")
	}

	if rest != nil {
		t.Fatal("Returned rest should be nil")
	}

	if err.Error() != "unexpected path elements past bloom" {
		t.Fatal("Wrong error")
	}
}

func TestTauBlockResolveNonLinkFields(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	testCases := map[string][]string{
		"coinbase":   []string{"%x", "52bc44d5378309ee2abf1539bf71de1b7d7be3b5"},
		"difficulty": []string{"%s", "12555463106190"},
		"extra":      []string{"%s", "0xd783010303844765746887676f312e342e32856c696e7578"},
		"gaslimit":   []string{"%s", "3141592"},
		"gasused":    []string{"%s", "231000"},
		"mixdigest":  []string{"%x", "5b10f4a08a6c209d426f6158bd24b574f4f7b7aa0099c67c14a1f693b4dd04d0"},
		"nonce":      []string{"%x", "f491f46b60fe04b3"},
		"number":     []string{"%s", "999999"},
		"time":       []string{"%s", "1455404037"},
	}

	for field, value := range testCases {
		obj, rest, err := tauBlock.Resolve([]string{field})
		checkError(err, t)

		format := value[0]
		result := value[1]
		if fmt.Sprintf(format, obj) != result {
			t.Fatalf("Wrong %v", field)
		}

		if len(rest) != 0 {
			t.Fatal("Wrong rest of the path returned")
		}
	}
}

func TestTauBlockResolveNonLinkFieldsExtraPathElements(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	testCases := []string{
		"coinbase",
		"difficulty",
		"extra",
		"gaslimit",
		"gasused",
		"mixdigest",
		"nonce",
		"number",
		"time",
	}

	for _, field := range testCases {
		obj, rest, err := tauBlock.Resolve([]string{field, "unexpected", "extra", "elements"})
		if obj != nil {
			t.Fatal("Returned obj should be nil")
		}

		if rest != nil {
			t.Fatal("Returned rest should be nil")
		}

		if err.Error() != "unexpected path elements past "+field {
			t.Fatal("Wrong error")
		}

	}
}

func TestTauBlockResolveLinkFields(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	testCases := map[string]string{
		"parent":   "z43AaGF6wP6uoLFEauru5oLK5JS5MGfNuGDK1xWEpQK4BqkJkL3",
		"receipts": "z44vkPhhDSTXPAswvC1rdDunzkgZ7FgAAnhGQtNDNDk9m9N2BZA",
		"root":     "z45oqTSAZvPiiPV8hMZDH5fi4NkaAkMYTJC6PmaeWBmYUpbMpoh",
		"tx":       "z443fKyHHMwVy13VXtD4fdRcUXSqkr79Q5E8hcmEravVBq3Dc51",
		"uncles":   "z43c7o74hjCAqnyneWetkyXU2i5KuGQLbYfVWZMvJMG4VTYABtz",
	}

	for field, result := range testCases {
		obj, rest, err := tauBlock.Resolve([]string{field, "anything", "goes", "here"})
		checkError(err, t)

		lnk, ok := obj.(*node.Link)
		if !ok {
			t.Fatal("Returned object is not a link")
		}

		if lnk.Cid.String() != result {
			t.Fatalf("Wrong %s", field)
		}

		for i, p := range []string{"anything", "goes", "here"} {
			if rest[i] != p {
				t.Fatal("Wrong rest of the path returned")
			}
		}
	}
}

func TestTauBlockTreeBadParams(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	tree := tauBlock.Tree("non-empty-string", 0)
	if tree != nil {
		t.Fatal("Expected nil to be returned")
	}

	tree = tauBlock.Tree("non-empty-string", 1)
	if tree != nil {
		t.Fatal("Expected nil to be returned")
	}

	tree = tauBlock.Tree("", 0)
	if tree != nil {
		t.Fatal("Expected nil to be returned")
	}
}

func TestEThBlockTree(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	tree := tauBlock.Tree("", 1)
	lookupElements := map[string]interface{}{
		"bloom":      nil,
		"coinbase":   nil,
		"difficulty": nil,
		"extra":      nil,
		"gaslimit":   nil,
		"gasused":    nil,
		"mixdigest":  nil,
		"nonce":      nil,
		"number":     nil,
		"parent":     nil,
		"receipts":   nil,
		"root":       nil,
		"time":       nil,
		"tx":         nil,
		"uncles":     nil,
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

/*
  The two functions above: TestTauBlockResolveNonLinkFields and
  TestTauBlockResolveLinkFields did all the heavy lifting. Then, we will
  just test two use cases.
*/
func TestTauBlockResolveLinksBadLink(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	obj, rest, err := tauBlock.ResolveLink([]string{"supercalifragilist"})
	if obj != nil {
		t.Fatalf("Expected obj to be nil")
	}
	if rest != nil {
		t.Fatal("Expected rest to be nil")
	}
	if err.Error() != "no such link" {
		t.Fatal("Expected error")
	}
}

func TestTauBlockResolveLinksGoodLink(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	obj, rest, err := tauBlock.ResolveLink([]string{"tx", "0", "0", "0"})
	if obj == nil {
		t.Fatalf("Expected valid *node.Link obj to be returned")
	}

	if rest == nil {
		t.Fatal("Expected rest to be returned")
	}
	for i, p := range []string{"0", "0", "0"} {
		if rest[i] != p {
			t.Fatal("Wrong rest of the path returned")
		}
	}

	if err != nil {
		t.Fatal("Non error expected")
	}
}

/*
  These functions below should go away
  We are working on test coverage anyways...
*/
func TestTauBlockCopy(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic")
		}
		if r != "dont use this yet" {
			t.Fatal("Expected panic message 'dont use this yet'")
		}
	}()

	_ = tauBlock.Copy()
}

func TestTauBlockStat(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	obj, err := tauBlock.Stat()
	if obj == nil {
		t.Fatal("Expected a not null object node.NodeStat")
	}

	if err != nil {
		t.Fatal("Expected a nil error")
	}
}

func TestTauBlockSize(t *testing.T) {
	tauBlock := prepareDecodedTauBlock("test_data/tau-block-header-rlp-999999", t)

	size, err := tauBlock.Size()
	if size != 0 {
		t.Fatal("Expected a size equal to 0")
	}

	if err != nil {
		t.Fatal("Expected a nil error")
	}
}

/*
  AUXILIARS
*/

// checkError makes 3 lines into 1.
func checkError(err error, t *testing.T) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		t.Fatalf("[%v:%v] %v", fn, line, err)
	}
}

// parseFloat is a convenience function to test json output
func parseFloat(v interface{}) string {
	return strconv.FormatFloat(v.(float64), 'f', 0, 64)
}

// parseMapElement is a convenience function to tets json output
func parseMapElement(v interface{}) string {
	return v.(map[string]interface{})["/"].(string)
}

// prepareStoredTauBlock reads the block from a file source to get its rawdata
// and computes its cid, for then, feeding it into a new IPLD block function.
// So we can pretend that we got this block from the datastore
func prepareStoredTauBlock(filepath string, t *testing.T) *block.BasicBlock {
	// Prepare the "fetched block". This one is supposed to be in the datastore
	// and given away by github.com/ipfs/go-ipfs/merkledag
	fi, err := os.Open(filepath)
	checkError(err, t)

	b, err := ioutil.ReadAll(fi)
	checkError(err, t)

	c := rawdataToCid(MTauBlock, b)
	// It's good to clarify that this one below is an IPLD block
	storedTauBlock, err := block.NewBlockWithCid(b, c)
	checkError(err, t)

	return storedTauBlock
}

// prepareDecodedTauBlock is more complex than function above, as it stores a
// basic block and RLP-decodes it
func prepareDecodedTauBlock(filepath string, t *testing.T) *TauBlock {
	// Get the block from the datastore and decode it.
	storedTauBlock := prepareStoredTauBlock("test_data/tau-block-header-rlp-999999", t)
	tauBlock, err := DecodeTauBlock(storedTauBlock.Cid(), storedTauBlock.RawData())
	checkError(err, t)

	return tauBlock
}

// testTauBlockFields checks the fields of TauBlock one by one.
func testTauBlockFields(tauBlock *TauBlock, t *testing.T) {
	// Was the cid calculated?
	if tauBlock.Cid().String() != "z43AaGF4uHSY4waU68L3DLUKHZP7yfZoo6QbLmid5HomZ4WtbWw" {
		t.Fatal("Wrong cid")
	}

	// Do we have the rawdata available?
	if fmt.Sprintf("%x", tauBlock.RawData()[:10]) != "f90218a0d33c9dde9fff" {
		t.Fatal("Wrong Rawdata")
	}

	// Proper Fields of types.Header
	if fmt.Sprintf("%x", tauBlock.ParentHash) != "d33c9dde9fff0ebaa6e71e8b26d2bda15ccf111c7af1b633698ac847667f0fb4" {
		t.Fatal("Wrong ParentHash")
	}
	if fmt.Sprintf("%x", tauBlock.UncleHash) != "1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347" {
		t.Fatal("Wrong UncleHash field")
	}
	if fmt.Sprintf("%x", tauBlock.Coinbase) != "52bc44d5378309ee2abf1539bf71de1b7d7be3b5" {
		t.Fatal("Wrong Coinbase")
	}
	if fmt.Sprintf("%x", tauBlock.Root) != "ed98aa4b5b19c82fb35364f08508ae0a6dec665fa57663dca94c5d70554cde10" {
		t.Fatal("Wrong Root")
	}
	if fmt.Sprintf("%x", tauBlock.TxHash) != "447cbd8c48f498a6912b10831cdff59c7fbfcbbe735ca92883d4fa06dcd7ae54" {
		t.Fatal("Wrong TxHash")
	}
	if fmt.Sprintf("%x", tauBlock.ReceiptHash) != "7fa0f6ca2a01823208d80801edad37e3e3a003b55c89319b45eb1f97862ad229" {
		t.Fatal("Wrong ReceiptHash")
	}
	if len(tauBlock.Bloom) != 256 {
		t.Fatal("Wrong Bloom Length")
	}
	if fmt.Sprintf("%x", tauBlock.Bloom[71:76]) != "0000000000" { // You wouldn't want me to print out the whole bloom field?
		t.Fatal("Wrong Bloom")
	}
	if tauBlock.Difficulty.String() != "12555463106190" {
		t.Fatal("Wrong Difficulty")
	}
	if tauBlock.Number.String() != "999999" {
		t.Fatal("Wrong Block Number")
	}
	if tauBlock.GasLimit.String() != "3141592" {
		t.Fatal("Wrong Gas Limit")
	}
	if tauBlock.GasUsed.String() != "231000" {
		t.Fatal("Wrong Gas Used")
	}
	if tauBlock.Time.String() != "1455404037" {
		t.Fatal("Wrong Time")
	}
	if fmt.Sprintf("%x", tauBlock.Extra) != "d783010303844765746887676f312e342e32856c696e7578" {
		t.Fatal("Wrong Extra")
	}
	if fmt.Sprintf("%x", tauBlock.Nonce) != "f491f46b60fe04b3" {
		t.Fatal("Wrong Nonce")
	}
	if fmt.Sprintf("%x", tauBlock.MixDigest) != "5b10f4a08a6c209d426f6158bd24b574f4f7b7aa0099c67c14a1f693b4dd04d0" {
		t.Fatal("Wrong MixDigest")
	}
}
