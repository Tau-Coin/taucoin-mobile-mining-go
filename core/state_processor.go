// Copyright 2015 The go-tau Authors
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

package core

import (
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/consensus"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/state"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/types"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/vm"
	// "github.com/Tau-Coin/taucoin-mobile-mining-go/crypto" ctc delete
	"github.com/Tau-Coin/taucoin-mobile-mining-go/params"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Tau rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB) error {
	var (
		header = block.Header()
	)
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		statedb.Prepare((*tx).Hash(), block.Hash(), i)
		err := ApplyTransaction(p.config, p.bc, nil, statedb, header, tx)
		if err != nil {
			return err
		}
	}
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Transactions())

	return nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, statedb *state.StateDB, header *types.Header, tx *types.Transaction) error {
	msg, err := (*tx).AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return err
	}
	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc, author)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config)
	// Apply the transaction to the current state (included in the env)
	_, _, _, err = ApplyMessage(vmenv, msg)
	if err != nil {
		return err
	}
	return err
}
