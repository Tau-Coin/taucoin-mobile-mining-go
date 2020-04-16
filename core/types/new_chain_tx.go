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
	"math/big"
	"sync/atomic"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common/hexutil"
)

//go:generate gencodec -type NewChainTxData -field-override NewChainTxDataMarshaling -out new_chain_tx_json.go
type OneByte []byte
type Byte32s []byte
type Byte20s []byte
type Byte144s []byte

type NewChainTx struct {
	tx NewChainTxData

	hash atomic.Value
	size atomic.Value
	from atomic.Value
}
type NewChainTxData struct {
	Version   OneByte         `json:"version"     gencodec:"required"`
	Option    OneByte         `json:"option"      gencodec:"required"`
	ChainID   Byte32s         `json:"chainid"     gencodec:"required"`
	Nounce    uint64          `json:"nounce"      gencodec:"required"`
	TimeStamp uint32          `json:"timestamp"   gencodec:"required"`
	Fee       OneByte         `json:"fee"         gencodec:"required"`
	V         *big.Int        `json:"v"           gencodec:"required"`
	R         *big.Int        `json:"r"           gencodec:"required"`
	S         *big.Int        `json:"s"           gencodec:"required"`
	Sender    *common.Address `json:"sender"        rlp:"required"`

	Name        Byte20s  `json:"name"      gencodec:"required"`
	Contact     Byte32s  `json:"contact"   gencodec:"required"`
	Title       Byte144s `json:"title"     gencodec:"required"`
	Description Byte32s  `json:"description" gencodec:"requried"`
}

type NewChainTxDataMarshaling struct {
	Version   hexutil.Bytes
	Option    hexutil.Bytes
	ChainID   hexutil.Bytes
	Nounce    hexutil.Uint64
	TimeStamp hexutil.Uint32
	Fee       hexutil.Bytes
	V         *hexutil.Big
	R         *hexutil.Big
	S         *hexutil.Big

	Name        hexutil.Bytes
	Contact     hexutil.Bytes
	Title       hexutil.Bytes
	Description hexutil.Bytes
}

func NewNewChainTransaction() {

}
