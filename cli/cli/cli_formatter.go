package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	gtemplate "text/template"

	log "github.com/Sirupsen/logrus"

	apitypes "github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/rexray/cli/cli/template"
)

type templateObject struct {
	C *CLI
	L apitypes.Client
	D interface{}
}

func (c *CLI) fmtOutput(ch chan interface{}) error {
	if ch == nil {
		return nil
	}
	errs := make(chan error)
	if strings.EqualFold(c.outputFormat, "json") ||
		strings.EqualFold(c.outputFormat, "jsonp") {
		go c.streamOutputJSON(ch, errs)
	} else {
		go c.streamOutput(ch, errs)
	}
	if err := <-errs; err != nil {
		return err
	}
	return nil
}

func (c *CLI) streamOutput(ch chan interface{}, errs chan error) {

	var (
		twMinWidthStr  = os.Getenv("REXRAY_CLI_TABWRITER_MINWIDTH")
		twMaxWidthStr  = os.Getenv("REXRAY_CLI_TABWRITER_MAXWIDTH")
		twTabWidthStr  = os.Getenv("REXRAY_CLI_TABWRITER_TABWIDTH")
		twPaddingStr   = os.Getenv("REXRAY_CLI_TABWRITER_PADDING")
		twPadCharStr   = os.Getenv("REXRAY_CLI_TABWRITER_PADCHAR")
		twFlagsStr     = os.Getenv("REXRAY_CLI_TABWRITER_FLAGS")
		twAutoFlushStr = os.Getenv("REXRAY_CLI_TABWRITER_AUTOFLUSH")

		twMinWith        = 0
		twMaxWidth       = 15
		twTabWidth       = 0
		twPadding        = 2
		twPadChar   byte = ' '
		twFlags          = tabwriter.DiscardEmptyColumns
		twAutoFlush      = false
	)

	if twMinWidthStr != "" {
		if i, err := strconv.ParseInt(twMinWidthStr, 10, 64); err == nil {
			twMinWith = int(i)
		}
	}
	if twMaxWidthStr != "" {
		if i, err := strconv.ParseInt(twMaxWidthStr, 10, 64); err == nil {
			twMaxWidth = int(i)
		}
	}
	if twTabWidthStr != "" {
		if i, err := strconv.ParseInt(twTabWidthStr, 10, 64); err == nil {
			twTabWidth = int(i)
		}
	}
	if twPaddingStr != "" {
		if i, err := strconv.ParseInt(twPaddingStr, 10, 64); err == nil {
			twPadding = int(i)
		}
	}
	if twPadCharStr != "" {
		if i, err := strconv.ParseInt(twPadCharStr, 10, 8); err == nil {
			twPadChar = byte(i)
		}
	}
	if twFlagsStr != "" {
		if i, err := strconv.ParseUint(twFlagsStr, 10, 64); err == nil {
			twFlags = uint(i)
		}
	}
	if twAutoFlushStr != "" {
		twAutoFlush, _ = strconv.ParseBool(twAutoFlushStr)
	}

	ellipses := func(o interface{}) interface{} {
		var s string
		switch to := o.(type) {
		case string:
			s = to
		default:
			s = fmt.Sprintf("%v", o)
		}
		if len(s) == 0 {
			return o
		}
		if diff := len(s) - twMaxWidth; diff > 0 {
			idx := len(s) - (3 + diff)
			if idx > len(s) {
				idx = len(s) - 1
			}
			if idx < 0 {
				idx = 0
			}
			return fmt.Sprintf("%s...", s[0:idx])
		}
		return o
	}

	var (
		w = tabwriter.NewWriter(
			os.Stdout,
			twMinWith,
			twTabWidth,
			twPadding,
			twPadChar,
			twFlags)
		format  = c.outputFormat
		funcMap = gtemplate.FuncMap{
			"volumeStatus": c.volumeStatus,
			"ellipses":     ellipses,
		}
	)

	flush := w.Flush
	defer flush()

	writeItem := func(format string, o interface{}) error {
		tpl, err := template.NewTemplate("tpl", format, funcMap)
		if err != nil {
			return err
		}
		if err := tpl.Execute(w, o); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if twAutoFlush {
			return flush()
		}
		return nil
	}

	if strings.EqualFold(format, "tmpl") {
		printedHeaders := false
		printHeaders := func(fields string) error {
			if c.quiet || printedHeaders {
				return nil
			}
			printedHeaders = true
			if fields == "" {
				return nil
			}
			if _, err := fmt.Fprintf(w, "%s\n", fields); err != nil {
				return err
			}
			return nil
		}

		printHeadersAndItem := func(ti *templateInfo, o interface{}) error {
			if err := printHeaders(ti.fields); err != nil {
				errs <- err
				close(errs)
				return err
			}
			if err := writeItem(ti.format, o); err != nil {
				errs <- err
				close(errs)
				return err
			}
			return nil
		}

		for o := range ch {
			switch to := o.(type) {
			case *apitypes.Volume:
				ti := defaultTemplateInfo[templateInfoVolume]
				if err := printHeadersAndItem(ti, o); err != nil {
					return
				}
			case *volumeWithPath:
				ti := defaultTemplateInfo[templateInfoVolumeWithPath]
				if err := printHeadersAndItem(ti, o); err != nil {
					return
				}
			case *apitypes.Snapshot:
				ti := defaultTemplateInfo[templateInfoSnapshot]
				if err := printHeadersAndItem(ti, o); err != nil {
					return
				}
			case *apitypes.ServiceInfo:
				ti := defaultTemplateInfo[templateInfoService]
				if err := printHeadersAndItem(ti, o); err != nil {
					return
				}
			case *apitypes.Instance:
				ti := defaultTemplateInfo[templateInfoInstance]
				if err := printHeadersAndItem(ti, o); err != nil {
					return
				}
			case *apitypes.MountInfo:
				ti := defaultTemplateInfo[templateInfoMount]
				if err := printHeadersAndItem(ti, o); err != nil {
					return
				}
			case error:
				_, err := fmt.Fprintf(os.Stderr, "%s\n", to.Error())
				if err != nil {
					errs <- err
					close(errs)
					return
				}
			case func() error:
				if err := to(); err != nil {
					errs <- err
					close(errs)
					return
				}
			default:
				ti := defaultTemplateInfo[templateInfoObject]
				if err := writeItem(ti.format, o); err != nil {
					errs <- err
					close(errs)
					return
				}
			}
		}
	} else {
		format = fmt.Sprintf(`"%s"`, strings.Replace(format, `"`, `\"`, -1))
		uf, err := strconv.Unquote(format)
		if err != nil {
			errs <- err
			close(errs)
			return
		}
		format = uf
		for o := range ch {
			if err := writeItem(format, o); err != nil {
				errs <- err
				close(errs)
				return
			}
		}
	}

	close(errs)
}

func (c *CLI) streamOutputJSON(ch chan interface{}, errs chan error) {
	var w io.Writer = os.Stdout
	pretty := strings.EqualFold(c.outputFormat, "jsonp")
	enc := func(o interface{}) ([]byte, error) {
		var (
			buf []byte
			err error
		)
		if pretty {
			buf, err = json.MarshalIndent(o, "  ", "  ")
		} else {
			buf, err = json.Marshal(o)
		}
		if err != nil {
			return nil, err
		}
		return buf, nil
	}
	fmt.Fprint(w, "[")
	i := 0
	for o := range ch {
		if i == 0 {
			if pretty {
				fmt.Fprint(w, "\n  ")
			}
		}
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		buf, err := enc(o)
		if err != nil {
			errs <- err
			close(errs)
			return
		}
		w.Write(buf)
		i++
	}
	fmt.Fprint(w, "]")
	close(errs)
}

func (c *CLI) marshalOutput(ch chan interface{}) {
	if ch == nil {
		return
	}
	if err := c.fmtOutput(ch); err != nil {
		log.Fatal(err)
	}
}

type templateName string

const (
	templateInfoCustom         templateName = "printCustom"
	templateInfoObject         templateName = "printObject"
	templateInfoVolume         templateName = "printVolume"
	templateInfoVolumeWithPath templateName = "printVolumeWithPath"
	templateInfoSnapshot       templateName = "printSnapshot"
	templateInfoInstance       templateName = "printInstance"
	templateInfoService        templateName = "printService"
	templateInfoMount          templateName = "printMount"
)

type templateInfo struct {
	format string
	fields string
}

var defaultTemplateInfo = map[templateName]*templateInfo{
	templateInfoObject: &templateInfo{
		format: `{{printf "%v" .}}`,
	},
	templateInfoVolume: &templateInfo{
		format: "{{.ID|ellipses}}\t{{.Name|ellipses}}\t" +
			"{{.|volumeStatus|ellipses}}\t{{.Size|ellipses}}",
		fields: "ID\tName\tStatus\tSize\tPath",
	},
	templateInfoVolumeWithPath: &templateInfo{
		format: "{{.ID|ellipses}}\t{{.Name|ellipses}}\t" +
			"{{.|volumeStatus|ellipses}}\t{{.Size|ellipses}}\t{{.Path|ellipses}}",
		fields: "ID\tName\tStatus\tSize\tPath",
	},
	templateInfoSnapshot: &templateInfo{
		format: "{{.ID|ellipses}}\t{{.Name|ellipses}}\t" +
			"{{.Status|ellipses}}\t{{.VolumeID|ellipses}}",
		fields: "ID\tName\tStatus\tStatus\tVolumeID",
	},
	templateInfoInstance: &templateInfo{
		format: "{{.InstanceID.ID|ellipses}}\t{{.Name|ellipses}}\t" +
			"{{.ProviderName|ellipses}}\t{{.Region|ellipses}}",
		fields: "ID\tName\tStatus\tProvider\tRegion",
	},
	templateInfoService: &templateInfo{
		format: "{{.Name|ellipses}}\t{{.Driver.Name|ellipses}}",
		fields: "Name\tDriver",
	},
	templateInfoMount: &templateInfo{
		format: "{{.ID|ellipses}}\t{{.Source|ellipses}}\t" +
			"{{.MountPoint|ellipses}}",
		fields: "ID\tSource\tMountPoint",
	},
}
