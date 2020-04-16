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

//go:generate gencodec -type PersonalInfoTxData -field-override PersonalInfoTxDataMarshaling -out personal_info_tx_json.go
type PersonalInfoTx struct {
	tx PersonalInfoTxData

	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type PersonalInfoTxData struct {
	Version   OneByte         `json:"version"     gencodec:"required"`
	Option    OneByte         `json:"option"      gencodec:"required"`
	ChainID   Byte32s         `json:"chainid"     gencodec:"required"`
	Nonce     uint64          `json:"nonce"      gencodec:"required"`
	TimeStamp uint32          `json:"timestamp"   gencodec:"required"`
	Fee       *big.Int        `json:"fee"         gencodec:"required"`
	V         *big.Int        `json:"v"           gencodec:"required"`
	R         *big.Int        `json:"r"           gencodec:"required"`
	S         *big.Int        `json:"s"           gencodec:"required"`
	Sender    *common.Address `json:"sender"        rlp:"required"`

	ContactName Byte32s `json:"contactname" gencodec:"required"`
	Name        Byte20s `json:"name"        gencodec:"required"`
	Profile     Byte32s `json:"profile"     gencodec:"required"`
}

type PersonalInfoTxDataMarshaling struct {
	Version   hexutil.Bytes
	Option    hexutil.Bytes
	ChainID   hexutil.Bytes
	Nonce    hexutil.Uint64
	TimeStamp hexutil.Uint32
	Fee       *hexutil.Big
	V         *hexutil.Big
	R         *hexutil.Big
	S         *hexutil.Big

	ContactName hexutil.Bytes
	Name        hexutil.Bytes
	Profile     hexutil.Bytes
}

func NewPersonalInfoTransaction(version OneByte, option OneByte, chainid Byte32s, nonce uint64, timestamp uint32, fee *big.Int, sender common.Address, contactname Byte32s, name Byte20s, profile Byte32s) *PersonalInfoTx {
	return newPersonalInfoTransaction(version, option, chainid, nonce, timestamp, fee, &sender, contactname, name, profile)
}

func newPersonalInfoTransaction(version OneByte, option OneByte, chainid Byte32s, nonce uint64, timestamp uint32, fee *big.Int, sender *common.Address, contactname Byte32s, name Byte20s, profile Byte32s) *PersonalInfoTx {
	d := PersonalInfoTxData{
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

		ContactName: contactname,
		Name:        name,
		Profile:     profile,
	}

	return &PersonalInfoTx{tx: d}
}

func (pitx *PersonalInfoTx) ChainId() Byte32s {
	return pitx.tx.ChainID
}

func (pitx *PersonalInfoTx) Protected() bool {
	return true
}

func (pitx *PersonalInfoTx) isProtectedV(V *big.Int) bool {
	v := V.Uint64()
	if v == 27 || v == 28 {
		return false
	}

	return true
}

func (pitx *PersonalInfoTx) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &pitx.tx)
}

func (pitx *PersonalInfoTx) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&pitx.tx)
	if err == nil {
		pitx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}

	return err
}

func (pitx *PersonalInfoTx) MarshalJSON() ([]byte, error) {
	data := pitx.tx
	return data.MarshalJSON()
}

func (pitx *PersonalInfoTx) UnmarshalJSON(input []byte) error {
	var dec PersonalInfoTxData
	if err := dec.UnmarshalJSON(input); err != nil {
		return err
	}

	withSignature := dec.V.Sign() != 0 || dec.R.Sign() != 0 || dec.S.Sign() != 0
	if withSignature {
		var V byte
		if pitx.isProtectedV(dec.V) {
			chainID := deriveChainId(dec.V).Uint64()
			V = byte(dec.V.Uint64() - 35 - 2*chainID)
		} else {
			V = byte(dec.V.Uint64() - 27)
		}
		if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
			return ErrInvalidSig
		}
	}

	*pitx = PersonalInfoTx{tx: dec}
	return nil
}

func (pitx *PersonalInfoTx) Fee() *big.Int {
	big := new(big.Int)
	return big.Set(pitx.tx.Fee)
}

func (pitx *PersonalInfoTx) Value() *big.Int     { return new(big.Int) }
func (pitx *PersonalInfoTx) Nonce() uint64       { return pitx.tx.Nonce }
func (pitx *PersonalInfoTx) CheckNonce() bool    { return true }
func (pitx *PersonalInfoTx) To() *common.Address { return &common.Address{} }

func (pitx *PersonalInfoTx) Hash() (h common.Hash) {
	if hash := pitx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, pitx)
	hw.Sum(h[:0])

	pitx.hash.Store(h)
	return h
}

func (pitx *PersonalInfoTx) Size() common.StorageSize {
	if size := pitx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &pitx.tx)
	pitx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

func (pitx *PersonalInfoTx) AsMessage(s Signer) (Message, error) {
	msg := Message{
		from:       *pitx.tx.Sender,
		to:         nil,
		nonce:      pitx.tx.Nonce,
		amount:     nil,
		fee:        pitx.tx.Fee,
		checkNonce: true,
	}

	var err error
	//msg.from, err = Sender(s, ttx)
	return msg, err
}

func (pitx *PersonalInfoTx) WithSignature(singer Signer, sig []byte) (bool, error) {
	V, R, S, err := singer.SignatureValues(sig)
	if err != nil {
		return false, err
	}
	//contain signature in ttx itself
	//fill field of versioned signature in ttx
	pitx.tx.V = V
	pitx.tx.R = R
	pitx.tx.S = S
	return true, nil
}

func (pitx *PersonalInfoTx) Cost() *big.Int {
	fee := new(big.Int)
	fee.Set(pitx.tx.Fee)
	return fee
}

func (pitx *PersonalInfoTx) RawSignatureValues() (v, r, s *big.Int) {
	return pitx.tx.V, pitx.tx.R, pitx.tx.S
}

func (pitx *PersonalInfoTx) GetFrom() atomic.Value {
	return pitx.from
}

func (pitx *PersonalInfoTx) GetSigV() *big.Int {
	if pitx.tx.V != nil {
		return pitx.tx.V
	}
	return nil
}

func (pitx *PersonalInfoTx) GetSigR() *big.Int {
	if pitx.tx.R != nil {
		return pitx.tx.R
	}
	return nil
}

func (pitx *PersonalInfoTx) GetSigS() *big.Int {
	if pitx.tx.S != nil {
		return pitx.tx.S
	}
	return nil
}

func (pitx *PersonalInfoTx) GetNounce() uint64 {
	return pitx.tx.Nonce
}
func (pitx *PersonalInfoTx) GetFee() uint64 {
	return pitx.tx.Fee.Uint64()
}
func (pitx *PersonalInfoTx) GetReceiver() common.Address {
	return common.Address{}
}
func (pitx *PersonalInfoTx) GetAmount() big.Int {
	return big.Int{}
}
