package ipfsfs

import (
	"bytes"
    "context"

	"github.com/Tau-Coin/taucoin-mobile-mining-go/log"
	"github.com/Tau-Coin/taucoin-mobile-mining-go/taudb/ipfsdb/ipfsfs/utils"

	ipfs      "github.com/ipfs/go-ipfs/lib"
    coreiface "github.com/ipfs/interface-go-ipfs-core"
    caopts    "github.com/ipfs/interface-go-ipfs-core/options"
	mh        "github.com/multiformats/go-multihash"
)

type IPFSdb struct {
    ctx  context.Context
    api  coreiface.CoreAPI
}

func NewIPFSdb(ctx context.Context) *IPFSdb {
	coreapi, _:= ipfs.API()
    return &IPFSdb{
		ctx: ctx,
		api: coreapi,
    }
}

func (db *IPFSdb) Put(key, value []byte) error {
	// value -> io.Reader
	reader := bytes.NewReader(value)

	opt := func(bs *caopts.BlockPutSettings) error {
		bs.Codec = "raw"
		bs.MhType = mh.KECCAK_256
		bs.MhLength = -1
		bs.Pin = true
		return nil
	}

	blockStat, err:= db.api.Block().Put(db.ctx, reader, opt)
	log.Info("IPFS Put", "block cid", blockStat.Path())
	return err
}

func (db *IPFSdb) Get(key []byte) ([]byte, error) {
	// key -> path
	path, err:= utils.Keccak256ToPath(0xa0, key)
	if err != nil {
		return nil, err
	}

	reader, err:= db.api.Block().Get(db.ctx, path)
	if err != nil{
		return nil, err
	}

    var data []byte
	_, errRead := reader.Read(data)

	return data, errRead
}

func (db *IPFSdb) Delete(key []byte) error {
	// key -> path
	path, err:= utils.Keccak256ToPath(0x90, key)
	if err != nil {
		log.Info("Ipfsfs Delete", "err", err)
		return err
	}

    return db.api.Block().Rm(db.ctx, path)
}

func (db *IPFSdb) Has(key []byte) (bool, error) {
	// key -> path
	path, err:= utils.Keccak256ToPath(0x90, key)
	if err != nil {
		return false, err
	}

	blockStat, err:= db.api.Block().Stat(db.ctx, path)

	return blockStat.Size()> 0, err
}

// TBD
func (db *IPFSdb) Write(batch *Batch) error {
	if batch == nil || batch.Len() == 0 {
        return nil
    }

	log.Info("Ipfsfs write", "size", batch.internalLen)

	for i:= 0; i< batch.internalLen; i++{

		keyStart := batch.index[i].keyPos
		keyEnd := keyStart+ batch.index[i].keyLen

		valueStart := batch.index[i].valuePos
		valueEnd := valueStart+ batch.index[i].valueLen

		keyTmp := batch.data[keyStart : keyEnd]
		valueTmp := batch.data[valueStart : valueEnd]

		err:= db.Put(keyTmp, valueTmp)
		if err != nil {
			return err
		}
	}
	return nil
}
