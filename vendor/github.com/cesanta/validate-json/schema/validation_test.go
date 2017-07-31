package schema

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	json "github.com/cesanta/ucl"
	"github.com/fatih/color"
)

func serveRemotes(t *testing.T, done chan struct{}) {
	defer close(done)
	err := http.ListenAndServe("localhost:1234", http.FileServer(http.Dir("schema-tests/remotes")))
	if err != nil {
		t.Log("ListenAndServe failed: %s", err)
	}
}

func testFiles(t *testing.T, files []string, loader *Loader) {
	var passing, total, schemaErrors int
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			t.Fatalf("Failed to open %q: %s", file, err)
		}
		v, err := json.Parse(f)
		f.Close()
		if err != nil {
			t.Fatalf("Failed to parse %q: %s", file, err)
		}

		tests, ok := v.(*json.Array)
		if !ok {
			t.Fatalf("Content of %q is not an array: %T", file, v)
		}
		for i, tt := range tests.Value {
			test, ok := tt.(*json.Object)
			if !ok {
				t.Fatalf("Test %d in %q is not an object", i, file)
			}
			t.Logf(color.BlueString("=====> Testing %s, case %d: %s", file, i, test.Find("description")))
			schema := test.Find("schema")
			err := ValidateDraft04Schema(schema)
			if err != nil {
				t.Errorf(color.RedString("schema does not pass validation: %s", err))
				schemaErrors++
			}
			v, _ := NewValidator(schema, loader) // not checking the error since schema is already validated
			cases := test.Find("tests").(*json.Array)
			for _, c := range cases.Value {
				total++
				case_ := c.(*json.Object)
				valid := case_.Find("valid").(*json.Bool).Value
				err := v.Validate(case_.Find("data"))
				switch {
				case err == nil && valid:
					passing++
					t.Logf("%s %s", color.GreenString("PASSED"), case_.Find("description"))
				case err != nil && !valid:
					passing++
					t.Logf("%s %s: %s", color.GreenString("PASSED"), case_.Find("description"), err)
				case err != nil && valid:
					t.Errorf("%s %s: %s", color.RedString("FAILED"), case_.Find("description"), err)
				case err == nil && !valid:
					t.Errorf("%s %s", color.RedString("FAILED"), case_.Find("description"))
				}
			}
		}
	}
	t.Logf("Passing %d out of %d tests (%g%%)", passing, total, float64(passing)/float64(total)*100)
	if schemaErrors > 0 {
		t.Logf("%d schemas failed validation", schemaErrors)
	}
}

func TestCompliance(t *testing.T) {
	done := make(chan struct{})
	go serveRemotes(t, done)
	f, err := os.Open("draft04schema.json")
	if err != nil {
		t.Fatalf("Failed to open draft04schema.json: %s", err)
	}
	s, err := json.Parse(f)
	f.Close()
	if err != nil {
		t.Fatalf("Failed to parse draft04schema.json: %s", err)
	}
	loader := NewLoader()
	loader.Add(s)
	loader.EnableNetworkAccess(true)

	select {
	case <-done:
		t.Fatalf("HTTP server died")
	default:
	}

	files, err := filepath.Glob("schema-tests/tests/draft4/*.json")
	if err != nil {
		t.Fatalf("Test files not found: %s", err)
	}
	testFiles(t, files, loader)
}

func TestFormat(t *testing.T) {
	testFiles(t, []string{"schema-tests/tests/draft4/optional/format.json"}, nil)
}

func TestZeroTerminatedFloats(t *testing.T) {
	testFiles(t, []string{"schema-tests/tests/draft4/optional/zeroTerminatedFloats.json"}, nil)
}
