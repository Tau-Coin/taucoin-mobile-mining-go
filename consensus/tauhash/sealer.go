// Copyright 2017 The go-tau Authors
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

package tauhash

import (
	crand "crypto/rand"
	"errors"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/consensus"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/types"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/log"
	"math"
	"math/big"
	"math/rand"
	"runtime"
	"sync"
)

var (
	errNoMiningWork      = errors.New("no mining work available yet")
	errInvalidSealResult = errors.New("invalid or stale proof-of-work solution")
)

// Seal implements consensus.Engine, attempting to find a nonce that satisfies
// the block's difficulty requirements.
func (tauhash *Tauash) Seal(chain consensus.ChainReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	// Create a runner and the multiple search threads it directs
	abort := make(chan struct{})

	tauhash.lock.Lock()
	threads := tauhash.threads
	if tauhash.rand == nil {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			tauhash.lock.Unlock()
			return err
		}
		tauhash.rand = rand.New(rand.NewSource(seed.Int64()))
	}
	tauhash.lock.Unlock()
	if threads == 0 {
		threads = runtime.NumCPU()
	}
	if threads < 0 {
		threads = 0 // Allows disabling local mining without extra logic around local/remote
	}
	var (
		pend   sync.WaitGroup
		locals = make(chan *types.Block)
	)
	for i := 0; i < threads; i++ {
		pend.Add(1)
		go func(id int, nonce uint64) {
			defer pend.Done()
			tauhash.mine(block, id, nonce, abort, locals)
		}(i, uint64(tauhash.rand.Int63()))
	}
	// Wait until sealing is terminated or a nonce is found
	go func() {
		var result *types.Block
		select {
		case <-stop:
			// Outside abort, stop all miner threads
			close(abort)
		case result = <-locals:
			// One of the threads found a block, abort all others
			select {
			case results <- result:
			default:
				log.Warn("Sealing result is not read by miner", "mode", "local", "sealhash", tauhash.SealHash(block.Header()))
			}
			close(abort)
		case <-tauhash.update:
			// Thread count was changed on user request, restart
			close(abort)
			if err := tauhash.Seal(chain, block, results, stop); err != nil {
				log.Error("Failed to restart sealing after update", "err", err)
			}
		}
		// Wait for all miners to terminate and return the block
		pend.Wait()
	}()
	return nil
}

// mine is the actual proof-of-work miner that searches for a nonce starting from
// seed that results in correct final block difficulty.
func (tauhash *Tauash) mine(block *types.Block, id int, seed uint64, abort chan struct{}, found chan *types.Block) {
	// Extract some data from the header
	var (
		header  = block.Header()
		hash    = tauhash.SealHash(header).Bytes()
		target  = new(big.Int).Div(two256, header.Difficulty)
		number  = header.Number.Uint64()
		dataset = tauhash.dataset(number, false)
	)
	// Start generating random nonces until we abort or find a good one
	var (
		nonce = seed
	)
	logger := log.New("miner", id)
	logger.Trace("Started tauhash search for new nonces", "seed", seed)
search:
	for {
		select {
		case <-abort:
			// Mining terminated, update stats and abort
			logger.Trace("Tauash nonce search aborted", "attempts", nonce-seed)
			break search

		default:
			// Compute the PoW value of this nonce
			digest, result := hashimotoFull(dataset.dataset, hash, nonce)
			if new(big.Int).SetBytes(result).Cmp(target) <= 0 {
				// Correct nonce found, create a new header with it
				header = types.CopyHeader(header)
				header.Nonce = types.EncodeNonce(nonce)
				header.MixDigest = common.BytesToHash(digest)

				// Seal and return a block (if still needed)
				select {
				case found <- block.WithSeal(header):
					logger.Trace("Tauash nonce found and reported", "attempts", nonce-seed, "nonce", nonce)
				case <-abort:
					logger.Trace("Tauash nonce found but discarded", "attempts", nonce-seed, "nonce", nonce)
				}
				break search
			}
			nonce++
		}
	}
	// Datasets are unmapped in a finalizer. Ensure that the dataset stays live
	// during sealing so it's not unmapped while being read.
	runtime.KeepAlive(dataset)
}
