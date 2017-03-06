package types

import "crypto/tls"

// TLSConfig is a custom TLS configuration that includes the concept of a
// peer certificate's fingerprint.
type TLSConfig struct {
	tls.Config

	// PeerFingerprint is the expected SHA256 fingerprint of a peer certificate.
	PeerFingerprint []byte
}
