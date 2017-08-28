// +build pflag

package cli

func init() {
	initConfigFlagsFunc = func(c *CLI) {
		for _, fs := range c.config.FlagSets() {
			c.c.PersistentFlags().AddFlagSet(fs)
		}
	}
}
