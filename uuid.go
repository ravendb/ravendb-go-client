package ravendb

import (
	"crypto/rand"
	"encoding/hex"
)

// implements generating random uuid4 that mimics python's uuid.uuid4()
// it doesn't try to fully UUIDv4 compliant

// UUID represents a random 16-byte number
type UUID struct {
	data [16]byte
}

// NewUUID creates a new UUID
func NewUUID() *UUID {
	res := &UUID{}
	n, _ := rand.Read(res.data[:])
	panicIf(n != 16, "rand.Read() returned %d, expected 16", n)
	return res
}

// String returns xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx representation
func (u *UUID) String() string {
	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u.data[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u.data[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u.data[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u.data[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u.data[10:])

	return string(buf)
}

// Hex returns hex-encoded version.
// Equivalent of python's uuid.uuid4().hex
func (u *UUID) Hex() string {
	dst := make([]byte, 32)
	n := hex.Encode(dst, u.data[:])
	panicIf(n != 32, "hex.Encode() returned %d, expected 32", n)
	return string(dst)
}
