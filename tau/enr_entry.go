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

package tau

import (
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/forkid"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/p2p/enode"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/rlp"
)

// tauEntry is the "tau" ENR entry which advertises tau protocol
// on the discovery network.
type tauEntry struct {
	ForkID forkid.ID // Fork identifier per EIP-2124

	// Ignore additional fields (for forward compatibility).
	Rest []rlp.RawValue `rlp:"tail"`
}

// ENRKey implements enr.Entry.
func (e tauEntry) ENRKey() string {
	return "tau"
}

func (tau *Tau) startTauEntryUpdate(ln *enode.LocalNode) {
	var newHead = make(chan core.ChainHeadEvent, 10)
	sub := tau.blockchain.SubscribeChainHeadEvent(newHead)

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-newHead:
				ln.Set(tau.currentTauEntry())
			case <-sub.Err():
				// Would be nice to sync with tau.Stop, but there is no
				// good way to do that.
				return
			}
		}
	}()
}

func (tau *Tau) currentTauEntry() *tauEntry {
	return &tauEntry{ForkID: forkid.NewID(tau.blockchain)}
}
