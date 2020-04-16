// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
// borrowed from bitcoind
package crypto

import (
	"crypto/ecdsa"
	"math/big"
)

// These constants define the lengths of serialized public keys.
const (
	PubKeyBytesLenCompressed   = 33
	PubKeyBytesLenUncompressed = 65
	PubKeyBytesLenHybrid       = 65
)

func isOdd(a *big.Int) bool {
	return a.Bit(0) == 1
}

const (
	pubkeyCompressed   byte = 0x2 // y_bit + x coord
	pubkeyUncompressed byte = 0x4 // x coord + y coord
	pubkeyHybrid       byte = 0x6 // y_bit + x coord + y coord
)

// IsCompressedPubKey returns true the the passed serialized public key has
// been encoded in compressed format, and false otherwise.
func IsCompressedPubKey(pubKey []byte) bool {
	// The public key is only compressed if it is the correct length and
	// the format (first byte) is one of the compressed pubkey values.
	return len(pubKey) == PubKeyBytesLenCompressed &&
		(pubKey[0]&^byte(0x1) == pubkeyCompressed)
}

// PublicKey is an ecdsa.PublicKey with additional functions to
// serialize in uncompressed, compressed, and hybrid formats.
type PublicKey ecdsa.PublicKey

// ToECDSA returns the public key as a *ecdsa.PublicKey.
func (p *PublicKey) ToECDSA() *ecdsa.PublicKey {
	return (*ecdsa.PublicKey)(p)
}

// SerializeUncompressed serializes a public key in a 65-byte uncompressed
// format.
func (p *PublicKey) SerializeUncompressed() []byte {
	b := make([]byte, 0, PubKeyBytesLenUncompressed)
	b = append(b, pubkeyUncompressed)
	b = paddedAppend(32, b, p.X.Bytes())
	return paddedAppend(32, b, p.Y.Bytes())
}

// SerializeCompressed serializes a public key in a 33-byte compressed format.
func (p *PublicKey) SerializeCompressed() []byte {
	b := make([]byte, 0, PubKeyBytesLenCompressed)
	format := pubkeyCompressed
	if isOdd(p.Y) {
		format |= 0x1
	}
	b = append(b, format)
	return paddedAppend(32, b, p.X.Bytes())
}

// SerializeHybrid serializes a public key in a 65-byte hybrid format.
func (p *PublicKey) SerializeHybrid() []byte {
	b := make([]byte, 0, PubKeyBytesLenHybrid)
	format := pubkeyHybrid
	if isOdd(p.Y) {
		format |= 0x1
	}
	b = append(b, format)
	b = paddedAppend(32, b, p.X.Bytes())
	return paddedAppend(32, b, p.Y.Bytes())
}

// IsEqual compares this PublicKey instance to the one passed, returning true if
// both PublicKeys are equivalent. A PublicKey is equivalent to another, if they
// both have the same X and Y coordinate.
func (p *PublicKey) IsEqual(otherPubKey *PublicKey) bool {
	return p.X.Cmp(otherPubKey.X) == 0 &&
		p.Y.Cmp(otherPubKey.Y) == 0
}

// paddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}
