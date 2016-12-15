// +build pflag

package cli

import flag "github.com/spf13/pflag"

func init() {
	additionalFlagSetsFunc = func(c *CLI) map[string]*flag.FlagSet {
		afs := map[string]*flag.FlagSet{}
		for fsn, fs := range c.config.FlagSets() {
			if fsn == "Global Flags" || !fs.HasFlags() {
				continue
			}
			afs[fsn] = fs
		}
		return afs
	}
}
