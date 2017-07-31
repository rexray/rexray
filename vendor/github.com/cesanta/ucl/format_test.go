package ucl

import (
	"bytes"
	"testing"
)

// These tests are here to ensure stable formatting. If your changes make some of them fail,
// please add another option to FormatConfig (and a test) and gate your changes on that option.

func run(t *testing.T, config *FormatConfig, cases map[string]string) {
	for in, expected := range cases {
		v, err := parse([]rune(in))
		if err != nil {
			t.Errorf("Failed to parse %q: %s", in, err)
			continue
		}
		w := bytes.NewBuffer(nil)
		if err = Format(v, config, w); err != nil {
			t.Errorf("Format(%q) failed: %s", in, err)
			continue
		}
		if w.String() != expected {
			t.Errorf("Format for %q changed. Got:\n%s\nWant:\n%s", in, w.String(), expected)
		}
	}
}

func TestCanonicalFormat(t *testing.T) {
	cases := map[string]string{
		`[]`: `[]`,
		`{}`: `{}`,
		`{"a":"b"}`: `{
  "a": "b"
}`,
		`{"a":[123,456]," ":{"b":1,"a":2}}`: `{
  " ": {
    "a": 2,
    "b": 1
  },
  "a": [
    123,
    456
  ]
}`,
		`["a/b"]`: `[
  "a/b"
]`,
	}
	run(t, nil, cases)
}

func TestMultilineObjectThreshold(t *testing.T) {
	config := &FormatConfig{MultilineObjectThreshold: 20}
	cases := map[string]string{
		`{"a":1}`:       `{"a": 1}`,
		`{"a":1,"b":2}`: `{"a": 1, "b": 2}`,
		`{"a":{"a":1}}`: `{"a": {"a": 1}}`,
		`{"a":[123]}`: `{
  "a": [
    123
  ]
}`,
		`{"a":"this string is longer than 20 characters"}`: `{
  "a": "this string is longer than 20 characters"
}`,
	}
	run(t, config, cases)
}

func TestMultilineArrayThreshold(t *testing.T) {
	config := &FormatConfig{MultilineArrayThreshold: 20}
	cases := map[string]string{
		`["a"]`:         `["a"]`,
		`["a","b"]`:     `["a", "b"]`,
		`["a",[["b"]]]`: `["a", [["b"]]]`,
		`[{"a":[1,2,3]}]`: `[
  {
    "a": [1, 2, 3]
  }
]`,
		`["a","this string is longer than 20 characters"]`: `[
  "a",
  "this string is longer than 20 characters"
]`,
	}
	run(t, config, cases)
}

func TestMultilineArrayObjectThreshold(t *testing.T) {
	config := &FormatConfig{
		MultilineArrayThreshold:  20,
		MultilineObjectThreshold: 20,
	}
	cases := map[string]string{
		`{"a":[123]}`: `{"a": [123]}`,
	}
	run(t, config, cases)
}

func TestPreserveObjectKeysOrder(t *testing.T) {
	config := &FormatConfig{PreserveObjectKeysOrder: true}
	cases := map[string]string{
		`{"a":[123,456]," ":{"b":1,"a":2}}`: `{
  "a": [
    123,
    456
  ],
  " ": {
    "b": 1,
    "a": 2
  }
}`,
	}
	run(t, config, cases)
}
