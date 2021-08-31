package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	gtemplate "text/template"

	log "github.com/sirupsen/logrus"

	"github.com/AVENTER-UG/rexray/cli/template"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
)

type templateObject struct {
	C *CLI
	L apitypes.Client
	D interface{}
}

func (c *CLI) fmtOutput(w io.Writer, templateName string, o interface{}) error {

	var (
		tabWriter    *tabwriter.Writer
		newTabWriter func()
		templName    = templateName
		format       = c.outputFormat
		tplFormat    = c.outputTemplate
		tplTabs      = c.outputTemplateTabs
		tplBuf       = buildDefaultTemplate()
		funcMap      = gtemplate.FuncMap{
			"printAttState": func(s apitypes.VolumeAttachmentStates) string {
				switch s {
				case apitypes.VolumeAttached:
					return "attached"
				case apitypes.VolumeAvailable:
					return "available"
				case apitypes.VolumeUnavailable:
					return "unavailable"
				default:
					return "unknown"
				}
			},
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
			return c.fmtOutput(w, templName, []*apitypes.Volume{to})
		case []*apitypes.Volume:
			isEnc := false
			for _, v := range to {
				if v.Encrypted {
					isEnc = true
					break
				}
			}
			if templName == "" {
				if isEnc {
					templName = templateNamePrintEncVolumeFields
				} else {
					templName = templateNamePrintVolumeFields
				}
			}
		case []*volumeWithPath:
			isEnc := false
			for _, v := range to {
				if v.Encrypted {
					isEnc = true
					break
				}
			}
			if templName == "" {
				if isEnc {
					templName = templateNamePrintEncVolumeWithPathFields
				} else {
					templName = templateNamePrintVolumeWithPathFields
				}
			}
		case *apitypes.Snapshot:
			return c.fmtOutput(w, templName, []*apitypes.Snapshot{to})
		case []*apitypes.Snapshot:
			if templName == "" {
				templName = templateNamePrintSnapshotFields
			}
		case map[string]*apitypes.ServiceInfo:
			if templName == "" {
				templName = templateNamePrintServiceFields
			}
		case map[string]*apitypes.Instance:
			if templName == "" {
				templName = templateNamePrintInstanceFields
			}
		case []*apitypes.MountInfo:
			if templName == "" {
				templName = templateNamePrintMountFields
			}
		case *scriptInfo:
			return c.fmtOutput(w, templName, []*scriptInfo{to})
		case []*scriptInfo:
			if templName == "" {
				templName = templateNamePrintScriptFields
			}
		case []string:
			sort.Strings(to)
			if templName == "" {
				templName = templateNamePrintStringSlice
			}
		case []*authToken:
			if templName == "" {
				templName = templateNamePrintAuthTokenFields
			}
		default:
			if templName == "" {
				templName = templateNamePrintObject
			}
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
			if !c.quiet {
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
	c.marshalOutputWithTemplateName("", v)
}

func (c *CLI) marshalOutputWithTemplateName(
	templateName string, v interface{}) {

	if err := c.fmtOutput(os.Stdout, templateName, v); err != nil {
		log.Fatal(err)
	}
}

func (c *CLI) mustMarshalOutput(v interface{}, err error) {
	c.mustMarshalOutputWithTemplateName("", v, err)
}

func (c *CLI) mustMarshalOutputWithTemplateName(
	templateName string, v interface{}, err error) {

	if err != nil {
		log.Fatal(err)
	}
	switch tv := v.(type) {
	case *lsVolumesResult:
		c.marshalOutputWithTemplateName(templateName, tv.vols)
	default:
		c.marshalOutputWithTemplateName(templateName, v)
	}
}

func (c *CLI) mustMarshalOutput3(v, noop interface{}, err error) {
	c.mustMarshalOutput3WithTemplateName("", v, nil, err)
}

func (c *CLI) mustMarshalOutput3WithTemplateName(
	templateName string, v, noop interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}
	c.marshalOutputWithTemplateName(templateName, v)
}

const (
	templateNamePrintCustom                  = "printCustom"
	templateNamePrintObject                  = "printObject"
	templateNamePrintStringSlice             = "printStringSlice"
	templateNamePrintJSON                    = "printJSON"
	templateNamePrintPrettyJSON              = "printPrettyJSON"
	templateNamePrintVolumeFields            = "printVolumeFields"
	templateNamePrintVolumeID                = "printVolumeID"
	templateNamePrintVolumeWithPathFields    = "printVolumeWithPathFields"
	templateNamePrintEncVolumeFields         = "printEncVolumeFields"
	templateNamePrintEncVolumeWithPathFields = "printEncVolumeWithPathFields"
	templateNamePrintSnapshotFields          = "printSnapshotFields"
	templateNamePrintInstanceFields          = "printInstanceFields"
	templateNamePrintServiceFields           = "printServiceFields"
	templateNamePrintMountFields             = "printMountFields"
	templateNamePrintScriptFields            = "printScriptFields"
	templateNamePrintAuthTokenFields         = "printAuthTokenFields"
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
	templateNamePrintStringSlice: &templateMetadata{
		format: `{{.}}`,
	},
	templateNamePrintJSON: &templateMetadata{
		format: "{{. | json}}",
	},
	templateNamePrintPrettyJSON: &templateMetadata{
		format: "{{. | jsonp}}",
	},
	templateNamePrintVolumeID: &templateMetadata{
		fields: []string{"ID"},
		sortBy: "ID",
	},
	templateNamePrintVolumeFields: &templateMetadata{
		fields: []string{
			"ID",
			"Name",
			"Status={{.AttachmentState | printAttState}}",
			"Size",
		},
		sortBy: "Name",
	},
	templateNamePrintVolumeWithPathFields: &templateMetadata{
		fields: []string{
			"ID",
			"Name",
			"Status={{.Volume.AttachmentState | printAttState}}",
			"Size",
			"Path",
		},
		sortBy: "Name",
	},
	templateNamePrintEncVolumeFields: &templateMetadata{
		fields: []string{
			"ID",
			"Name",
			"Status={{.AttachmentState | printAttState}}",
			"Size",
			"Encrypted",
		},
		sortBy: "Name",
	},
	templateNamePrintEncVolumeWithPathFields: &templateMetadata{
		fields: []string{
			"ID",
			"Name",
			"Status={{.Volume.AttachmentState | printAttState}}",
			"Size",
			"Encrypted",
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
	templateNamePrintScriptFields: &templateMetadata{
		fields: []string{
			"Path",
			"Installed",
			"Modified",
		},
		sortBy: "Path",
	},
	templateNamePrintAuthTokenFields: &templateMetadata{
		fields: []string{
			"Subject",
			"IssuedAt",
			"NotBefore",
			"Expires",
		},
		sortBy: "Subject",
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
