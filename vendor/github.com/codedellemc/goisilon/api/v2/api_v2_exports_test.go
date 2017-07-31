package v2

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codedellemc/goisilon/api/json"
)

func TestExportEncodeJSON(t *testing.T) {
	clients := []string{}
	ex := &Export{ID: 3, Clients: &clients}
	buf, err := json.Marshal(ex)
	if err != nil {
		t.Fatal(err)
	}
	s := string(buf)
	if !assert.Equal(t, `{"clients":[]}`, s) {
		t.FailNow()
	}
	t.Log(s)
}
func TestExportDecodeJSON(t *testing.T) {
	j := `{"id":3,"clients":[]}`
	var ex Export
	if err := json.Unmarshal([]byte(j), &ex); err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "%+v\n", ex)
}

func TestPersonaIDTypeMarshal(t *testing.T) {
	pidt := PersonaIDTypeUser
	assert.Equal(t, "user", pidt.String())

	buf, err := json.Marshal(pidt)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `"user"`, string(buf))

	assert.Equal(t, PersonaIDTypeUser, ParsePersonaIDType("user"))
	assert.Equal(t, PersonaIDTypeUser, ParsePersonaIDType("USER"))

	assert.Equal(t, PersonaIDTypeGroup, ParsePersonaIDType("group"))
	assert.Equal(t, PersonaIDTypeGroup, ParsePersonaIDType("GROUP"))

	assert.Equal(t, PersonaIDTypeUID, ParsePersonaIDType("uid"))
	assert.Equal(t, PersonaIDTypeUID, ParsePersonaIDType("UID"))

	assert.Equal(t, PersonaIDTypeGID, ParsePersonaIDType("gid"))
	assert.Equal(t, PersonaIDTypeGID, ParsePersonaIDType("GID"))

	if err := json.Unmarshal([]byte(`"user"`), &pidt); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, PersonaIDTypeUser, pidt)

}

func TestPersonaIDMarshal(t *testing.T) {

	pid := &PersonaID{
		ID:   "akutz",
		Type: PersonaIDTypeUser,
	}

	buf, err := json.Marshal(pid)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `"user:akutz"`, string(buf))
}

func TestOneExportListMarshal(t *testing.T) {
	testAllExportListMarshal(t, getOneExportJSON)
}

func TestAllExportListMarshal(t *testing.T) {
	testAllExportListMarshal(t, getAllExportsJSON)
}

func TestAllExportListMarshal2(t *testing.T) {
	testAllExportListMarshal(t, getAllExports2JSON)
}

func TestAllExportListMarshal3(t *testing.T) {
	testAllExportListMarshal(t, getAllExports3JSON)
}

func testAllExportListMarshal(t *testing.T, list []byte) {
	var exList ExportList
	if err := json.Unmarshal(list, &exList); err != nil {
		t.Fatal(err)
	}

	buf, err := json.Marshal(exList)
	if err != nil {
		t.Fatal(err)
	}

	map1 := map[string]interface{}{}
	if err := json.Unmarshal(buf, &map1); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(buf, &exList); err != nil {
		t.Fatal(err)
	}

	buf, err = json.Marshal(exList)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(buf))

	map2 := map[string]interface{}{}
	if err := json.Unmarshal(buf, &map2); err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, map1, map2)
}
