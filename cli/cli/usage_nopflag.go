// +build !pflag

package usage

import flag "github.com/spf13/pflag"

func (c *CLI) additionalFlagSets() map[string]*flag.FlagSet {
	return nil
}
