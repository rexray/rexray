package logrus

import (
	"bytes"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func newTextFormatter() *TextFormatter {
	return &TextFormatter{log.TextFormatter{DisableColors: true}}
}

func TestLogrusWithTextFormatter(t *testing.T) {
	p := &Person{
		Name:  "Bruce",
		Alias: "Batman",
		Hideout: &Hideout{
			Name:        "JLU Tower",
			DimensionId: 52,
		},
	}

	log.SetFormatter(newTextFormatter())

	entry := log.NewEntry(log.StandardLogger())
	entry.Message = "the dark knight"
	entry.Data = log.Fields{"hero": p}
	s, _ := entry.String()

	if !strings.Contains(s, "hero.name=Bruce") {
		t.Fatalf("missing hero.name=Bruce")
	}

	if !strings.Contains(s, "hero.alias=Batman") {
		t.Fatalf("missing hero.alias=Batman")
	}

	if !strings.Contains(s, `hero.hideout.name="JLU Tower"`) {
		t.Fatalf(`missing hero.hideout.name="JLU Tower"`)
	}

	if !strings.Contains(s, "hero.hideout.dimensionId=52") {
		t.Fatalf("missing hero.hideout.dimensionId=52")
	}
}

func TestGolfsWithTextFormatter(t *testing.T) {
	p := &Person{
		Name:  "Bruce",
		Alias: "Batman",
		Hideout: &Hideout{
			Name:        "JLU Tower",
			DimensionId: 52,
		},
	}

	tf := newTextFormatter()
	b, _ := tf.Format(&log.Entry{
		Message: "the dark knight", Data: log.Fields{"hero": p}})

	if bytes.Index(b, ([]byte)("hero.name=Bruce")) < 0 {
		t.Fatalf("missing hero.name=Bruce")
	}

	if bytes.Index(b, ([]byte)("hero.alias=Batman")) < 0 {
		t.Fatalf("missing hero.alias=Batman")
	}

	if bytes.Index(b, ([]byte)(`hero.hideout.name="JLU Tower"`)) < 0 {
		t.Fatalf(`missing hero.hideout.name="JLU Tower"`)
	}

	if bytes.Index(b, ([]byte)("hero.hideout.dimensionId=52")) < 0 {
		t.Fatalf("missing hero.hideout.dimensionId=52")
	}
}

func TestGolfsWithTextFormatterAndNonGolfer(t *testing.T) {
	h := &Hideout{
		Name:        "JLU Tower",
		DimensionId: 52,
	}

	tf := newTextFormatter()
	b, _ := tf.Format(&log.Entry{
		Message: "secret base", Data: log.Fields{"hideout": h}})
	t.Log(string(b))

	if bytes.Index(b, ([]byte)("hideout=&{JLU Tower 52}")) < 0 {
		t.Fatalf("missing hideout={JLU Tower 52}")
	}
}
