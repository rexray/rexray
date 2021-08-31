package auth

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	jcrypto "github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

var rxBearer = regexp.MustCompile(`Bearer (.+)`)

// GetBearerTokenFromReq retrieves the bearer token from the HTTP request.
func GetBearerTokenFromReq(ctx types.Context, req *http.Request) string {
	m := rxBearer.FindStringSubmatch(
		req.Header.Get(types.AuthorizationHeader))
	if len(m) == 0 {
		return ""
	}
	return m[1]
}

// ValidateAuthTokenWithCtx validates the auth token from the provided context.
func ValidateAuthTokenWithCtx(
	ctx types.Context,
	config *types.AuthConfig) (*types.AuthToken, error) {

	if tok, ok := context.AuthToken(ctx); ok {
		logFields := map[string]interface{}{"sub": tok.Subject}
		err := validateAuthTokenAllowed(ctx, config, logFields, tok)
		if err != nil {
			return nil, err
		}
		ctx.WithFields(logFields).Info("validated security token")
		return tok, nil
	}

	return nil, nil
}

// ValidateAuthTokenWithCtxOrReq validates the auth token from the provided
// context. If not present in the context attempt a validation against the
// provided HTTP request.
func ValidateAuthTokenWithCtxOrReq(
	ctx types.Context,
	config *types.AuthConfig,
	req *http.Request) (*types.AuthToken, error) {

	tok, err := ValidateAuthTokenWithCtx(ctx, config)
	if err != nil {
		return nil, err
	}
	if tok != nil {
		return tok, nil
	}
	return ValidateAuthTokenWithReq(ctx, config, req)
}

var jwtRX = regexp.MustCompile(`^(?:(.+):)?(.+\..+\..+)$`)

func getSubject(s string) string {
	m := jwtRX.FindStringSubmatch(s)
	if len(m) == 0 {
		return s
	}
	jwt, err := jws.ParseJWT([]byte(m[2]))
	if err != nil {
		return s
	}
	if v, ok := jwt.Claims().Subject(); ok {
		return v
	}
	return s
}

func validateAuthTokenAllowed(
	ctx types.Context,
	config *types.AuthConfig,
	logFields map[string]interface{},
	tok *types.AuthToken) error {

	for _, v := range config.Deny {
		if strings.EqualFold(getSubject(v), tok.Subject) {
			ctx.WithFields(logFields).Error("access denied")
			return &types.ErrSecTokInvalid{Denied: true}
		}
	}
	for _, v := range config.Allow {
		if strings.EqualFold(getSubject(v), tok.Subject) {
			return nil
		}
	}

	ctx.WithFields(logFields).Error("access not granted")
	return &types.ErrSecTokInvalid{Denied: true}
}

// ValidateAuthTokenWithReq validates the auth token from the provided HTTP req.
func ValidateAuthTokenWithReq(
	ctx types.Context,
	config *types.AuthConfig,
	req *http.Request) (*types.AuthToken, error) {

	return ValidateAuthTokenWithJWT(
		ctx, config, GetBearerTokenFromReq(ctx, req))
}

// ValidateAuthTokenWithJWT validates the auth token from the provided JWT.
func ValidateAuthTokenWithJWT(
	ctx types.Context,
	config *types.AuthConfig,
	encJWT string) (*types.AuthToken, error) {

	lf := map[string]interface{}{"encJWT": encJWT}
	ctx.WithFields(lf).Debug("validating jwt")

	jwt, err := jws.ParseJWT([]byte(encJWT))
	if err != nil {
		ctx.WithFields(lf).WithError(err).Error("error parsing jwt")
		return nil, &types.ErrSecTokInvalid{InvalidToken: true, InnerError: err}
	}

	sm := parseSigningMethod(config.Alg)
	lf["signingMethod"] = sm.Alg()
	ctx.WithFields(lf).Debug("parsed jwt signing method")

	if err := jwt.Validate(config.Key, sm); err != nil {
		ctx.WithFields(lf).WithError(err).Error("error validating jwt")
		return nil, &types.ErrSecTokInvalid{InvalidSig: true, InnerError: err}
	}

	ctx.WithFields(lf).Debug("validated jwt signature")

	if len(jwt.Claims()) == 0 {
		ctx.WithFields(lf).Error("jwt missing claims")
		return nil, &types.ErrSecTokInvalid{}
	}

	var (
		sub string
		iat time.Time
		exp time.Time
		nbf time.Time
		ok  bool
	)

	if sub, ok = jwt.Claims().Subject(); !ok {
		ctx.WithFields(lf).Error("jwt missing sub claim")
		return nil, &types.ErrSecTokInvalid{MissingClaim: "sub"}
	}

	if iat, ok = jwt.Claims().IssuedAt(); !ok {
		ctx.WithFields(lf).Error("jwt missing iat claim")
		return nil, &types.ErrSecTokInvalid{MissingClaim: "iat"}
	}

	if exp, ok = jwt.Claims().Expiration(); !ok {
		ctx.WithFields(lf).Error("jwt missing exp claim")
		return nil, &types.ErrSecTokInvalid{MissingClaim: "exp"}
	}

	if nbf, ok = jwt.Claims().NotBefore(); !ok {
		ctx.WithFields(lf).Error("jwt missing nbf claim")
		return nil, &types.ErrSecTokInvalid{MissingClaim: "nbf"}
	}

	tok := &types.AuthToken{
		Subject:   sub,
		IssuedAt:  iat.UTC().Unix(),
		Expires:   exp.UTC().Unix(),
		NotBefore: nbf.UTC().Unix(),
	}

	lf["sub"] = tok.Subject
	lf["iat"] = tok.IssuedAt
	lf["exp"] = tok.Expires
	lf["nbf"] = tok.NotBefore

	if err := validateAuthTokenAllowed(ctx, config, lf, tok); err != nil {
		return nil, err
	}

	ctx.WithFields(lf).Info("validated security token")
	return tok, nil
}

var signingMethods = []jcrypto.SigningMethod{
	jcrypto.SigningMethodES256,
	jcrypto.SigningMethodES384,
	jcrypto.SigningMethodES512,
	jcrypto.SigningMethodHS256,
	jcrypto.SigningMethodHS384,
	jcrypto.SigningMethodHS512,
	jcrypto.SigningMethodPS256,
	jcrypto.SigningMethodPS384,
	jcrypto.SigningMethodPS512,
	jcrypto.SigningMethodRS256,
	jcrypto.SigningMethodRS384,
	jcrypto.SigningMethodRS512,
}

// parseSigningMethod parses the signing method from the provided method name.
func parseSigningMethod(methodName string) jcrypto.SigningMethod {
	for _, v := range signingMethods {
		if strings.EqualFold(v.Alg(), methodName) {
			return v
		}
	}
	return jcrypto.SigningMethodHS256
}
