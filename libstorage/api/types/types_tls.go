package types

import (
	"crypto/tls"
	"fmt"
)

// TLSConfig is a custom TLS configuration that includes the concept of a
// peer certificate's fingerprint.
type TLSConfig struct {
	tls.Config

	// VerifyPeers is a flag that indicates whether peer certificates
	// should be validated against a PeerFingerprint or known hosts files.
	VerifyPeers bool

	// SysKnownHosts is the path to the system's known_hosts file.
	SysKnownHosts string

	// UsrKnownHosts is the path to the user's known_hosts file.
	UsrKnownHosts string

	// KnownHost is the trusted, remote host information.
	KnownHost *TLSKnownHost
}

// TLSKnownHost contains the identifying information of trusted, remote peer.
type TLSKnownHost struct {

	// Host is the name of the known host. This value is derived from the
	// CommonName value in the remote host's certiicate.
	Host string

	// Alg is the cryptographic algorithm used to calculate the fingerprint.
	Alg string

	// Fingerprint is known host's certificate's fingerprint.
	Fingerprint []byte
}

// String returns a known_hosts string.
func (kh *TLSKnownHost) String() string {
	return fmt.Sprintf("%s %s %x", kh.Host, kh.Alg, kh.Fingerprint)
}
