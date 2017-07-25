package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codedellemc/rexray/libstorage/api/context"
	"github.com/codedellemc/rexray/libstorage/api/types"
)

const (
	jwtAlg       = "HS256"
	jwtKey       = "key"
	jwtAkutz     = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MjI2ODg1NTAsImlhdCI6MTQ5MTIzODk1MCwibmJmIjoxNDkxMjM4OTUwLCJzdWIiOiJha3V0eiJ9.3eAA7AQZUGrwA42H64qKbu8QF_AHpSsJSMR0FALnKj8`
	jwtCduchesne = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MjI2OTM1ODQsImlhdCI6MTQ5MTI0Mzk4NCwibmJmIjoxNDkxMjQzOTg0LCJzdWIiOiJjZHVjaGVzbmUifQ.AUOrtC41LQB5FO1NsBE357o_Zsx-lhZ-3I7v_UMsTh4`
)

func TestValidateAuthToken_ValidTokSigKey(t *testing.T) {
	sc := &types.AuthConfig{
		Key:   []byte(jwtKey),
		Alg:   jwtAlg,
		Allow: []string{"akutz"},
	}
	tok, err := ValidateAuthTokenWithJWT(context.Background(), sc, jwtAkutz)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NotNil(t, tok) {
		t.FailNow()
	}
	if !assert.Equal(t, "akutz", tok.Subject) {
		t.FailNow()
	}
}

func TestValidateAuthToken_NotInAllowList(t *testing.T) {
	sc := &types.AuthConfig{
		Key: []byte(jwtKey),
		Alg: jwtAlg,
	}
	tok, err := ValidateAuthTokenWithJWT(context.Background(), sc, jwtAkutz)
	if !assert.Error(t, err) {
		t.FailNow()
	}
	if !assert.Nil(t, tok) {
		t.FailNow()
	}
	if !assert.IsType(t, &types.ErrSecTokInvalid{}, err) {
		t.FailNow()
	}
	terr := err.(*types.ErrSecTokInvalid)
	if !assert.True(t, terr.Denied) {
		t.FailNow()
	}
}

func TestValidateAuthToken_InDenyList(t *testing.T) {
	sc := &types.AuthConfig{
		Key:  []byte(jwtKey),
		Alg:  jwtAlg,
		Deny: []string{"akutz"},
	}
	tok, err := ValidateAuthTokenWithJWT(context.Background(), sc, jwtAkutz)
	if !assert.Error(t, err) {
		t.FailNow()
	}
	if !assert.Nil(t, tok) {
		t.FailNow()
	}
	if !assert.IsType(t, &types.ErrSecTokInvalid{}, err) {
		t.FailNow()
	}
	terr := err.(*types.ErrSecTokInvalid)
	if !assert.True(t, terr.Denied) {
		t.FailNow()
	}
}

func TestValidateAuthToken_InvalidKey(t *testing.T) {
	sc := &types.AuthConfig{
		Key: []byte("invalidKey"),
		Alg: jwtAlg,
	}
	tok, err := ValidateAuthTokenWithJWT(context.Background(), sc, jwtAkutz)
	if !assert.Error(t, err) {
		t.FailNow()
	}
	if !assert.Nil(t, tok) {
		t.FailNow()
	}
	if !assert.IsType(t, &types.ErrSecTokInvalid{}, err) {
		t.FailNow()
	}
	terr := err.(*types.ErrSecTokInvalid)
	if !assert.True(t, terr.InvalidSig) {
		t.FailNow()
	}
}
