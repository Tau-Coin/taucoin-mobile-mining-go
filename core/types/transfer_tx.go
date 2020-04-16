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
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
// maintained by likeopen
package types

import (
	"container/heap"
	"io"
	"math/big"
	"sync/atomic"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common/hexutil"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/crypto"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/rlp"
	"golang.org/x/crypto/sha3"
)

//go:generate gencodec -type TransferTxData -field-override TransferTxDataMarshaling -out transfer_tx_json.go
type TransferTx struct {
	tx TransferTxData

	//cache
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}
type Byte5s []byte

type TransferTxData struct {
	Version   OneByte `json:"version"     gencodec:"required"`
	Option    OneByte `json:"option"      gencodec:"required"`
	ChainID   Byte32s `json:"chainid"     gencodec:"required"`
	Nonce     uint64  `json:"nounce"      gencodec:"required"`
	TimeStamp uint32  `json:"timestamp"   gencodec:"required"`
	//Fee       OneByte         `json:"fee"         gencodec:"required"`
	Fee    *big.Int        `json:"fee"         gencodec:"required"`
	V      *big.Int        `json:"v"           gencodec:"required"`
	R      *big.Int        `json:"r"           gencodec:"required"`
	S      *big.Int        `json:"s"           gencodec:"required"`
	Sender *common.Address `json:"sender"        rlp:"required"`

	Receiver *common.Address `json:"receiver"        rlp:"required"`
	//Amount   Byte5s          `json:"amount"       gencodec:"required"`
	Amount *big.Int `json:"value"    gencodec:"required"`
}

type TransferTxDataMarshaling struct {
	Version   hexutil.Bytes
	Option    hexutil.Bytes
	ChainID   hexutil.Bytes
	Nonce     hexutil.Uint64
	TimeStamp hexutil.Uint32
	//Fee       hexutil.Bytes
	Fee *hexutil.Big
	V   *hexutil.Big
	R   *hexutil.Big
	S   *hexutil.Big

	//Amount hexutil.Bytes
	Amount *hexutil.Big
}

func NewTransferTransaction(version OneByte, option OneByte, chainid Byte32s, nounce uint64, timestamp uint32, fee *big.Int, sender common.Address, receiver common.Address, amount *big.Int) *TransferTx {
	return newTransferTransaction(version, option, chainid, nounce, timestamp, fee, &sender, &receiver, amount)
}

func newTransferTransaction(version OneByte, option OneByte, chainid Byte32s, nounce uint64, timestamp uint32, fee *big.Int, sender *common.Address, receiver *common.Address, amount *big.Int) *TransferTx {
	d := TransferTxData{
		Version:   version,
		Option:    option,
		ChainID:   chainid,
		Nonce:     nounce,
		TimeStamp: timestamp,
		Fee:       fee,
		V:         new(big.Int),
		R:         new(big.Int),
		S:         new(big.Int),
		Sender:    sender,
		Receiver:  receiver,
		//Amount:    amount,
		Amount: new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	return &TransferTx{tx: d}
}

func (ttx *TransferTx) ChainId() Byte32s {
	return ttx.tx.ChainID
}

func (ttx *TransferTx) Protected() bool {
	return true
}

func (ttx *TransferTx) isProtectedV(V *big.Int) bool {
	v := V.Uint64()
	if v == 27 || v == 28 {
		return false
	}

	return true
}

func (ttx *TransferTx) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &ttx.tx)
}

func (ttx *TransferTx) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&ttx.tx)
	if err == nil {
		ttx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}

	return err
}

func (ttx *TransferTx) MarshalJSON() ([]byte, error) {
	data := ttx.tx
	return data.MarshalJSON()
}

func (ttx *TransferTx) UnmarshalJSON(input []byte) error {
	var dec TransferTxData
	if err := dec.UnmarshalJSON(input); err != nil {
		return err
	}

	withSignature := dec.V.Sign() != 0 || dec.R.Sign() != 0 || dec.S.Sign() != 0
	if withSignature {
		var V byte
		if ttx.isProtectedV(dec.V) {
			chainID := deriveChainId(dec.V).Uint64()
			V = byte(dec.V.Uint64() - 35 - 2*chainID)
		} else {
			V = byte(dec.V.Uint64() - 27)
		}
		if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
			return ErrInvalidSig
		}
	}

	*ttx = TransferTx{tx: dec}
	return nil
}

func (ttx *TransferTx) Fee() *big.Int {
	big := new(big.Int)
	return big.Set(ttx.tx.Fee)
}

//func (ttx *TransferTx) Value() Byte5s  { return ttx.tx.Amount }
func (ttx *TransferTx) Value() *big.Int     { return new(big.Int).Set(ttx.tx.Amount) }
func (ttx *TransferTx) Nonce() uint64       { return ttx.tx.Nonce }
func (ttx *TransferTx) CheckNonce() bool    { return true }
func (ttx *TransferTx) To() *common.Address { return ttx.tx.Receiver }

func (ttx *TransferTx) Hash() (h common.Hash) {
	if hash := ttx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, ttx)
	hw.Sum(h[:0])

	ttx.hash.Store(h)
	return h
}

func (ttx *TransferTx) Size() common.StorageSize {
	if size := ttx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &ttx.tx)
	ttx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

func (ttx *TransferTx) AsMessage(s Signer) (Message, error) {
	msg := Message{
		from:       *ttx.tx.Sender,
		to:         ttx.tx.Receiver,
		nonce:      ttx.tx.Nonce,
		amount:     ttx.tx.Amount,
		fee:        ttx.tx.Fee,
		checkNonce: true,
	}

	var err error
	//msg.from, err = Sender(s, ttx)
	return msg, err
}

func (ttx *TransferTx) WithSignature(singer Signer, sig []byte) (bool, error) {
	V, R, S, err := singer.SignatureValues(sig)
	if err != nil {
		return false, err
	}
	//contain signature in ttx itself
	//fill field of versioned signature in ttx
	ttx.tx.V = V
	ttx.tx.R = R
	ttx.tx.S = S
	return true, nil
}

func (ttx *TransferTx) Cost() *big.Int {
	cost := ttx.tx.Amount
	fee := new(big.Int)
	fee.Set(ttx.tx.Fee)
	return cost.Add(cost, fee)
}

func (ttx *TransferTx) RawSignatureValues() (v, r, s *big.Int) {
	return ttx.tx.V, ttx.tx.R, ttx.tx.S
}

func (ttx *TransferTx) GetFrom() atomic.Value {
	return ttx.from
}

func (ttx *TransferTx) GetSigV() *big.Int {
	if ttx.tx.V != nil {
		return ttx.tx.V
	}
	return nil
}

func (ttx *TransferTx) GetSigR() *big.Int {
	if ttx.tx.R != nil {
		return ttx.tx.R
	}
	return nil
}

func (ttx *TransferTx) GetSigS() *big.Int {
	if ttx.tx.S != nil {
		return ttx.tx.S
	}
	return nil
}

func (ttx *TransferTx) GetNounce() uint64 {
	return ttx.tx.Nonce
}
func (ttx *TransferTx) GetFee() uint64 {
	return ttx.tx.Fee.Uint64()
}
func (ttx *TransferTx) GetReceiver() common.Address {
	return *(ttx.tx.Sender)
}
func (ttx *TransferTx) GetAmount() big.Int {
	return *(ttx.tx.Amount)
}

type TransferTxs []*TransferTx

func (s TransferTxs) Len() int { return len(s) }

func (s TransferTxs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s TransferTxs) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

func (s TransferTxs) TxDifference(a, b TransferTxs) TransferTxs {
	keep := make(TransferTxs, 0, len(a))

	remove := make(map[common.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}

type TransferTxByNonce TransferTxs

func (s TransferTxByNonce) Len() int { return len(s) }
func (s TransferTxByNonce) Less(i, j int) bool {
	return s[i].tx.Nonce < s[j].tx.Nonce
}
func (s TransferTxByNonce) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *TransferTxByNonce) Push(x interface{}) {
	*s = append(*s, x.(*TransferTx))
}

func (s *TransferTxByNonce) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type TransferTxByFee TransferTxs

func (s TransferTxByFee) Len() int { return len(s) }
func (s TransferTxByFee) Less(i, j int) bool {
	fee1 := new(big.Int).Set(s[i].tx.Fee)
	fee2 := new(big.Int).Set(s[j].tx.Fee)
	return fee1.Cmp(fee2) > 0
}
func (s TransferTxByFee) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *TransferTxByFee) Push(x interface{}) {
	*s = append(*s, x.(*TransferTx))
}

func (s *TransferTxByFee) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type TransferTxByFeeAndNonce struct {
	txs    map[common.Address]TransferTxs
	heads  TransferTxByFee
	signer Signer
}

//watch out TransferTxs is sorted by Nonce first
func NewTransferTxByFeeAndNonce(signer Signer, txs map[common.Address]TransferTxs) *TransferTxByFeeAndNonce {
	heads := make(TransferTxByFee, 0, len(txs))
	for _, accTxs := range txs {
		heads = append(heads, accTxs[0])
		//to make sure a list txs from txs is from same account
		//to do to complete singer
	}
	heap.Init(&heads)
	return &TransferTxByFeeAndNonce{
		txs:   txs,
		heads: heads,
		//This singer need to make adaption mpdify abount unique tx
		signer: signer,
	}
}

func (t *TransferTxByFeeAndNonce) Peek() *TransferTx {
	if len(t.heads) == 0 {
		return nil
	}

	return t.heads[0]
}

func (t *TransferTxByFeeAndNonce) Shift() {
	//if Account x contains other txs sorted by Nonce, the others should
	//come up with new fee sorting because of some fee element changing.
	acc := t.heads[0].tx.Sender
	if txs, ok := t.txs[*acc]; ok && len(txs) > 0 {
		t.heads[0], t.txs[*acc] = txs[0], txs[1:]
		heap.Fix(&t.heads, 0)
	} else {
		heap.Pop(&t.heads)
	}
}

func (t *TransferTxByFeeAndNonce) Pop() {
	heap.Pop(&t.heads)
}

//todo
//these messages need to define to adapt new ipfs system.
type TauTxMessage struct {
	from   common.Address
	to     *common.Address
	nonce  uint64
	amount *big.Int
	fee    *big.Int
}

func NewTauTxMessage(from common.Address, to *common.Address, nonce uint64, amount *big.Int, fee *big.Int) TauTxMessage {
	return TauTxMessage{
		from:   from,
		to:     to,
		nonce:  nonce,
		amount: amount,
		fee:    fee,
	}
}

func (m TauTxMessage) From() common.Address { return m.from }
func (m TauTxMessage) To() *common.Address  { return m.to }
func (m TauTxMessage) Nonce() uint64        { return m.nonce }
func (m TauTxMessage) Value() *big.Int      { return m.amount }
func (m TauTxMessage) Fee() *big.Int        { return m.fee }
