package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"

	log "github.com/Sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	apitypes "github.com/emccode/libstorage/api/types"
)

const (
	volFormat  = "{{.ID}}\t{{.Name}}\t{{.Status}}\t{{.Size}}\n"
	volsFormat = "ID\tName\tStatus\tSize\n" +
		`{{range sort . "Name"}}` + volFormat + `{{end}}`

	snpFormat  = "{{.ID}}\t{{.Name}}\t{{.Status}}\t{{.VolumeID}}\t{{.VolumeSize}}\n"
	snpsFormat = "ID\tName\tStatus\tVolumeID\tVolumeSize\n" +
		`{{range sort . "Name"}}` + snpFormat + `{{end}}`

	instFormat  = "{{.InstanceID.ID}}\t{{.Name}}\t{{.ProviderName}}\t{{.Region}}\n"
	instsFormat = "ID\tName\tProvider\tRegion\n" +
		`{{range sort . "Name"}}` + instFormat + `{{end}}`

	siFormat = "Name\tDriver\n" +
		"{{.Name}}\t{{.Driver.Name}}\n"
	sisFormat = `{{range sort . "Name"}}` + siFormat + `{{end}}`

	mntFormat = "ID\tRoot\tSource\tMountPont\n" +
		"{{.ID}}\t{{.Root}}\t{{.Source}}\t{{.MountPoint}}\n"
	mntsFormat = `{{range sort . "Source"}}` + mntFormat + `{{end}}`
)

var (
	errUnsupportedType = errors.New("unsupported type")

	volTempl  = newTemplate("vol", volFormat)
	volsTempl = newTemplate("vol", volsFormat)

	snpTempl  = newTemplate("snp", snpFormat)
	snpsTempl = newTemplate("snps", snpsFormat)

	instTempl  = newTemplate("inst", instFormat)
	instsTempl = newTemplate("insts", instsFormat)

	siTempl  = newTemplate("si", siFormat)
	sisTempl = newTemplate("sis", sisFormat)

	mntTempl  = newTemplate("mnt", mntFormat)
	mntsTempl = newTemplate("mnts", mntsFormat)
)

func newTemplate(name, format string) *template.Template {
	return template.Must(template.New(name).Funcs(funcMap).Parse(format))
}

func fmtOutput(w io.Writer, f string, o interface{}) error {

	var (
		tw  *tabwriter.Writer
		twf func() error
	)

	defer func() {
		if twf != nil {
			twf()
		}
	}()

	newTabWriter := func() {
		tw = tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		twf = tw.Flush
		w = tw
	}

	switch {
	case strings.EqualFold(f, "json"):
		enc := json.NewEncoder(w)
		return enc.Encode(o)
	case strings.EqualFold(f, "yaml"):
		buf, err := yaml.Marshal(o)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w, bytes.NewReader(buf)); err != nil {
			return err
		}
	case strings.EqualFold(f, "tmpl"):
		newTabWriter()
		switch to := o.(type) {
		case string:
			fmt.Fprint(w, o)
		case *apitypes.Volume:
			return fmtOutput(w, f, []*apitypes.Volume{to})
		case []*apitypes.Volume:
			if len(to) == 0 {
				return nil
			}
			return volsTempl.Execute(w, o)
		case *apitypes.Snapshot:
			return fmtOutput(w, f, []*apitypes.Snapshot{to})
		case []*apitypes.Snapshot:
			if len(to) == 0 {
				return nil
			}
			return snpsTempl.Execute(w, o)
		case map[string]*apitypes.ServiceInfo:
			if len(to) == 0 {
				return nil
			}
			return sisTempl.Execute(w, o)
		case map[string]*apitypes.Instance:
			if len(to) == 0 {
				return nil
			}
			return instsTempl.Execute(w, o)
		case []*apitypes.MountInfo:
			if len(to) == 0 {
				return nil
			}
			return mntsTempl.Execute(w, o)
		default:
			panic(errUnsupportedType)
		}
	default:
		newTabWriter()
		f = fmt.Sprintf(`"%s"`, strings.Replace(f, `"`, `\"`, -1))
		uf, err := strconv.Unquote(f)
		if err != nil {
			return err
		}
		t, err := template.New("temp").Funcs(funcMap).Parse(uf)
		if err != nil {
			return err
		}
		return t.Execute(w, o)
	}

	return nil
}

func (c *CLI) marshalOutput(v interface{}) {
	if err := fmtOutput(os.Stdout, c.outputFormat, v); err != nil {
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
