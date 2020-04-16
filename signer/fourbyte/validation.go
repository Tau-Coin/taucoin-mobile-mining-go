// Copyright 2019 The go-tau Authors
// This file is part of the go-tau library.
//
// The go-tau library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-tau library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-tau library. If not, see <http://www.gnu.org/licenses/>.

package fourbyte

import (
	"bytes"
	"fmt"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/signer/core"
)

// ValidateTransaction does a number of checks on the supplied transaction, and
// returns either a list of warnings, or an error (indicating that the transaction
// should be immediately rejected).
func (db *Database) ValidateTransaction(selector *string, tx *core.SendTxArgs) (*core.ValidationMessages, error) {
	messages := new(core.ValidationMessages)

	// Not a contract creation, validate as a plain transaction
	if !tx.To.ValidChecksum() {
		messages.Warn("Invalid checksum on recipient address")
	}
	if bytes.Equal(tx.To.Address().Bytes(), common.Address{}.Bytes()) {
		messages.Crit("Transaction recipient is the zero address")
	}
	// Semantic fields validated, try to make heads or tails of the call data
	//tx-ctc
	var data []byte
	db.validateCallData(selector, data, messages)
	return messages, nil
}

// validateCallData checks if the ABI call-data + method selector (if given) can
// be parsed and seems to match.
func (db *Database) validateCallData(selector *string, data []byte, messages *core.ValidationMessages) {
	// If the data is empty, we have a plain value transfer, nothing more to do
	if len(data) == 0 {
		return
	}
	// Validate the call data that it has the 4byte prefix and the rest divisible by 32 bytes
	if len(data) < 4 {
		messages.Warn("Transaction data is not valid ABI (missing the 4 byte call prefix)")
		return
	}
	if n := len(data) - 4; n%32 != 0 {
		messages.Warn(fmt.Sprintf("Transaction data is not valid ABI (length should be a multiple of 32 (was %d))", n))
	}
	// If a custom method selector was provided, validate with that
	if selector != nil {
		if info, err := verifySelector(*selector, data); err != nil {
			messages.Warn(fmt.Sprintf("Transaction contains data, but provided ABI signature could not be matched: %v", err))
		} else {
			messages.Info(info.String())
			db.AddSelector(*selector, data[:4])
		}
		return
	}
	// No method selector was provided, check the database for embedded ones
	embedded, err := db.Selector(data[:4])
	if err != nil {
		messages.Warn(fmt.Sprintf("Transaction contains data, but the ABI signature could not be found: %v", err))
		return
	}
	if info, err := verifySelector(embedded, data); err != nil {
		messages.Warn(fmt.Sprintf("Transaction contains data, but provided ABI signature could not be varified: %v", err))
	} else {
		messages.Info(info.String())
	}
}
