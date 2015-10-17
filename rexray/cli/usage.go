package cli

import (
	"fmt"
	"io/ioutil"
	rx "regexp"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/emccode/rexray/util"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func initUsageTemplates() {

	var ut string
	utPath := fmt.Sprintf("%s/.rexray/usage.template", util.HomeDir())
	log.WithField("path", utPath).Debug("usage template path")

	if util.FileExists(utPath) {
		dat, err := ioutil.ReadFile(utPath)
		if err != nil {
			panic(err)
		}
		log.WithField("source", utPath).Debug("loaded usage template")
		ut = string(dat)
	} else {
		log.WithField("source", "UsageTemplate").Debug("loaded usage template")
		ut = usageTemplate
	}

	RexrayCmd.SetUsageTemplate(ut)
	RexrayCmd.SetHelpTemplate(ut)

	cobra.AddTemplateFuncs(template.FuncMap{
		"af":    additionalFlags,
		"hf":    hasFlags,
		"lf":    localFlags,
		"gf":    globalFlags,
		"ihf":   isHelpFlag,
		"ivf":   isVerboseFlag,
		"saf":   sansAdditionalFlags,
		"cmds":  commands,
		"rtrim": rtrim,
	})
}

func localFlags(cmd *cobra.Command) *flag.FlagSet {

	fs := &flag.FlagSet{}

	if cmd.HasParent() {
		cmd.LocalFlags().VisitAll(func(f *flag.Flag) {
			if f.Name != "help" {
				fs.AddFlag(f)
			}
		})
	} else {
		cmd.LocalFlags().VisitAll(func(f *flag.Flag) {
			if cmd.PersistentFlags().Lookup(f.Name) == nil {
				fs.AddFlag(f)
			}
		})
	}

	return sansAdditionalFlags(fs)
}

func globalFlags(cmd *cobra.Command) *flag.FlagSet {

	fs := &flag.FlagSet{}

	if cmd.HasParent() {
		fs.AddFlagSet(cmd.InheritedFlags())
		if fs.Lookup("help") == nil && cmd.Flag("help") != nil {
			fs.AddFlag(cmd.Flag("help"))
		}
	} else {
		fs.AddFlagSet(cmd.PersistentFlags())
	}

	return sansAdditionalFlags(fs)
}

func sansAdditionalFlags(flags *flag.FlagSet) *flag.FlagSet {
	fs := &flag.FlagSet{}
	flags.VisitAll(func(f *flag.Flag) {
		if r.Config.AdditionalFlags.Lookup(f.Name) == nil {
			fs.AddFlag(f)
		}
	})
	return fs
}

func hasFlags(flags *flag.FlagSet) bool {
	return flags != nil && flags.HasFlags()
}

func additionalFlags() *flag.FlagSet {
	return r.Config.AdditionalFlags
}

func isHelpFlag(cmd *cobra.Command) bool {
	v, e := cmd.Flags().GetBool("help")
	if e != nil {
		panic(e)
	}
	return v
}

func isVerboseFlag(cmd *cobra.Command) bool {
	v, e := cmd.Flags().GetBool("verbose")
	if e != nil {
		panic(e)
	}
	return v
}

func commands(cmd *cobra.Command) []*cobra.Command {
	if cmd.HasParent() {
		return cmd.Commands()
	}

	cArr := []*cobra.Command{}
	for _, c := range cmd.Commands() {
		if m, _ := rx.MatchString("((re)?start)|stop|status|((un)?install)", c.Name()); !m {
			cArr = append(cArr, c)
		}
	}
	return cArr
}

func rtrim(text string) string {
	return strings.TrimRight(text, " \n")
}

const usageTemplate = `{{$cmd := .}}{{with or .Long .Short }}{{. | trim}}{{end}}

Usage: {{if .Runnable}}
  {{.UseLine}}{{if .HasFlags}} [flags]{{end}}{{end}}{{if .HasSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases | rtrim}}{{end}}{{if .HasExample}}

Examples:
{{.Example | rtrim}}{{end}}{{ if .HasAvailableSubCommands}}

Available Commands: {{range cmds $cmd}}{{if (not .IsHelpCommand)}}
  {{rpad .Name .NamePadding }} {{.Short | rtrim}}{{end}}{{end}}{{end}}{{$lf := lf $cmd}}{{if hf $lf}}

Flags:
{{$lf.FlagUsages | rtrim}}{{end}}{{$gf := gf $cmd}}{{if hf $gf}}

Global Flags:
{{$gf.FlagUsages | rtrim}}{{end}}{{if ivf $cmd}}{{$af := af}}{{if hf $af}}

Additional Flags:
{{$af.FlagUsages | rtrim}}{{end}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics: {{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short | rtrim}}{{end}}}{{end}}{{end}}{{if .HasSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}

`
