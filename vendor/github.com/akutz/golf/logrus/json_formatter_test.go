package logrus

import (
	"bytes"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func newJsonFormatter() *JSONFormatter {
	return &JSONFormatter{log.JSONFormatter{}}
}

func TestLogrusWithJsonFormatter(t *testing.T) {
	p := &Person{
		Name:  "Bruce",
		Alias: "Batman",
		Hideout: &Hideout{
			Name:        "JLU Tower",
			DimensionId: 52,
		},
	}

	log.SetFormatter(newJsonFormatter())

	entry := log.NewEntry(log.StandardLogger())
	entry.Message = "the dark knight"
	entry.Data = log.Fields{"hero": p}
	s, _ := entry.String()

	if !strings.Contains(s, `"hero.name":"Bruce"`) {
		t.Fatalf(`missing "hero.name":"Bruce"`)
	}

	if !strings.Contains(s, `"hero.alias":"Batman"`) {
		t.Fatalf(`missing "hero.alias":"Batman"`)
	}

	if !strings.Contains(s, `"hero.hideout.name":"JLU Tower"`) {
		t.Fatalf(`missing "hero.hideout.name":"JLU Tower"`)
	}

	if !strings.Contains(s, `"hero.hideout.dimensionId":52`) {
		t.Fatalf(`missing "hero.hideout.dimensionId":52`)
	}
}

func TestGolfsWithJsonFormatter(t *testing.T) {
	p := &Person{
		Name:  "Bruce",
		Alias: "Batman",
		Hideout: &Hideout{
			Name:        "JLU Tower",
			DimensionId: 52,
		},
	}

	jf := newJsonFormatter()
	b, err := jf.Format(&log.Entry{
		Message: "the dark knight", Data: log.Fields{"hero": p}})
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}

	if bytes.Index(b, ([]byte)(`"hero.name":"Bruce"`)) < 0 {
		t.Fatalf(`missing "hero.name":"Bruce"`)
	}

	if bytes.Index(b, ([]byte)(`"hero.alias":"Batman"`)) < 0 {
		t.Fatalf(`missing "hero.alias":"Batman"`)
	}

	if bytes.Index(b, ([]byte)(`"hero.hideout.name":"JLU Tower"`)) < 0 {
		t.Fatalf(`missing "hero.hideout.name":"JLU Tower"`)
	}

	if bytes.Index(b, ([]byte)(`"hero.hideout.dimensionId":52`)) < 0 {
		t.Fatalf(`missing "hero.hideout.dimensionId":52`)
	}
}

func TestGolfsWithJsonFormatterAndNonGolfer(t *testing.T) {
	h := &Hideout{
		Name:        "JLU Tower",
		DimensionId: 52,
	}

	jf := newJsonFormatter()
	b, err := jf.Format(&log.Entry{
		Message: "secret base", Data: log.Fields{"hideout": h}})
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}
	t.Log(string(b))

	if bytes.Index(b, ([]byte)(`"hideout":{"Name":"JLU Tower","DimensionId":52}`)) < 0 {
		t.Fatalf(`missing "hideout":{"Name":"JLU Tower","DimensionId":52}`)
	}
}
