package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	gtemplate "text/template"

	log "github.com/Sirupsen/logrus"

	apitypes "github.com/emccode/libstorage/api/types"
	apiutils "github.com/emccode/libstorage/api/utils"
	"github.com/emccode/rexray/cli/cli/template"
)

type templateObject struct {
	C *CLI
	L apitypes.Client
	D interface{}
}

func (c *CLI) fmtOutput(w io.Writer, o interface{}) error {

	var (
		templName    string
		tabWriter    *tabwriter.Writer
		newTabWriter func()
		format       = c.outputFormat
		tplFormat    = c.outputTemplate
		tplTabs      = c.outputTemplateTabs
		tplBuf       = buildDefaultTemplate()
		funcMap      = gtemplate.FuncMap{
			"volumeStatus": c.volumeStatus,
		}
	)

	if tplTabs {
		newTabWriter = func() {
			tabWriter = tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			w = tabWriter
		}
		defer func() {
			if tabWriter != nil {
				tabWriter.Flush()
			}
		}()
	}

	switch {
	case strings.EqualFold(format, "json"):
		templName = templateNamePrintJSON
	case strings.EqualFold(format, "jsonp"):
		templName = templateNamePrintPrettyJSON
	case strings.EqualFold(format, "tmpl") && tplFormat == "":
		if tplTabs {
			newTabWriter()
		}
		switch to := o.(type) {
		case *apitypes.Volume:
			return c.fmtOutput(w, []*apitypes.Volume{to})
		case []*apitypes.Volume:
			templName = templateNamePrintVolumeFields
		case []*volumeWithPath:
			templName = templateNamePrintVolumeWithPathFields
		case *apitypes.Snapshot:
			return c.fmtOutput(w, []*apitypes.Snapshot{to})
		case []*apitypes.Snapshot:
			templName = templateNamePrintSnapshotFields
		case map[string]*apitypes.ServiceInfo:
			templName = templateNamePrintServiceFields
		case map[string]*apitypes.Instance:
			templName = templateNamePrintInstanceFields
		case []*apitypes.MountInfo:
			templName = templateNamePrintMountFields
		default:
			templName = templateNamePrintObject
		}
	default:
		if tplTabs {
			newTabWriter()
		}
		if tplFormat != "" {
			format = tplFormat
		}
		format = fmt.Sprintf(`"%s"`, strings.Replace(format, `"`, `\"`, -1))
		uf, err := strconv.Unquote(format)
		if err != nil {
			return err
		}
		format = uf
		templName = templateNamePrintCustom
	}

	if templName == templateNamePrintJSON ||
		templName == templateNamePrintPrettyJSON {

		templ := template.MustTemplate("json", tplBuf.String(), funcMap)
		return templ.ExecuteTemplate(w, templName, o)
	}

	if templName == templateNamePrintCustom {
		fmt.Fprintf(tplBuf, "{{define \"printCustom\"}}%s{{end}}", format)
	}

	if tplFormat == "" {
		tplMetadata, hasMetadata := defaultTemplates[templName]
		if hasMetadata {
			if fields := tplMetadata.fields; len(fields) > 0 {
				for i, f := range fields {
					fmt.Fprint(tplBuf, strings.SplitN(f, "=", 2)[0])
					if i < len(fields)-1 {
						fmt.Fprintf(tplBuf, "\t")
					}
				}
				fmt.Fprintf(tplBuf, "\n")
			}
		}
		fmt.Fprintf(tplBuf, "{{range ")
		if hasMetadata {
			if tplMetadata.sortBy == "" {
				fmt.Fprint(tplBuf, ".D")
			} else {
				fmt.Fprintf(tplBuf, "sort .D \"%s\"", tplMetadata.sortBy)
			}
		} else {
			fmt.Fprint(tplBuf, ".D")
		}
		fmt.Fprintf(tplBuf, " }}{{template \"%s\" .}}\n{{end}}", templName)
	} else {
		fmt.Fprintf(tplBuf, `{{template "%s" .}}`, templName)
	}

	format = tplBuf.String()
	c.ctx.WithField("template", format).Debug("built output template")

	templ, err := template.NewTemplate("tmpl", format, funcMap)
	if err != nil {
		return err
	}

	return templ.Execute(w, &templateObject{c, c.r, o})
}

func (c *CLI) marshalOutput(v interface{}) {
	if err := c.fmtOutput(os.Stdout, v); err != nil {
		log.Fatal(err)
	}
}

func (c *CLI) mustMarshalOutput(v interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}
	c.marshalOutput(v)
}

func (c *CLI) mustMarshalOutput3(v interface{}, noop interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}
	c.marshalOutput(v)
}

const (
	volumeStatusAttached    = "attached"
	volumeStatusAvailable   = "available"
	volumeStatusUnavailable = "unavailable"
	volumeStatusUnknown     = "unknown"
	volumeStatusError       = "error"
)

var voluemStatusStore = apiutils.NewStore()

func (c *CLI) volumeStatus(vol *apitypes.Volume) string {
	if len(vol.Attachments) == 0 {
		return volumeStatusAvailable
	}
	if c.r == nil || c.r.Executor() == nil {
		return volumeStatusUnknown
	}
	iid, err := c.r.Executor().InstanceID(c.ctx, voluemStatusStore)
	if err != nil {
		return volumeStatusError
	}
	for _, a := range vol.Attachments {
		if a.InstanceID.ID == iid.ID {
			return volumeStatusAttached
		}
	}
	return volumeStatusUnavailable
}

const (
	templateNamePrintCustom               = "printCustom"
	templateNamePrintObject               = "printObject"
	templateNamePrintJSON                 = "printJSON"
	templateNamePrintPrettyJSON           = "printPrettyJSON"
	templateNamePrintVolumeFields         = "printVolumeFields"
	templateNamePrintVolumeWithPathFields = "printVolumeWithPathFields"
	templateNamePrintSnapshotFields       = "printSnapshotFields"
	templateNamePrintInstanceFields       = "printInstanceFields"
	templateNamePrintServiceFields        = "printServiceFields"
	templateNamePrintMountFields          = "printMountFields"
)

type templateMetadata struct {
	format string
	fields []string
	sortBy string
}

var defaultTemplates = map[string]*templateMetadata{
	templateNamePrintObject: &templateMetadata{
		format: `{{printf "%v" .}}`,
	},
	templateNamePrintJSON: &templateMetadata{
		format: "{{. | json}}",
	},
	templateNamePrintPrettyJSON: &templateMetadata{
		format: "{{. | jsonp}}",
	},
	templateNamePrintVolumeFields: &templateMetadata{
		fields: []string{
			"ID",
			"Name",
			"Status={{. | volumeStatus}}",
			"Size",
		},
		sortBy: "Name",
	},
	templateNamePrintVolumeWithPathFields: &templateMetadata{
		fields: []string{
			"ID",
			"Name",
			"Status={{.Volume | volumeStatus}}",
			"Size",
			"Path",
		},
		sortBy: "Name",
	},
	templateNamePrintSnapshotFields: &templateMetadata{
		fields: []string{
			"ID",
			"Name",
			"Status",
			"VolumeID",
		},
		sortBy: "Name",
	},
	templateNamePrintInstanceFields: &templateMetadata{
		fields: []string{
			"ID={{.InstanceID.ID}}",
			"Name",
			"Provider={{.ProviderName}}",
			"Region",
		},
		sortBy: "Name",
	},
	templateNamePrintServiceFields: &templateMetadata{
		fields: []string{
			"Name",
			"Driver={{.Driver.Name}}",
		},
		sortBy: "Name",
	},
	templateNamePrintMountFields: &templateMetadata{
		fields: []string{
			"ID",
			"Device={{.Source}}",
			"MountPoint",
		},
		sortBy: "Source",
	},
}

func buildDefaultTemplate() *bytes.Buffer {
	buf := &bytes.Buffer{}
	for tplName, tplMetadata := range defaultTemplates {
		fmt.Fprintf(buf, `{{define "%s"}}`, tplName)
		if tplMetadata.format != "" {
			fmt.Fprintf(buf, "%s{{end}}", tplMetadata.format)
			continue
		}
		for i, field := range tplMetadata.fields {
			fieldParts := strings.SplitN(field, "=", 2)
			if len(fieldParts) == 1 {
				fmt.Fprintf(buf, "{{.%s}}", fieldParts[0])
			} else {
				fmt.Fprintf(buf, fieldParts[1])
			}
			if i < len(tplMetadata.fields)-1 {
				fmt.Fprintf(buf, "\t")
			}
		}
		fmt.Fprintf(buf, "{{end}}")
	}
	return buf
}
