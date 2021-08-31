package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	jcrypto "github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/akutz/gotil"
	"github.com/AVENTER-UG/rexray/util"
	"github.com/spf13/cobra"
)

const (
	lsSrvrAuthKey = "libstorage.server.auth.key"
	rrTokKeyEnv   = "REXRAY_TOKEN_KEY"
	rrTokKeyKey   = "rexray.token.key"

	lsSrvrAuthAlg = "libstorage.server.auth.alg"
	rrTokAlgEnv   = "REXRAY_TOKEN_ALG"
	rrTokAlgKey   = "rexray.token.alg"

	rrTokExpEnv = "REXRAY_TOKEN_EXP"
	rrTokExpKey = "rexray.token.exp"

	// 8760 hours is 365 days
	defaultTokenExp = time.Duration(8760 * time.Hour)
)

var (
	zd = time.Duration(0)
)

func init() {
	initCmdFuncs = append(initCmdFuncs, func(c *CLI) {
		c.initTokenCmds()
		c.initTokenFlags()
	})
}

func (c *CLI) initTokenCmds() {
	c.tokenCmd = &cobra.Command{
		Use:   "token",
		Short: "The token manager",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	c.c.AddCommand(c.tokenCmd)

	c.tokenNewCmd = &cobra.Command{
		Use:     "new",
		Aliases: []string{"create", "c", "n"},
		Short:   "Create a new token",
		Example: util.BinFileName + " token new [OPTIONS] SUBJECT [SUBJECT...]",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Usage()
				os.Exit(1)
			}
			key, alg := c.tokenValidateKeyAndAlg()
			exp := c.tokenValidateExp()
			now := time.Now().UTC()
			for _, a := range args {
				clm := jws.Claims{}
				if exp != zd {
					clm.SetExpiration(now.Add(exp))
				}
				clm.SetIssuedAt(now)
				clm.SetNotBefore(now)
				clm.SetSubject(a)
				jwt := jws.NewJWT(clm, alg)
				buf, err := jwt.Serialize(key)
				if err != nil {
					c.ctx.WithError(err).Fatal("error serializing token")
				}
				fmt.Println(string(buf))
			}
		},
	}
	c.tokenCmd.AddCommand(c.tokenNewCmd)

	c.tokenDecodeCmd = &cobra.Command{
		Use:     "decode",
		Aliases: []string{"dec", "d"},
		Short:   "Decodes a token",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				// is there stdin?
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					rdr := bufio.NewScanner(os.Stdin)
					for rdr.Scan() {
						args = append(args, rdr.Text())
					}
				} else {
					cmd.Usage()
					os.Exit(1)
				}
			}
			var (
				key  []byte
				alg  jcrypto.SigningMethod
				toks []*authToken
			)
			if c.verify {
				key, alg = c.tokenValidateKeyAndAlg()
			}
			for _, a := range args {
				jwt, err := jws.ParseJWT([]byte(a))
				if err != nil {
					c.ctx.WithError(err).Fatal("error: invalid token")
				}
				if c.verify {
					fmt.Printf("key=%v\n", string(key))
					fmt.Printf("alg=%v\n", alg.Alg())
					if err := jwt.Validate(key, alg); err != nil {
						c.ctx.WithError(err).Fatal(
							"error: invalid token signature")
					}
				}
				t := &authToken{}
				if v, ok := jwt.Claims().Subject(); ok {
					t.Subject = v
				}
				if v, ok := jwt.Claims().Expiration(); ok {
					t.Expires = v
				}
				if v, ok := jwt.Claims().IssuedAt(); ok {
					t.IssuedAt = v
				}
				if v, ok := jwt.Claims().NotBefore(); ok {
					t.NotBefore = v
				}
				toks = append(toks, t)
			}
			c.mustMarshalOutput(toks, nil)
		},
	}
	c.tokenCmd.AddCommand(c.tokenDecodeCmd)

}

func (c *CLI) tokenValidateKeyAndAlg() ([]byte, jcrypto.SigningMethod) {
	var keyBuf []byte
	if c.key == "" {
		if c.key = os.Getenv(rrTokKeyEnv); c.key == "" {
			if c.key = c.config.GetString(rrTokKeyKey); c.key == "" {
				c.key = c.config.GetString(lsSrvrAuthKey)
				if c.key == "" {
					c.ctx.Fatal("error: missing token key")
				}
			}
		}
	}
	if gotil.FileExists(c.key) {
		buf, err := ioutil.ReadFile(c.key)
		if err != nil {
			c.ctx.WithField("keyFilePath", c.key).WithError(err).Fatal(
				"error reading key file")
		}
		keyBuf = buf
	} else {
		keyBuf = []byte(c.key)
	}
	if c.alg == "" {
		if c.alg = os.Getenv(rrTokAlgEnv); c.alg == "" {
			if c.alg = c.config.GetString(rrTokAlgKey); c.alg == "" {
				c.alg = c.config.GetString(lsSrvrAuthAlg)
				if c.alg == "" {
					c.ctx.Fatal("error: missing token alg")
				}
			}
		}
	}
	alg, ok := parseSigningMethod(c.alg)
	if !ok {
		c.ctx.Fatal("error: invalid token alg")
	}
	return keyBuf, alg
}

func (c *CLI) tokenValidateExp() time.Duration {
	var (
		err error
		exp = c.expires
	)
	if exp == zd {
		if v := os.Getenv(rrTokExpEnv); v != "" {
			if exp, err = time.ParseDuration(v); err != nil {
				c.ctx.Fatal("error: invalid token exp")
			}
			return exp
		}
		if v := c.config.GetString(rrTokExpKey); v != "" {
			if exp, err = time.ParseDuration(v); err != nil {
				c.ctx.Fatal("error: invalid token exp")
			}
			return exp
		}
		c.ctx.WithField("exp", defaultTokenExp).Info(
			"Using default token expiration duration")
		exp = defaultTokenExp
	}
	return exp
}

func (c *CLI) initTokenFlags() {
	c.tokenNewCmd.Flags().StringVar(&c.key, "key", "",
		"The key used to sign the token.")
	c.tokenNewCmd.Flags().StringVar(&c.alg, "alg", "HS256",
		"The algorithm used to sign the token.")
	c.tokenNewCmd.Flags().DurationVar(&c.expires, "exp",
		zd,
		"The period of time until the token expires. Valid time units are "+
			`"ns", "us" (or "µs"), "ms", "s", "m", "h".`)

	c.tokenDecodeCmd.Flags().StringVar(&c.key, "key", "",
		"The key used to verify the token's signature")
	c.tokenDecodeCmd.Flags().BoolVar(&c.verify, "verify", false,
		"A flag indicating whether or not to verify the token's signature. "+
			"If true and the signature is invalid or cannot be validated the "+
			"token will not be decoded.")
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
func parseSigningMethod(methodName string) (jcrypto.SigningMethod, bool) {
	for _, v := range signingMethods {
		if strings.EqualFold(v.Alg(), methodName) {
			return v, true
		}
	}
	return nil, false
}

type authToken struct {
	Subject   string    `json:"sub"`
	Expires   time.Time `json:"exp"`
	NotBefore time.Time `json:"nbf"`
	IssuedAt  time.Time `json:"iat"`
}
