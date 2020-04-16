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
	"io"
	"math/big"
	"sync/atomic"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common/hexutil"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/crypto"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/rlp"
	"golang.org/x/crypto/sha3"
)

//go:generate gencodec -type NewMessageTxData -field-override NewMessageTxDataMarshaling -out new_message_tx_json.go
type NewMessageTx struct {
	tx NewMessageTxData

	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type NewMessageTxData struct {
	Version   OneByte         `json:"version"     gencodec:"required"`
	Option    OneByte         `json:"option"      gencodec:"required"`
	ChainID   Byte32s         `json:"chainid"     gencodec:"required"`
	Nonce     uint64          `json:"nounce"      gencodec:"required"`
	TimeStamp uint32          `json:"timestamp"   gencodec:"required"`
	Fee       *big.Int        `json:"fee"         gencodec:"required"`
	V         *big.Int        `json:"v"           gencodec:"required"`
	R         *big.Int        `json:"r"           gencodec:"required"`
	S         *big.Int        `json:"s"           gencodec:"required"`
	Sender    *common.Address `json:"sender"        rlp:"required"`

	Referid *common.Hash `json:"referid"       rlp:"-"`
	Title   Byte144s     `json:"title"         gencodec:"required"`
	Content Byte32s      `json:"contentcid"    gencodec:"required"`
}

type NewMessageTxDataMarshaling struct {
	Version   hexutil.Bytes
	Option    hexutil.Bytes
	ChainID   hexutil.Bytes
	Nonce     hexutil.Uint64
	TimeStamp hexutil.Uint32
	Fee       *hexutil.Big
	V         *hexutil.Big
	R         *hexutil.Big
	S         *hexutil.Big

	Title   hexutil.Bytes
	Content hexutil.Bytes
}

func NewMessageTransaction(version OneByte, option OneByte, chainid Byte32s, nonce uint64, timestamp uint32, fee *big.Int, sender common.Address, referid common.Hash, title Byte144s, content Byte32s) *NewMessageTx {
	return newMessageTransaction(version, option, chainid, nonce, timestamp, fee, &sender, referid, title, content)
}

func newMessageTransaction(version OneByte, option OneByte, chainid Byte32s, nonce uint64, timestamp uint32, fee *big.Int, sender *common.Address, referid common.Hash, title Byte144s, content Byte32s) *NewMessageTx {
	d := NewMessageTxData{
		Version:   version,
		Option:    option,
		ChainID:   chainid,
		Nonce:     nonce,
		TimeStamp: timestamp,
		Fee:       fee,
		V:         new(big.Int),
		R:         new(big.Int),
		S:         new(big.Int),
		Sender:    sender,

		Referid: &referid,
		Title:   title,
		Content: content,
	}

	return &NewMessageTx{tx: d}
}

func (mtx *NewMessageTx) ChainId() Byte32s {
	return mtx.tx.ChainID
}

func (mtx *NewMessageTx) Protected() bool {
	return true
}

func (mtx *NewMessageTx) isProtectedV(V *big.Int) bool {
	v := V.Uint64()
	if v == 27 || v == 28 {
		return false
	}

	return true
}

func (mtx *NewMessageTx) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &mtx.tx)
}

func (mtx *NewMessageTx) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&mtx.tx)
	if err == nil {
		mtx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}

	return err
}

func (mtx *NewMessageTx) MarshalJSON() ([]byte, error) {
	data := mtx.tx
	return data.MarshalJSON()
}

func (mtx *NewMessageTx) UnmarshalJSON(input []byte) error {
	var dec NewMessageTxData
	if err := dec.UnmarshalJSON(input); err != nil {
		return err
	}

	withSignature := dec.V.Sign() != 0 || dec.R.Sign() != 0 || dec.S.Sign() != 0
	if withSignature {
		var V byte
		if mtx.isProtectedV(dec.V) {
			chainID := deriveChainId(dec.V).Uint64()
			V = byte(dec.V.Uint64() - 35 - 2*chainID)
		} else {
			V = byte(dec.V.Uint64() - 27)
		}
		if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
			return ErrInvalidSig
		}
	}

	*mtx = NewMessageTx{tx: dec}
	return nil
}

func (mtx *NewMessageTx) Fee() *big.Int {
	big := new(big.Int)
	return big.Set(mtx.tx.Fee)
}

func (mtx *NewMessageTx) Value() *big.Int     { return new(big.Int) }
func (mtx *NewMessageTx) Nonce() uint64       { return mtx.tx.Nonce }
func (mtx *NewMessageTx) CheckNonce() bool    { return true }
func (mtx *NewMessageTx) To() *common.Address { return &common.Address{} }

func (mtx *NewMessageTx) Hash() (h common.Hash) {
	if hash := mtx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, mtx)
	hw.Sum(h[:0])

	mtx.hash.Store(h)
	return h
}

func (mtx *NewMessageTx) Size() common.StorageSize {
	if size := mtx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &mtx.tx)
	mtx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

func (mtx *NewMessageTx) AsMessage(s Signer) (Message, error) {
	msg := Message{
		from:       *mtx.tx.Sender,
		to:         nil,
		nonce:      mtx.tx.Nonce,
		amount:     nil,
		fee:        mtx.tx.Fee,
		checkNonce: true,
	}

	var err error
	//msg.from, err = Sender(s, ttx)
	return msg, err
}

func (mtx *NewMessageTx) WithSignature(singer Signer, sig []byte) (bool, error) {
	V, R, S, err := singer.SignatureValues(sig)
	if err != nil {
		return false, err
	}
	//contain signature in ttx itself
	//fill field of versioned signature in ttx
	mtx.tx.V = V
	mtx.tx.R = R
	mtx.tx.S = S
	return true, nil
}

func (mtx *NewMessageTx) Cost() *big.Int {
	fee := new(big.Int)
	fee.Set(mtx.tx.Fee)
	return fee
}

func (mtx *NewMessageTx) RawSignatureValues() (v, r, s *big.Int) {
	return mtx.tx.V, mtx.tx.R, mtx.tx.S
}

func (mtx *NewMessageTx) GetFrom() atomic.Value {
	return mtx.from
}

func (mtx *NewMessageTx) GetSigV() *big.Int {
	if mtx.tx.V != nil {
		return mtx.tx.V
	}
	return nil
}

func (mtx *NewMessageTx) GetSigR() *big.Int {
	if mtx.tx.R != nil {
		return mtx.tx.R
	}
	return nil
}

func (mtx *NewMessageTx) GetSigS() *big.Int {
	if mtx.tx.S != nil {
		return mtx.tx.S
	}
	return nil
}

func (mtx *NewMessageTx) GetNounce() uint64 {
	return mtx.tx.Nonce
}

func (mtx *NewMessageTx) GetFee() uint64 {
	return mtx.tx.Fee.Uint64()
}

func (mtx *NewMessageTx) GetReceiver() common.Address {
	return common.Address{}
}

func (mtx *NewMessageTx) GetAmount() big.Int {
	return big.Int{}
}
