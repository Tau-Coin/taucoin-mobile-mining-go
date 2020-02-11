package types

import (
	"math/big"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common/hexutil"
)

//go:generate gencodec -type GeneralTx -out common_tx_json.go
const (
	ChainIDlength = 32
)

// type GeneralTx struct {
// 	Version   byte                `json:"version"    gencodec:"required"`
// 	Option    byte                `json:"option"     gencodec:"required"`
// 	ChainID   [ChainIDlength]byte `json:"chainid"  gencodec:"required"`
// 	Nounce    uint64              `json:"nounce"     gencodec:"required"`
// 	TimeStamp uint32              `json:"timestamp"   gencodec:"required"`
// 	Fee       byte                `json:"fee"         gencodec:"required"`
// 	Signature TXSignature         `json:"signature"   gencodec:"required"`
// 	//whatever marshing/unmarshing method sender is needed
// 	Sender *common.Address `json:"sender"        rlp:"required"`
// }

type GeneralTx struct {
	Version   hexutil.OneByte       `json:"version"    gencodec:"required"`
	Option    hexutil.OneByte       `json:"option"     gencodec:"required"`
	ChainID   hexutil.Byte32        `json:"chainid"  gencodec:"required"`
	Nounce    hexutil.Uint64        `json:"nounce"     gencodec:"required"`
	TimeStamp hexutil.Uint32        `json:"timestamp"   gencodec:"required"`
	Fee       hexutil.OneByte       `json:"fee"         gencodec:"required"`
	Signature TXSignatureMarshaling `json:"signature"   gencodec:"required"`
	//whatever marshing/unmarshing method sender is needed
	Sender *common.Address `json:"sender"        rlp:"required"`
}

type GeneralTxMarshaling struct {
	Version   hexutil.OneByte
	Option    hexutil.OneByte
	ChainID   hexutil.Byte32
	Nounce    hexutil.Uint64
	TimeStamp hexutil.Uint32
	Fee       hexutil.OneByte
	Signature TXSignatureMarshaling
}
type TXSignature struct {
	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`
}

type TXSignatureMarshaling struct {
	V *hexutil.Big
	R *hexutil.Big
	S *hexutil.Big
}
