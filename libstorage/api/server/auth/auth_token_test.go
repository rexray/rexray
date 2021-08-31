package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

const (
	jwtAlg       = "HS256"
	jwtKey       = "key"
	jwtAkutz     = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjMzMDgyMTkyMDk1LCJpYXQiOjE0OTEyMzg5NTAsIm5iZiI6MTQ5MTIzODk1MCwic3ViIjoiYWt1dHoifQ.DfMArIDErbr6aU2n01UgGz6vGXsAqUJ3UOmtr0SaQzA`
	jwtCduchesne = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjMzMDgyMTkyMDk1LCJpYXQiOjE0OTEyNDM5ODQsIm5iZiI6MTQ5MTI0Mzk4NCwic3ViIjoiY2R1Y2hlc25lIn0.CKPVnD2eFb9RasLg-i2QZyjt0kgVNYpzpML086LWFDw`
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
