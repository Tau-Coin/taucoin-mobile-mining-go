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

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/common"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common/hexutil"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/common/math"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/rawdb"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/state"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/core/vm"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/crypto"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/params"
)

// VMTest checks EVM execution without block or transaction context.
// See https://github.com/tau/tests/wiki/VM-Tests for the test format specification.
type VMTest struct {
	json vmJSON
}

func (t *VMTest) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.json)
}

type vmJSON struct {
	Env           stEnv                 `json:"env"`
	Exec          vmExec                `json:"exec"`
	Logs          common.UnprefixedHash `json:"logs"`
	GasRemaining  *math.HexOrDecimal64  `json:"gas"`
	Out           hexutil.Bytes         `json:"out"`
	Pre           core.GenesisAlloc     `json:"pre"`
	Post          core.GenesisAlloc     `json:"post"`
	PostStateRoot common.Hash           `json:"postStateRoot"`
}

//go:generate gencodec -type vmExec -field-override vmExecMarshaling -out gen_vmexec.go

type vmExec struct {
	Address  common.Address `json:"address"  gencodec:"required"`
	Caller   common.Address `json:"caller"   gencodec:"required"`
	Origin   common.Address `json:"origin"   gencodec:"required"`
	Value    *big.Int       `json:"value"    gencodec:"required"`
	Fee      *big.Int       `json:"gasPrice" gencodec:"required"`
}

type vmExecMarshaling struct {
	Address  common.UnprefixedAddress
	Caller   common.UnprefixedAddress
	Origin   common.UnprefixedAddress
	Value    *math.HexOrDecimal256
	Fee *math.HexOrDecimal256
}

func (t *VMTest) Run(vmconfig vm.Config) error {
	statedb := MakePreState(rawdb.NewMemoryDatabase(), t.json.Pre)
	ret, _ := t.exec(statedb, vmconfig)

	// Test declares gas, expecting outputs to match.
	if !bytes.Equal(ret, t.json.Out) {
		return fmt.Errorf("return data mismatch: got %x, want %x", ret, t.json.Out)
	}

	return nil
}

func (t *VMTest) exec(statedb *state.StateDB, vmconfig vm.Config) ([]byte, error) {
	evm := t.newEVM(statedb, vmconfig)
	e := t.json.Exec
	return evm.Call(vm.AccountRef(e.Caller), e.Address, e.Fee.Uint64(), e.Value)
}

func (t *VMTest) newEVM(statedb *state.StateDB, vmconfig vm.Config) *vm.EVM {
	initialCall := true
	canTransfer := func(db vm.StateDB, address common.Address, amount *big.Int) bool {
		if initialCall {
			initialCall = false
			return true
		}
		return core.CanTransfer(db, address, amount)
	}
	transfer := func(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {}
	context := vm.Context{
		CanTransfer: canTransfer,
		Transfer:    transfer,
		GetHash:     vmTestBlockHash,
		Origin:      t.json.Exec.Origin,
		Coinbase:    t.json.Env.Coinbase,
		BlockNumber: new(big.Int).SetUint64(t.json.Env.Number),
		Time:        new(big.Int).SetUint64(t.json.Env.Timestamp),
		Difficulty:  t.json.Env.Difficulty,
		Fee:         t.json.Exec.Fee,
	}
	vmconfig.NoRecursion = true
	return vm.NewEVM(context, statedb, params.MainnetChainConfig)
}

func vmTestBlockHash(n uint64) common.Hash {
	return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
}
