package types

// AuthToken is a JSON Web Token.
//
// All fields related to times are stored as UTC epochs in seconds.
type AuthToken struct {
	// Subject is the intended principal of the token.
	Subject string `json:"sub"`

	// Expires is the time at which the token expires.
	Expires int64 `json:"exp"`

	// NotBefore is the the time at which the token becomes valid.
	NotBefore int64 `json:"nbf"`

	// IssuedAt is the time at which the token was issued.
	IssuedAt int64 `json:"iat"`

	// Encoded is the encoded JWT string.
	Encoded string `json:"enc"`
}

// String returns the subject of the security token.
func (s *AuthToken) String() string {
	return s.Subject
}

// AuthConfig is the auth configuration.
type AuthConfig struct {

	// Disabled is a flag indicating whether the auth configuration is disabled.
	Disabled bool

	// Allow is a list of allowed tokens.
	Allow []string

	// Deny is a list of denied tokens.
	Deny []string

	// Key is the signing key.
	Key []byte

	// Alg is the cryptographic algorithm used to sign and verify the token.
	Alg string
}
