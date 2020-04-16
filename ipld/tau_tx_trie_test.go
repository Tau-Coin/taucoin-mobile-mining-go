package ipldtau

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	block "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	node "github.com/ipfs/go-ipld-format"
)

/*
  TauBlock
*/

func TestTxTriesInBlockBodyJSONParsing(t *testing.T) {
	// HINT: 306 txs
	// cat test_data/tau-block-body-json-4139497 | jsontool | grep transactionIndex | wc -l
	// or, https://tauerscan.io/block/4139497
	fi, err := os.Open("test_data/tau-block-body-json-4139497")
	checkError(err, t)

	_, _, output, err := FromBlockJSON(fi)
	checkError(err, t)

	if len(output) != 331 {
		t.Fatal("Wrong number of obtained tx trie nodes")
	}
}

/*
  OUTPUT
*/

func TestTxTrieDecodeExtension(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieExtension(t)

	if tauTxTrie.nodeKind != "extension" {
		t.Fatal("Wrong nodeKind")
	}

	if len(tauTxTrie.elements) != 2 {
		t.Fatal("Wrong number of elements for an extension node")
	}

	if fmt.Sprintf("%x", tauTxTrie.elements[0].([]byte)) != "0001" {
		t.Fatal("Wrong key")
	}

	if tauTxTrie.elements[1].(*cid.Cid).String() !=
		"z443fKyJaFfaE7Hsozvv7HGEHqNWPEhkNgzgnXjVKdxqCE74PgF" {
		t.Fatal("Wrong Value")
	}
}

func TestTxTrieDecodeLeaf(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieLeaf(t)

	if tauTxTrie.nodeKind != "leaf" {
		t.Fatal("Wrong nodeKind")
	}

	if len(tauTxTrie.elements) != 2 {
		t.Fatal("Wrong number of elements for a leaf node")
	}

	if fmt.Sprintf("%x", tauTxTrie.elements[0].([]byte)) != "" {
		t.Fatal("Wrong key")
	}

	if _, ok := tauTxTrie.elements[1].(*TauTx); !ok {
		t.Fatal("Wrong Type. Element should be a transaction")
	}

	if tauTxTrie.elements[1].(*TauTx).String() !=
		"<TauereumTx z44VCrqXZu4u7rxeXCTgn6epNgUrum8Nrr7bCFToDBr3EWwe2N6>" {
		t.Fatal("Wrong element, supposed to be a transaction")
	}
}

func TestTxTrieDecodeBranch(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieBranch(t)

	if tauTxTrie.nodeKind != "branch" {
		t.Fatal("Wrong nodeKind")
	}

	if len(tauTxTrie.elements) != 17 {
		t.Fatal("Wrong number of elements for a branch node")
	}

	for i, element := range tauTxTrie.elements {
		switch {
		case i < 9:
			if _, ok := element.(*cid.Cid); !ok {
				t.Fatal("Expected element to be a cid")
			}
			continue
		default:
			if element != nil {
				t.Fatal("Expected element to be a nil")
			}
		}
	}
}

/*
  Block INTERFACE
*/

func TestTauTxTrieBlockElements(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieExtension(t)

	if fmt.Sprintf("%x", tauTxTrie.RawData())[:10] != "e4820001a0" {
		t.Fatal("Wrong Data")
	}

	if tauTxTrie.Cid().String() !=
		"z443fKyR2PNJ3gNLTrPEmkHJh4YJ2mNMU9QX4HuBFNfBGnkb444" {
		t.Fatal("Wrong Cid")
	}
}

func TestTauTxTrieString(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieExtension(t)

	if tauTxTrie.String() != "<TauereumTxTrie z443fKyR2PNJ3gNLTrPEmkHJh4YJ2mNMU9QX4HuBFNfBGnkb444>" {
		t.Fatalf("Wrong String()")
	}
}

func TestTauTxTrieLoggable(t *testing.T) {

	tauTxTrie := prepareDecodedTauTxTrieExtension(t)
	l := tauTxTrie.Loggable()
	if _, ok := l["type"]; !ok {
		t.Fatal("Loggable map expected the field 'type'")
	}

	if l["type"] != "tau-tx-trie" {
		t.Fatal("Wrong Loggable 'type' value")
	}
}

/*
  Node INTERFACE
*/

func TestTxTrieResolveExtension(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieExtension(t)

	_ = tauTxTrie
}

func TestTxTrieResolveLeaf(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieLeaf(t)

	_ = tauTxTrie
}

func TestTxTrieResolveBranch(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieBranch(t)

	indexes := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}

	for j, index := range indexes {
		obj, rest, err := tauTxTrie.Resolve([]string{index, "nonce"})

		switch {
		case j < 9:
			_, ok := obj.(*node.Link)
			if !ok {
				t.Fatalf("Returned object is not a link (index: %d)", j)
			}

			if rest[0] != "nonce" {
				t.Fatal("Wrong rest of the path returned")
			}

			if err != nil {
				t.Fatal("Error should be nil")
			}

		default:
			if obj != nil {
				t.Fatalf("Returned object should have been nil")
			}

			if rest != nil {
				t.Fatalf("Rest of the path returned should be nil")
			}

			if err.Error() != "no such link in this branch" {
				t.Fatalf("Wrong error")
			}
		}
	}

	otherSuccessCases := [][]string{
		[]string{"0", "1", "banana"},
		[]string{"1", "banana"},
		[]string{"7bc", "def"},
		[]string{"bc", "def"},
	}

	for i := 0; i < len(otherSuccessCases); i = i + 2 {
		osc := otherSuccessCases[i]
		expectedRest := otherSuccessCases[i+1]

		obj, rest, err := tauTxTrie.Resolve(osc)
		_, ok := obj.(*node.Link)
		if !ok {
			t.Fatalf("Returned object is not a link")
		}

		for j, _ := range expectedRest {
			if rest[j] != expectedRest[j] {
				t.Fatal("Wrong rest of the path returned")
			}
		}

		if err != nil {
			t.Fatal("Error should be nil")
		}

	}
}

func TestTraverseTxTrieWithResolve(t *testing.T) {
	var err error

	txMap := prepareTxTrieMap(t)

	// This is the cid of the tx root at the block 4,139,497
	currentNode := txMap["z443fKyMXhAPwsjNDytVEjCu6EtDdWGUZ5AzFE3dkJtTYnTHAoy"]

	// This is the path we want to traverse
	// the transaction id 256, which is RLP encoded to 820100
	var traversePath []string
	for _, s := range "820100" {
		traversePath = append(traversePath, string(s))
	}
	traversePath = append(traversePath, "value")

	var obj interface{}
	for {
		obj, traversePath, err = currentNode.Resolve(traversePath)
		link, ok := obj.(*node.Link)
		if !ok {
			break
		}
		if err != nil {
			t.Fatal("Error should be nil")
		}

		currentNode = txMap[link.Cid.String()]
		if currentNode == nil {
			t.Fatal("transaction trie node not found in memory map")
		}
	}

	if fmt.Sprintf("%v", obj) != "0xc495a958603400" {
		t.Fatalf("Wrong value %v", obj)
	}
}

func TestTxTrieTreeBadParams(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieBranch(t)

	tree := tauTxTrie.Tree("non-empty-string", 0)
	if tree != nil {
		t.Fatal("Expected nil to be returned")
	}

	tree = tauTxTrie.Tree("non-empty-string", 1)
	if tree != nil {
		t.Fatal("Expected nil to be returned")
	}

	tree = tauTxTrie.Tree("", 0)
	if tree != nil {
		t.Fatal("Expected nil to be returned")
	}
}

func TestTxTrieTreeExtension(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieExtension(t)

	tree := tauTxTrie.Tree("", -1)

	if len(tree) != 1 {
		t.Fatalf("An extension should have one element")
	}

	if tree[0] != "01" {
		t.Fatal("Wrong trie element")
	}
}

func TestTxTrieTreeBranch(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieBranch(t)

	tree := tauTxTrie.Tree("", -1)

	lookupElements := map[string]interface{}{
		"0": nil,
		"1": nil,
		"2": nil,
		"3": nil,
		"4": nil,
		"5": nil,
		"6": nil,
		"7": nil,
		"8": nil,
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

func TestTxTrieLinksBranch(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieBranch(t)

	desiredValues := []string{
		"z443fKyJBiTCxynCqKP1r3BSvJ4nvR4bSEpWFMc7ZJ57L6NJUdH",
		"z443fKySwQ2WU6av9YcvidCRCYcBrcY1FbsJfdxtTTeKbpiZD8k",
		"z443fKyTqcL3923Cwqeun2Lo1qs9MPXNV16KFJBHRs6ghNHaFpf",
		"z443fKyDyheaZ5qTSjSS6XLj6trLWasneACqrkBfwNpnjN2Fuia",
		"z443fKyNhK436C7wMxoiM9NfjcnHpmdWgbW6CKvtA4f9kUnoD9P",
		"z443fKyUZTcKeGxvmCfecLxAF8rHEAzCFNVaTwonX2Atd6BB4CS",
		"z443fKyFbQsGGz5fuym6Gv8hyHErR962okHt1zTNKwXebjwUo3w",
		"z443fKyG5m6cHmnhBfi4qNvRXNmL18w71XZGxJifbtUPyUNfk5Z",
		"z443fKyRJvB8PQEdWTL44qqoo2DeZr8QwkasSAfEcWJ6uDUWyh6",
	}

	links := tauTxTrie.Links()

	for i, v := range desiredValues {
		if links[i].Cid.String() != v {
			t.Fatalf("Wrong cid for link %d", i)
		}
	}
}

/*
  TauTxTrie Functions
*/

func TestTxTrieJSONMarshalExtension(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieExtension(t)

	jsonOutput, err := tauTxTrie.MarshalJSON()
	checkError(err, t)

	var data map[string]interface{}
	err = json.Unmarshal(jsonOutput, &data)
	checkError(err, t)

	if parseMapElement(data["01"]) !=
		"z443fKyJaFfaE7Hsozvv7HGEHqNWPEhkNgzgnXjVKdxqCE74PgF" {
		t.Fatal("Wrong Marshaled Value")
	}

	if data["type"] != "extension" {
		t.Fatal("Expected type to be extension")
	}
}

func TestTxTrieJSONMarshalLeaf(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieLeaf(t)

	jsonOutput, err := tauTxTrie.MarshalJSON()
	checkError(err, t)

	var data map[string]interface{}
	err = json.Unmarshal(jsonOutput, &data)
	checkError(err, t)

	if data["type"] != "leaf" {
		t.Fatal("Expected type to be leaf")
	}

	if fmt.Sprintf("%v", data[""].(map[string]interface{})["nonce"]) !=
		"40243" {
		t.Fatal("Wrong nonce value")
	}
}

func TestTxTrieJSONMarshalBranch(t *testing.T) {
	tauTxTrie := prepareDecodedTauTxTrieBranch(t)

	jsonOutput, err := tauTxTrie.MarshalJSON()
	checkError(err, t)

	var data map[string]interface{}
	err = json.Unmarshal(jsonOutput, &data)
	checkError(err, t)

	desiredValues := map[string]string{
		"0": "z443fKyJBiTCxynCqKP1r3BSvJ4nvR4bSEpWFMc7ZJ57L6NJUdH",
		"1": "z443fKySwQ2WU6av9YcvidCRCYcBrcY1FbsJfdxtTTeKbpiZD8k",
		"2": "z443fKyTqcL3923Cwqeun2Lo1qs9MPXNV16KFJBHRs6ghNHaFpf",
		"3": "z443fKyDyheaZ5qTSjSS6XLj6trLWasneACqrkBfwNpnjN2Fuia",
		"4": "z443fKyNhK436C7wMxoiM9NfjcnHpmdWgbW6CKvtA4f9kUnoD9P",
		"5": "z443fKyUZTcKeGxvmCfecLxAF8rHEAzCFNVaTwonX2Atd6BB4CS",
		"6": "z443fKyFbQsGGz5fuym6Gv8hyHErR962okHt1zTNKwXebjwUo3w",
		"7": "z443fKyG5m6cHmnhBfi4qNvRXNmL18w71XZGxJifbtUPyUNfk5Z",
		"8": "z443fKyRJvB8PQEdWTL44qqoo2DeZr8QwkasSAfEcWJ6uDUWyh6",
	}

	for k, v := range desiredValues {
		if parseMapElement(data[k]) != v {
			t.Fatal("Wrong Marshaled Value")
		}
	}

	for _, v := range []string{"a", "b", "c", "d", "e", "f"} {
		if data[v] != nil {
			t.Fatal("Expected value to be nil")
		}
	}

	if data["type"] != "branch" {
		t.Fatal("Expected type to be branch")
	}
}

/*
  AUXILIARS
*/

// prepareDecodedTauTxTrie simulates an IPLD block available in the datastore,
// checks the source RLP and tests for the absence of errors during the decoding fase.
func prepareDecodedTauTxTrie(branchDataRLP string, t *testing.T) *TauTxTrie {
	b, err := hex.DecodeString(branchDataRLP)
	checkError(err, t)

	c := rawdataToCid(MTauTxTrie, b)

	storedTauTxTrie, err := block.NewBlockWithCid(b, c)
	checkError(err, t)

	tauTxTrie, err := DecodeTauTxTrie(storedTauTxTrie.Cid(), storedTauTxTrie.RawData())
	checkError(err, t)

	return tauTxTrie
}

func prepareDecodedTauTxTrieExtension(t *testing.T) *TauTxTrie {
	extensionDataRLP :=
		"e4820001a057ac34d6471cc3f5c6ab992c4c0fe5ec131d8d9961fe6d5de8e5e367513243b4"
	return prepareDecodedTauTxTrie(extensionDataRLP, t)
}

func prepareDecodedTauTxTrieLeaf(t *testing.T) *TauTxTrie {
	leafDataRLP :=
		"f87220b86ff86d829d3384ee6b280083015f9094e0e6c781b8cba08bc840" +
			"7eac0101b668d1fa6f4987c495a9586034008026a0981b6223c9d3c31971" +
			"6da3cf057da84acf0fef897f4003d8a362d7bda42247dba066be134c4bc4" +
			"32125209b5056ef274b7423bcac7cc398cf60b83aaff7b95469f"
	return prepareDecodedTauTxTrie(leafDataRLP, t)
}

func prepareDecodedTauTxTrieBranch(t *testing.T) *TauTxTrie {
	branchDataRLP :=
		"f90131a051e622bd20e77781a010b9903832e73fd3665e89407ded8c840d8b2db34dd9" +
			"dca0d3f45a40fcad18a6c3d7edbe8e7e92ace9d45e086cbd04a66254b9931375bee1a0" +
			"e15476fc93dc41ef612ac86750dd242d14498c1e48a6ba4fc89fcc501ee7c58ca01363" +
			"826032eeaf1c4540ed2e8e10dc3a34c3fbc4900c7a7c449e69e2ca8a8e1ba094e9d98b" +
			"ebb67807ecd96a6cac608f95a14a07e6a9c06975861e0b86b6c14736a0ec0cfff9d5ab" +
			"a2ac0da8d2c4725bc8253b60f7b6f1c6b4229ea967fcaef319d3a02b652173155b7d9b" +
			"b152ec5d255b82534d3075bcc171a928eba737da9381effaa032a8447e172dc85a1584" +
			"d0f77466ee52a1c00f71caf57e0e1aa01de18a3ca834a0bbc043cc0d03623ba4c7b514" +
			"7d5aca56450b548f797d712d5198f5e8b35f542d8080808080808080"
	return prepareDecodedTauTxTrie(branchDataRLP, t)
}

func prepareTxTrieMap(t *testing.T) map[string]*TauTxTrie {
	fi, err := os.Open("test_data/tau-block-body-json-4139497")
	checkError(err, t)

	_, _, txTrieNodes, err := FromBlockJSON(fi)
	checkError(err, t)

	out := make(map[string]*TauTxTrie)

	for _, txTrieNode := range txTrieNodes {
		decodedNode, err := DecodeTauTxTrie(txTrieNode.Cid(), txTrieNode.RawData())
		checkError(err, t)

		out[txTrieNode.Cid().String()] = decodedNode
	}

	return out
}
