package cli

import (
	"os"
	"testing"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
)

func newCLI(format, tpl string, tplTabs bool) *CLI {
	ctx := context.Background()
	if testing.Verbose() {
		context.SetLogLevel(ctx, log.DebugLevel)
	} else {
		context.SetLogLevel(ctx, log.WarnLevel)
	}
	return &CLI{
		ctx:                ctx,
		outputFormat:       format,
		outputTemplate:     tpl,
		outputTemplateTabs: tplTabs,
	}
}

func TestFormatEmptyVolumes(t *testing.T) {
	newCLI("tmpl", "", true).fmtOutput(os.Stdout, "", []*apitypes.Volume{})
}

func TestFormatVolume(t *testing.T) {
	newCLI("tmpl", "", true).fmtOutput(os.Stdout, "", &apitypes.Volume{
		ID:     "vol-1234",
		Name:   "abuilds",
		Size:   10240,
		Status: "attached",
	})
}

func TestFormatVolumes(t *testing.T) {
	newCLI("tmpl", "", true).fmtOutput(os.Stdout, "", []*apitypes.Volume{
		&apitypes.Volume{
			ID:     "vol-5",
			Name:   "bbuilds",
			Size:   10240,
			Status: "attached",
		},
		&apitypes.Volume{
			ID:   "vol-1234",
			Name: "abuilds",
			Size: 10240,
			Attachments: []*apitypes.VolumeAttachment{
				&apitypes.VolumeAttachment{
					MountPoint: "/data",
					InstanceID: &apitypes.InstanceID{ID: "test"},
				},
			},
		},
	})
}

func TestFormatVolumesCustom(t *testing.T) {
	newCLI("{{.ID}}", "", true).fmtOutput(os.Stdout, "", []*apitypes.Volume{
		&apitypes.Volume{
			ID:     "vol-5",
			Name:   "bbuilds",
			Size:   10240,
			Status: "attached",
		},
		&apitypes.Volume{
			ID:   "vol-1234",
			Name: "abuilds",
			Size: 10240,
		},
	})
}

func TestFormatVolumesCustomRaw(t *testing.T) {
	if err := newCLI(
		"tmpl",
		`{{range .D}}{{.Name}}\n{{end}}`,
		true).fmtOutput(
		os.Stdout, "", []*apitypes.Volume{
			&apitypes.Volume{
				ID:     "vol-5",
				Name:   "bbuilds",
				Size:   10240,
				Status: "attached",
			},
			&apitypes.Volume{
				ID:   "vol-1234",
				Name: "abuilds",
				Size: 10240,
			},
		}); err != nil {
		t.Fatal(err)
	}
	if err := newCLI(
		"tmpl",
		`{{with .D}}{{range .}}{{.Name}}\n{{end}}{{end}}`,
		true).fmtOutput(
		os.Stdout, "", []*apitypes.Volume{
			&apitypes.Volume{
				ID:     "vol-5",
				Name:   "bbuilds",
				Size:   10240,
				Status: "attached",
			},
			&apitypes.Volume{
				ID:   "vol-1234",
				Name: "abuilds",
				Size: 10240,
			},
		}); err != nil {
		t.Fatal(err)
	}
}

func TestFormatVolumesAsJSON(t *testing.T) {
	newCLI("json", "", true).fmtOutput(os.Stdout, "", []*apitypes.Volume{
		&apitypes.Volume{
			ID:     "vol-5",
			Name:   "bbuilds",
			Size:   10240,
			Status: "attached",
		},
		&apitypes.Volume{
			ID:   "vol-1234",
			Name: "abuilds",
			Size: 10240,
		},
	})
}

func TestFormatVolumesAsJSONP(t *testing.T) {
	newCLI("jsonp", "", true).fmtOutput(os.Stdout, "", []*apitypes.Volume{
		&apitypes.Volume{
			ID:     "vol-5",
			Name:   "bbuilds",
			Size:   10240,
			Status: "attached",
		},
		&apitypes.Volume{
			ID:   "vol-1234",
			Name: "abuilds",
			Size: 10240,
		},
	})
}

func TestTemplateDefs(t *testing.T) {
	tpl := template.Must(template.New("temp").Parse(`{{define "T1"}}ONE{{end}}
{{define "T2"}}TWO{{end}}
{{define "T3"}}{{template "T1"}} {{template "T2"}}{{end}}
{{define "T4"}}{{template "T1"}} {{template "T2"}}{{end}}`))
	tpl.ExecuteTemplate(os.Stdout, "T3", nil)
	tpl.ExecuteTemplate(os.Stdout, "T4", nil)
}
