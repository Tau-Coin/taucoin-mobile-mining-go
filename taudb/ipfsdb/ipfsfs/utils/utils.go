package utils

import (
	"errors"

    "github.com/ipfs/interface-go-ipfs-core/path"

    cid "github.com/ipfs/go-cid"
	mh  "github.com/multiformats/go-multihash"
)

func ByteToPath(b []byte) (path.Path, error){
	// byte into cid
	c, err := cid.Decode(string(b))
	if err != nil{
       return nil, err
	}

	// cid into path
	res := path.IpfsPath(c)

	return res, nil
}

// keccak256ToCid takes a keccak256 hash and returns its cid based on
// the codec given.
func Keccak256ToPath(codec uint64, h []byte) (path.Path, error) {
    buf, err := mh.Encode(h, mh.KECCAK_256)
    if err != nil {
        panic(err)
    }

    return path.IpfsPath(cid.NewCidV1(codec, mh.Multihash(buf))), nil
}

var (
	ErrReleased    = errors.New("ipfsdb: resource already relesed")
	ErrHasReleaser = errors.New("ipfsdb: releaser already defined")
)

// Releaser is the interface that wraps the basic Release method.
type Releaser interface {
	// Release releases associated resources. Release should always success
	// and can be called multiple times without causing error.
	Release()
}

// ReleaseSetter is the interface that wraps the basic SetReleaser method.
type ReleaseSetter interface {
	// SetReleaser associates the given releaser to the resources. The
	// releaser will be called once coresponding resources released.
	// Calling SetReleaser with nil will clear the releaser.
	//
	// This will panic if a releaser already present or coresponding
	// resource is already released. Releaser should be cleared first
	// before assigned a new one.
	SetReleaser(releaser Releaser)
}

// BasicReleaser provides basic implementation of Releaser and ReleaseSetter.
type BasicReleaser struct {
	releaser Releaser
	released bool
}

// Released returns whether Release method already called.
func (r *BasicReleaser) Released() bool {
	return r.released
}

// Release implements Releaser.Release.
func (r *BasicReleaser) Release() {
	if !r.released {
		if r.releaser != nil {
			r.releaser.Release()
			r.releaser = nil
		}
		r.released = true
	}
}

// SetReleaser implements ReleaseSetter.SetReleaser.
func (r *BasicReleaser) SetReleaser(releaser Releaser) {
	if r.released {
		panic(ErrReleased)
	}
	if r.releaser != nil && releaser != nil {
		panic(ErrHasReleaser)
	}
	r.releaser = releaser
}

type NoopReleaser struct{}

func (NoopReleaser) Release() {}
