// +build !agent
// +build !controller

package cli

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	"github.com/spf13/cobra"

	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/scripts"
	"github.com/AVENTER-UG/rexray/util"
)

func init() {
	initCmdFuncs = append(initCmdFuncs, func(c *CLI) {
		c.initScriptsCmd()
		c.initScriptsFlags()
	})
}

func (c *CLI) initScriptsCmd() {
	c.scriptsCmd = &cobra.Command{
		Use:   "scripts",
		Short: "The REX-Ray script manager",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	c.c.AddCommand(c.scriptsCmd)

	c.scriptsListCmd = &cobra.Command{
		Use:     "list",
		Short:   "List the available scripts",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			c.mustMarshalOutput(c.lsScripts(c.ctx))
		},
	}
	c.scriptsCmd.AddCommand(c.scriptsListCmd)

	c.scriptsInstallCmd = &cobra.Command{
		Use:     "install",
		Short:   "Install one or more script(s)",
		Aliases: []string{"i"},
		Example: util.BinFileName + " scripts install " +
			"[USER:REPO:]NAME[:COMMIT]|GIST|URL [[USER:REPO:]NAME[:COMMIT]|GIST|URL...]",
		Run: func(cmd *cobra.Command, args []string) {
			c.mustMarshalOutput(c.installScripts(c.ctx, args...))
		},
	}
	c.scriptsCmd.AddCommand(c.scriptsInstallCmd)

	c.scriptsUninstallCmd = &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstalls one or more script(s)",
		Aliases: []string{"u"},
		Example: util.BinFileName + " scripts uninstall NAME [NAME...]",
		Run: func(cmd *cobra.Command, args []string) {
			c.mustMarshalOutput(c.uninstallScripts(c.ctx, args...))
		},
	}
	c.scriptsCmd.AddCommand(c.scriptsUninstallCmd)
}

const (
	scriptsDirName = "scripts"
)

func (c *CLI) initScriptsFlags() {}

type fileInfoEx interface {
	os.FileInfo
	MD5Checksum() string
}

func getLocalScriptInfo(
	ctx apitypes.Context, aix fileInfoEx) (bool, bool, error) {

	filePath := util.ScriptFilePath(ctx, aix.Name())
	if !gotil.FileExists(filePath) {
		return false, false, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return false, false, err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return true, false, err
	}

	md5checksum := fmt.Sprintf("%x", h.Sum(nil))
	if md5checksum != aix.MD5Checksum() {
		return true, true, nil
	}

	return true, false, nil
}

var (
	gistIDRX         = regexp.MustCompile(`^[\w\d]{32}$`)
	startsWithHTTPRX = regexp.MustCompile(`(?i)^https?://.+`)
)

func (c *CLI) installScripts(
	ctx apitypes.Context, args ...string) ([]*scriptInfo, error) {

	sia := []*scriptInfo{}
	for _, v := range args {
		ctx.WithField("scriptName", v).Debug("getting script asset")
		// is it an embedded script?
		ai, err := scripts.AssetInfo(v)
		if err == nil {
			si, err := c.installEmbeddedScript(ctx, ai)
			if err != nil {
				return nil, err
			}
			sia = append(sia, si)
			continue
		}

		// is it a gist id?
		if gistIDRX.MatchString(v) {
			si, err := c.installGist(ctx, v)
			if err != nil {
				return nil, err
			}
			sia = append(sia, si)
			continue
		}

		// is it a url?
		if startsWithHTTPRX.MatchString(v) {
			if u, err := url.Parse(v); err == nil {
				si, err := c.installURL(ctx, u)
				if err != nil {
					return nil, err
				}
				sia = append(sia, si)
				continue
			}
		}

		// treat it as a file path on github
		si, err := c.installGitHub(ctx, v)
		if err != nil {
			return nil, err
		}
		sia = append(sia, si)
	}
	return sia, nil
}

func (c *CLI) installEmbeddedScript(
	ctx apitypes.Context,
	info os.FileInfo) (*scriptInfo, error) {

	fp := util.ScriptFilePath(ctx, info.Name())
	if err := ioutil.WriteFile(
		fp,
		scripts.MustAsset(info.Name()),
		os.FileMode(0755)); err != nil {
		return nil, err
	}
	return &scriptInfo{
		Name:      info.Name(),
		Path:      fp,
		Installed: true,
		Modified:  false,
	}, nil
}

func (c *CLI) installGist(
	ctx apitypes.Context,
	id string) (*scriptInfo, error) {

	name, rdr, err := scripts.GetGist(c.ctx, id, "")
	if err != nil {
		return nil, err
	}
	defer rdr.Close()
	fp := util.ScriptFilePath(ctx, name)
	f, err := os.Create(fp)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(f, rdr); err != nil {
		return nil, err
	}
	return &scriptInfo{
		Name:      name,
		Path:      fp,
		Installed: true,
		Modified:  false,
	}, nil
}

func (c *CLI) installURL(
	ctx apitypes.Context,
	u *url.URL) (*scriptInfo, error) {

	rdr, err := scripts.GetHTTP(c.ctx, u.String())
	if err != nil {
		return nil, err
	}
	defer rdr.Close()
	name := path.Base(u.Path)
	fp := util.ScriptFilePath(ctx, name)
	f, err := os.Create(fp)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(f, rdr); err != nil {
		return nil, err
	}
	return &scriptInfo{
		Name:      name,
		Path:      fp,
		Installed: true,
		Modified:  false,
	}, nil
}

// [USER:REPO:]NAME[:COMMIT]
func (c *CLI) installGitHub(
	ctx apitypes.Context,
	v string) (*scriptInfo, error) {

	var (
		err  error
		rdr  io.ReadCloser
		p    = strings.Split(v, ":")
		name string
	)
	switch len(p) {
	case 1:
		name = p[0]
		rdr, err = scripts.GetGitHubBlob(c.ctx, "", "", "", name)
	case 2:
		name = p[0]
		rdr, err = scripts.GetGitHubBlob(c.ctx, "", "", p[1], name)
	case 3:
		name = p[2]
		rdr, err = scripts.GetGitHubBlob(c.ctx, p[0], p[1], p[3], name)
	case 4:
		name = p[2]
		rdr, err = scripts.GetGitHubBlob(c.ctx, p[0], p[1], p[3], name)
	default:
		return nil, errors.New("invalid argument")
	}
	if err != nil {
		return nil, err
	}
	fp := util.ScriptFilePath(ctx, name)
	f, err := os.Create(fp)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(f, rdr); err != nil {
		return nil, err
	}
	return &scriptInfo{
		Name:      name,
		Path:      fp,
		Installed: true,
		Modified:  false,
	}, nil
}

func (c *CLI) uninstallScripts(
	ctx apitypes.Context,
	args ...string) ([]string, error) {

	rmd := []string{}
	for _, v := range args {
		fp := util.ScriptFilePath(ctx, v)
		os.RemoveAll(fp)
		rmd = append(rmd, v)
	}
	return rmd, nil
}

func (c *CLI) lsScripts(ctx apitypes.Context) ([]*scriptInfo, error) {
	sia := []*scriptInfo{}
	for _, an := range scripts.AssetNames() {
		ai, err := scripts.AssetInfo(an)
		if err != nil {
			return nil, err
		}
		aix, ok := ai.(fileInfoEx)
		if !ok {
			return nil, goof.WithField("name", an, "invalid fileInfoEx")
		}
		installed, modified, err := getLocalScriptInfo(ctx, aix)
		if err != nil {
			return nil, err
		}
		si := &scriptInfo{
			Name:      path.Base(aix.Name()),
			Path:      util.ScriptFilePath(ctx, aix.Name()),
			Installed: installed,
			Modified:  modified,
		}
		sia = append(sia, si)
	}
	return sia, nil
}
