package utils

import (
	"crypto/rand"
	"fmt"
)

// The UUID reserved variants.
const (
	ReservedNCS       byte = 0x80
	ReservedRFC4122   byte = 0x40
	ReservedMicrosoft byte = 0x20
	ReservedFuture    byte = 0x00
)

// UUID is a UUID. This was totally stolen from
// https://github.com/nu7hatch/gouuid/blob/master/uuid.go, and all credit goes
// to that author. It was included like this in order to reduce external
// dependencies.
type UUID [16]byte

// NewUUID returns a new UUID.
func NewUUID() (*UUID, error) {
	u := &UUID{}
	if _, err := rand.Read(u[:]); err != nil {
		return nil, err
	}
	u.setVariant(ReservedRFC4122)
	u.setVersion(4)
	return u, nil
}

// String returns unparsed version of the generated UUID sequence.
func (u *UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

// setVariant sets the two most significant bits (bits 6 and 7) of the
// clock_seq_hi_and_reserved to zero and one, respectively.
func (u *UUID) setVariant(v byte) {
	switch v {
	case ReservedNCS:
		u[8] = (u[8] | ReservedNCS) & 0xBF
	case ReservedRFC4122:
		u[8] = (u[8] | ReservedRFC4122) & 0x7F
	case ReservedMicrosoft:
		u[8] = (u[8] | ReservedMicrosoft) & 0x3F
	}
}

// setVersion sets the four most significant bits (bits 12 through 15) of the
// time_hi_and_version field to the 4-bit version number.
func (u *UUID) setVersion(v byte) {
	u[6] = (u[6] & 0xF) | (v << 4)
}
