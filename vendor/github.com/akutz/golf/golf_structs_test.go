package golf

import (
	"testing"
)

func TestPersonStructDefaultOptions(t *testing.T) {
	f := Fore("hero", getBatman())
	assertMapLen(t, f, 4)
	assertKeyEquals(t, f, "hero.name", "Bruce")
	assertKeyEquals(t, f, "hero.alias", "Batman")
	assertKeyEquals(t, f, "hero.hideout.name", "JLU Tower")
	assertKeyEquals(t, f, "hero.hideout.dimensionId", 52)
}

func TestPersonStructMissingKey(t *testing.T) {
	p := getBatman()
	p.Hideout = nil
	f := Fore("hero", p)
	assertMapLen(t, f, 3)
	assertKeyEquals(t, f, "hero.name", "Bruce")
	assertKeyEquals(t, f, "hero.alias", "Batman")
	assertKeyMissing(t, f, "hero.hideout.name")
	assertKeyMissing(t, f, "hero.hideout.dimensionId")
}

func TestPersonStructPreferJsonTags(t *testing.T) {
	p := getBatman()
	p.Hideout = nil
	p.golfUseTypeExportedFields = true
	p.golfJsonTagBehavior = IgnoreGolfTags
	f := Fore("hero", p)
	assertMapLen(t, f, 3)
	assertKeyEquals(t, f, "hero.secretIdentity", "Bruce")
	assertKeyEquals(t, f, "hero.name", "Batman")
	assertKeyEquals(t, f, "hero.Hideout", nil)
	assertKeyMissing(t, f, "hero.hideout.dimensionId")
}

func TestPersonStructMissingKeyWithHideoutTypeExported(t *testing.T) {
	p := getBatman()
	p.Hideout = nil
	p.golfUseTypeExportedFields = true
	f := Fore("hero", p)
	assertMapLen(t, f, 2)
	assertKeyEquals(t, f, "hero.name", "Bruce")
	assertKeyEquals(t, f, "hero.alias", "Batman")
	assertKeyMissing(t, f, "hero.hideout.name")
	assertKeyMissing(t, f, "hero.hideout.dimensionId")
}

type Person struct {
	Name                      string   `golf:"name" json:"secretIdentity"`
	Alias                     string   `golf:"alias" json:"name"`
	Hideout                   *Hideout `golf:"hideout,omitempty"`
	golfUseTypeExportedFields bool
	golfJsonTagBehavior       int
}

func (p *Person) GolfExportedFields() map[string]interface{} {
	if p.golfUseTypeExportedFields {
		return nil
	}
	return map[string]interface{}{
		"name": p.Name, "alias": p.Alias, "hideout": p.Hideout}
}

func (p *Person) GolfJsonTagBehavior() int {
	return p.golfJsonTagBehavior
}

type Hideout struct {
	Name                      string `golf:"name"`
	DimensionId               int    `golf:"dimensionId" json:"id"`
	golfUseTypeExportedFields bool
	golfJsonTagBehavior       int
}

func (h *Hideout) GolfExportedFields() map[string]interface{} {
	if h.golfUseTypeExportedFields {
		return nil
	}
	return map[string]interface{}{
		"name": h.Name, "dimensionId": h.DimensionId}
}

func (h *Hideout) GolfJsonTagBehavior() int {
	return h.golfJsonTagBehavior
}

func getBatman() *Person {
	return &Person{
		Name:  "Bruce",
		Alias: "Batman",
		Hideout: &Hideout{
			Name:        "JLU Tower",
			DimensionId: 52,
		},
	}
}
