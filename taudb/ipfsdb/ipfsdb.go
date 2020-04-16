// Copyright 2018 The go-tau Authors
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

// +build !js

// Package leveldb implements the key-value database layer based on IpfsDB.
package ipfsdb

import (
	"context"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/log"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/taudb"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/taudb/ipfsdb/ipfsfs"

)

// Database is a persistent key-value store. Apart from basic data storage
// functionality it also supports batch writes and iterating over the keyspace in
// binary-alphabetical order.
type Database struct {
	idb *ipfsfs.IPFSdb // IPFSDB instance

	log log.Logger // Contextual logger tracking the database path
}

// New returns a wrapped LevelDB object. The namespace is the prefix that the
// metrics reporting should use for surfacing internal stats.
func New() (*Database, error) {
	logger := log.New("New ipfs database")

	var ctx context.Context

	// Open the idb and recover any potential corruptions
	idbTmp := ipfsfs.NewIPFSdb(ctx)

	// Assemble the wrapper with all the registered metrics
	db := &Database{
		idb:      idbTmp,
		log:      logger,
	}

	return db, nil
}

// Has retrieves if a key is present in the key-value store.
func (db *Database) Has(key []byte) (bool, error) {
	return db.idb.Has(key)
}

// Get retrieves the given key if it's present in the key-value store.
func (db *Database) Get(key []byte) ([]byte, error) {
	dat, err := db.idb.Get(key)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// Put inserts the given value into the key-value store.
func (db *Database) Put(key []byte, value []byte) error {
	log.Info("In IPFS DB", "Key", string(key))
	return db.idb.Put(key, value)
}

// Delete removes the key from the key-value store.
func (db *Database) Delete(key []byte) error {
	return db.idb.Delete(key)
}

// NewBatch creates a write-only key-value store that buffers changes to its host
// database until a final write is called.
func (db *Database) NewBatch() taudb.Batch {
	return &batch{
		db: db.idb,
		b:  new(ipfsfs.Batch),
	}
}

// batch is a write-only leveldb batch that commits changes to its host database
// when Write is called. A batch cannot be used concurrently.
type batch struct {
	db   *ipfsfs.IPFSdb
	b    *ipfsfs.Batch
	size int
}

// Put inserts the given value into the batch for later committing.
func (b *batch) Put(key, value []byte) error {
	b.b.Put(key, value)
	b.size += len(value)
	return nil
}

// Delete inserts the a key removal into the batch for later committing.
func (b *batch) Delete(key []byte) error {
	b.b.Delete(key)
	b.size++
	return nil
}

// ValueSize retrieves the amount of data queued up for writing.
func (b *batch) ValueSize() int {
	return b.size
}

// Write flushes any accumulated data to disk.
func (b *batch) Write() error {
	//log.Info("ctc debug batch write")
	return b.db.Write(b.b)
}

// Reset resets the batch for reuse.
func (b *batch) Reset() {
	b.b.Reset()
	b.size = 0
}

// Replay replays the batch contents.
func (b *batch) Replay(w taudb.KeyValueWriter) error {
	return b.b.Replay(&replayer{writer: w})
}

// replayer is a small wrapper to implement the correct replay methods.
type replayer struct {
	writer  taudb.KeyValueWriter
	failure error
}

// Put inserts the given value into the key-value data store.
func (r *replayer) Put(key, value []byte) {
	// If the replay already failed, stop executing ops
	if r.failure != nil {
		return
	}
	r.failure = r.writer.Put(key, value)
}

// Delete removes the key from the key-value data store.
func (r *replayer) Delete(key []byte) {
	// If the replay already failed, stop executing ops
	if r.failure != nil {
		return
	}
	r.failure = r.writer.Delete(key)
}
