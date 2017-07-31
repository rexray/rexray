package goof

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewe(t *testing.T) {
	g1 := WithError("g1", New("g2"))
	buf, _ := json.Marshal(g1)
	if !assert.Equal(t, `{"inner":"g2","msg":"g1"}`, string(buf)) {
		t.FailNow()
	}

	g1 = WithError("g1", errors.New("g2"))
	buf, _ = json.Marshal(g1)
	if !assert.Equal(t, `{"inner":{},"msg":"g1"}`, string(buf)) {
		t.FailNow()
	}

	g1 = Newe(g1)
	buf, _ = json.Marshal(g1)
	if !assert.Equal(t, `{"inner":"g2","msg":"g1"}`, string(buf)) {
		t.FailNow()
	}

	g1 = WithError("g1", WithError("g2", errors.New("g3")))
	buf, _ = json.Marshal(g1)
	if !assert.Equal(
		t, `{"inner":{"inner":{},"msg":"g2"},"msg":"g1"}`, string(buf)) {
		t.FailNow()
	}

	g1 = Newe(g1)
	buf, _ = json.Marshal(g1)
	if !assert.Equal(
		t, `{"inner":{"inner":"g3","msg":"g2"},"msg":"g1"}`, string(buf)) {
		t.FailNow()
	}
}

func TestFormat(t *testing.T) {
	e := WithField("hello", "world", "introduction error")
	assert.EqualValues(t, "introduction error", fmt.Sprint(e))
	assert.EqualValues(t, "introduction error", fmt.Sprintf("%s", e))
	assert.EqualValues(t, `"introduction error"`, fmt.Sprintf("%q", e))
	assert.EqualValues(t, "`introduction error`", fmt.Sprintf("%#q", e))
	assert.EqualValues(t, "       introduction error", fmt.Sprintf("%25s", e))
	assert.EqualValues(t, "introduction error       ", fmt.Sprintf("%-25s", e))
	assert.EqualValues(t, `     "introduction error"`, fmt.Sprintf("%25q", e))
	assert.EqualValues(t, `"introduction error"     `, fmt.Sprintf("%-25q", e))
	assert.EqualValues(t, "     `introduction error`", fmt.Sprintf("%#25q", e))
	assert.EqualValues(t, "`introduction error`     ", fmt.Sprintf("%-#25q", e))
	assert.EqualValues(t, "`introduction error`     ", fmt.Sprintf("%#-25q", e))

	assertMsgAndString(t, e, false, false, false)
	assertMsgAndString(t, e, true, false, false)
}

func TestError(t *testing.T) {
	e := WithField("hello", "world", "introduction error")
	assertMsgAndString(t, e, false, false, false)
	assertMsgAndString(t, e, false, true, false)
}

func TestString(t *testing.T) {
	e := WithField("hello", "world", "introduction error")
	assertMsgAndString(t, e, false, false, false)
	assertMsgAndString(t, e, false, false, true)
}

func assertMsgAndString(t *testing.T, e Goof, incErr, incFmt, incStr bool) {
	e.IncludeFieldsInError(incErr)
	e.IncludeFieldsInFormat(incFmt)
	e.IncludeFieldsInString(incStr)
	assertMsgAndStringActual(t, e.Error(), incErr)
	assertMsgAndStringActual(t, e.String(), incStr)
	assertMsgAndStringActual(t, fmt.Sprintf("%s", e), incFmt)
}

func assertMsgAndStringActual(t *testing.T, actual string, inc bool) {
	if inc {
		assert.EqualValues(t, `msg="introduction error" hello=world`, actual)
	} else {
		assert.EqualValues(t, "introduction error", actual)
	}
}

func TestMarshalToJSONSansMessage(t *testing.T) {
	e := WithFields(map[string]interface{}{
		"resourceID": 123,
	}, "invalid resource ID")
	buf, err := json.Marshal(e)
	assert.NoError(t, err)
	t.Log(string(buf))
}

func TestMarshalIndentToJSONSansMessage(t *testing.T) {
	e := WithFields(map[string]interface{}{
		"resourceID": 123,
	}, "invalid resource ID")
	buf, err := json.MarshalIndent(e, "", "  ")
	assert.NoError(t, err)
	t.Log(string(buf))
}

func TestMarshalToJSONWithMessage(t *testing.T) {
	e := WithFields(map[string]interface{}{
		"resourceID": 123,
	}, "invalid resource ID")
	e.IncludeMessageInJSON(true)
	buf, err := json.Marshal(e)
	assert.NoError(t, err)
	t.Log(string(buf))
}

func TestMarshalIndentToJSONWithMessage(t *testing.T) {
	e := WithFields(map[string]interface{}{
		"resourceID": 123,
	}, "invalid resource ID")
	e.IncludeMessageInJSON(true)
	buf, err := json.MarshalIndent(e, "", "  ")
	assert.NoError(t, err)
	t.Log(string(buf))
}

func newHTTPError() HTTPError {
	goofErr := WithFieldE(
		"fu", 3, "fubar",
		WithError(
			"dagnabbit", fmt.Errorf("broken"),
		),
	)
	return NewHTTPError(goofErr, 404)
}

func TestNewHTTPError(t *testing.T) {
	ValidateInnerErrorJSON = true
	httpErr := newHTTPError()
	buf, err := json.MarshalIndent(httpErr, "", "  ")
	assert.NoError(t, err)
	t.Log(string(buf))
}

func TestDecode(t *testing.T) {
	ValidateInnerErrorJSON = true
	expectedHTTPErr := newHTTPError()
	expectedBuf, err := json.MarshalIndent(expectedHTTPErr, "", "  ")
	assert.NoError(t, err)

	actualHTTPErr := &httpError{}
	err = actualHTTPErr.UnmarshalJSON(expectedBuf)
	assert.NoError(t, err)

	actualBuf, err := json.MarshalIndent(actualHTTPErr, "", "  ")
	assert.NoError(t, err)

	actualStr := string(actualBuf)
	expectedStr := string(expectedBuf)

	assert.EqualValues(t, expectedStr, actualStr)
	t.Log(expectedStr)
	t.Log(actualStr)
}

var httpErrorValue = []byte(`{
    "message": "fubar",
    "status": 404,
    "error": {
        "fu": 3,
        "inner": {
            "inner": "broken",
            "msg": "dagnabbit"
        }
    }
}`)
