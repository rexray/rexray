// +build pflag

package cli

import flag "github.com/spf13/pflag"

func (c *CLI) additionalFlagSets() map[string]*flag.FlagSet {
	afs := map[string]*flag.FlagSet{}
	for fsn, fs := range c.config.FlagSets() {
		if fsn == "Global Flags" || !fs.HasFlags() {
			continue
		}

		afs[fsn] = fs
	}
	return afs
}
