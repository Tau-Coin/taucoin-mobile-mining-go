package types

import (
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common/hexutil"
)

type NewChainTx struct {
	gx      GeneralTx
	name    hexutil.Byte20
	contact hexutil.Byte32
}
