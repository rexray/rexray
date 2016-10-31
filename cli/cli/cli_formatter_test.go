package cli

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/codedellemc/libstorage/api/context"
	apitypes "github.com/codedellemc/libstorage/api/types"
)

var r = rand.New(rand.NewSource(99))

func newCLI(format string) *CLI {
	ctx := context.Background()
	if testing.Verbose() {
		context.SetLogLevel(ctx, log.DebugLevel)
	} else {
		context.SetLogLevel(ctx, log.WarnLevel)
	}
	return &CLI{
		ctx:          ctx,
		outputFormat: format,
	}
}

func TestFormatEmptyVolumes(t *testing.T) {
	if err := newCLI("tmpl").fmtOutput(nil); err != nil {
		t.Fatal(err)
	}
}

func TestFormatVolume(t *testing.T) {
	ch := make(chan interface{})
	go func() {
		ch <- &apitypes.Volume{
			ID:     "vol-1234",
			Name:   "abuilds",
			Size:   10240,
			Status: "attached",
		}
		close(ch)
	}()
	if err := newCLI("tmpl").fmtOutput(ch); err != nil {
		t.Fatal(err)
	}
}

func newVolumeChannel() chan interface{} {
	return newVolumeChannelWithError(false)
}

func newVolumeChannelWithError(withError bool) chan interface{} {
	ch := make(chan interface{})
	go func() {
		for i := 0; i < 5; i++ {
			if withError && i == 2 {
				ch <- fmt.Errorf("error queryig volume vol-2")
				continue
			}

			vol := &apitypes.Volume{
				ID:     fmt.Sprintf("vol-%d", i),
				Name:   fmt.Sprintf("builds-%d", i),
				Size:   10240,
				Status: "attached",
			}
			if i == 3 {
				vol.Name = fmt.Sprintf("builds-%d-abcdefghik", i)
			}
			if r.Intn(1) == 1 {
				vol.Attachments = []*apitypes.VolumeAttachment{
					&apitypes.VolumeAttachment{
						MountPoint: fmt.Sprintf(
							"/var/lib/rexray/vol%d/data", i),
						InstanceID: &apitypes.InstanceID{ID: "test"},
					},
				}
			}
			if i == 2 {
				ch <- &volumeWithPath{vol, ""}
			} else {
				ch <- vol
			}
			time.Sleep(time.Duration(r.Intn(3)) * time.Second)
		}
		close(ch)
	}()
	return ch
}

func TestFormatVolumes(t *testing.T) {
	if err := newCLI("tmpl").fmtOutput(newVolumeChannel()); err != nil {
		t.Fatal(err)
	}
}

func TestFormatVolumesWithError(t *testing.T) {
	if err := newCLI("tmpl").fmtOutput(
		newVolumeChannelWithError(true)); err != nil {
		t.Fatal(err)
	}
}

func TestFormatVolumesCustom(t *testing.T) {
	if err := newCLI("{{.ID}}").fmtOutput(newVolumeChannel()); err != nil {
		t.Fatal(err)
	}
}

func TestFormatVolumesCustomRaw(t *testing.T) {
	if err := newCLI("{{.Name}}").fmtOutput(
		newVolumeChannel()); err != nil {
		t.Fatal(err)
	}
}

func TestFormatVolumesAsJSON(t *testing.T) {
	if err := newCLI("json").fmtOutput(
		newVolumeChannel()); err != nil {
		t.Fatal(err)
	}
}

func TestFormatVolumesAsJSONP(t *testing.T) {
	if err := newCLI("jsonp").fmtOutput(
		newVolumeChannel()); err != nil {
		t.Fatal(err)
	}
}

func TestTemplateDefs(t *testing.T) {
	tpl := template.Must(template.New("temp").Parse(`{{define "T1"}}ONE{{end}}
{{define "T2"}}TWO{{end}}
{{define "T3"}}{{template "T1"}} {{template "T2"}}{{end}}
{{define "T4"}}{{template "T1"}} {{template "T2"}}{{end}}`))
	tpl.ExecuteTemplate(os.Stdout, "T3", nil)
	tpl.ExecuteTemplate(os.Stdout, "T4", nil)
}
