package gournal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testExe         = os.Args[0]
	exitError       = &exec.ExitError{}
	doTestFatal, _  = strconv.ParseBool(os.Getenv("GOURNAL_TEST_FATAL"))
	testFatalLvl    = ParseLevel(os.Getenv("GOURNAL_TEST_FATAL_LEVEL"))
	testFatalFields = os.Getenv("GOURNAL_TEST_FATAL_FIELDS")
)

func TestMain(m *testing.M) {
	DefaultLevel = DebugLevel
	os.Exit(m.Run())
}

func TestContextKeys(t *testing.T) {
	ctx := context.Background()
	a := NewAppender()
	ctx1 := context.WithValue(ctx, appenderKey, a)
	ctx2 := context.WithValue(ctx, appenderKeyC, a)
	ctx3 := context.WithValue(ctx, AppenderKey(), a)
	assert.Equal(t, a, ctx1.Value(appenderKey))
	assert.Equal(t, a, ctx1.Value(appenderKeyC))
	assert.Equal(t, a, ctx1.Value(AppenderKey()))
	assert.Equal(t, a, ctx2.Value(appenderKey))
	assert.Equal(t, a, ctx2.Value(appenderKeyC))
	assert.Equal(t, a, ctx2.Value(AppenderKey()))
	assert.Equal(t, a, ctx3.Value(appenderKey))
	assert.Equal(t, a, ctx3.Value(appenderKeyC))
	assert.Equal(t, a, ctx3.Value(AppenderKey()))
}

func TestParseLevelTransformations(t *testing.T) {
	for lvl := UnknownLevel; lvl < levelCount; lvl++ {
		s := lvl.String()
		assert.Equal(t, lvl, ParseLevel(strings.ToUpper(s)))
		assert.Equal(t, lvl, ParseLevel(strings.ToLower(s)))
		assert.Equal(t, lvl, ParseLevel(strings.ToTitle(s)))
		t.Logf("validated level=%s", s)
	}
}

func TestChildContextLevelLog(t *testing.T) {
	buf, ctx := newTestContext()
	ctx = context.WithValue(ctx, LevelKey(), ErrorLevel)
	Info(ctx, "Hello %s", "Bob")
	assert.Zero(t, buf.Len())

	ctx = context.WithValue(ctx, LevelKey(), DebugLevel)
	Debug(ctx, "Hello %s", "Alice")
	assert.Equal(t, "[DEBUG] Hello Alice\n", buf.String())
}

func TestContextFieldsMap(t *testing.T) {
	buf, ctx := newTestContext()

	ctx = context.WithValue(ctx, FieldsKey(), map[string]interface{}{
		"point": struct {
			x int
			y int
		}{1, -1},
	})

	Info(ctx, "Discovered planet")
	assert.Equal(
		t, "[INFO] Discovered planet map[point:{1 -1}]\n", buf.String())
}

func TestContextFieldsFunc(t *testing.T) {
	buf, ctx := newTestContext()

	ctxFieldsFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"point": struct {
				x int
				y int
			}{1, -1},
		}
	}
	ctx = context.WithValue(ctx, FieldsKey(), ctxFieldsFunc)

	Info(ctx, "Discovered planet")
	assert.Equal(
		t, "[INFO] Discovered planet map[point:{1 -1}]\n", buf.String())
}

func TestContextFieldsFuncEx(t *testing.T) {
	buf, ctx := newTestContext()

	ctxFieldsFunc := func(
		ctx context.Context,
		lvl Level,
		fields map[string]interface{},
		msg string) map[string]interface{} {

		if v, ok := fields["size"].(int); ok {
			delete(fields, "size")
			return map[string]interface{}{
				"point": struct {
					x int
					y int
					z int
				}{1, -1, v},
			}
		}

		return map[string]interface{}{
			"point": struct {
				x int
				y int
			}{1, -1},
		}
	}
	ctx = context.WithValue(ctx, FieldsKey(), ctxFieldsFunc)
	ctxLogger := New(ctx)

	Info(ctx, "Discovered planet")
	assert.Equal(
		t, "[INFO] Discovered planet map[point:{1 -1}]\n", buf.String())

	buf.Reset()
	assert.Equal(t, buf.Len(), 0)

	ctxLogger.Info("Discovered planet")
	assert.Equal(
		t, "[INFO] Discovered planet map[point:{1 -1}]\n", buf.String())

	buf.Reset()
	assert.Equal(t, buf.Len(), 0)

	WithField("size", 3).Info(ctx, "Discovered planet")
	assert.Equal(
		t, "[INFO] Discovered planet map[point:{1 -1 3}]\n", buf.String())
}

func TestAppendWithNilContext(t *testing.T) {
	runLoggerTests(
		t,
		"TestAppendWithNilContext",
		nil,
		nil,
		"TestAppendWithNilContext")
}

func TestAppendWithContextNoAppenderNoLevel(t *testing.T) {
	runLoggerTests(
		t,
		"TestAppendWithContextNoAppenderNoLevel",
		context.Background(),
		nil,
		"TestAppendWithContextNoAppenderNoLevel")
}

func TestAppendWithContextPanicLevel(t *testing.T) {
	testAppendWithContextXYZLevel(t, PanicLevel)
}

func TestAppendWithContextFatalLevel(t *testing.T) {
	testAppendWithContextXYZLevel(t, FatalLevel)
}

func TestAppendWithContextErrorLevel(t *testing.T) {
	testAppendWithContextXYZLevel(t, ErrorLevel)
}

func TestAppendWithContextWarnLevel(t *testing.T) {
	testAppendWithContextXYZLevel(t, WarnLevel)
}

func TestAppendWithContextInfoLevel(t *testing.T) {
	testAppendWithContextXYZLevel(t, InfoLevel)
}

func TestAppendWithContextDebugLevel(t *testing.T) {
	testAppendWithContextXYZLevel(t, DebugLevel)
}

func testAppendWithContextXYZLevel(t *testing.T, lvl Level) {
	ctx := context.WithValue(context.Background(), LevelKey(), lvl)
	lvlStr := lvl.String()
	lvlStr = fmt.Sprintf("%c%s", lvlStr[0], strings.ToLower(lvlStr[1:]))
	name := fmt.Sprintf("TestAppendWithContext%sLevel", lvlStr)
	runLoggerTests(t, name, ctx, nil, "Run Barry, run.")
}

func TestAppendWithNilContextPanicLevel(t *testing.T) {
	testAppendWithNilContextXYZLevel(t, PanicLevel)
}

func TestAppendWithNilContextFatalLevel(t *testing.T) {
	testAppendWithNilContextXYZLevel(t, FatalLevel)
}

func TestAppendWithNilContextErrorLevel(t *testing.T) {
	testAppendWithNilContextXYZLevel(t, ErrorLevel)
}

func TestAppendWithNilContextWarnLevel(t *testing.T) {
	testAppendWithNilContextXYZLevel(t, WarnLevel)
}

func TestAppendWithNilContextInfoLevel(t *testing.T) {
	testAppendWithNilContextXYZLevel(t, InfoLevel)
}

func TestAppendWithNilContextDebugLevel(t *testing.T) {
	testAppendWithNilContextXYZLevel(t, DebugLevel)
}

func testAppendWithNilContextXYZLevel(t *testing.T, lvl Level) {
	lvlStr := lvl.String()
	lvlStr = fmt.Sprintf("%c%s", lvlStr[0], strings.ToLower(lvlStr[1:]))
	name := fmt.Sprintf("TestAppendWithNilContext%sLevel", lvlStr)
	runLoggerTests(t, name, nil, nil, "Run Barry, run.")
}

func TestAppendWithFieldsPanicLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, PanicLevel)
}

func TestAppendWithFieldsFatalLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, FatalLevel)
}

func TestAppendWithFieldsErrorLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, ErrorLevel)
}

func TestAppendWithFieldsWarnLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, WarnLevel)
}

func TestAppendWithFieldsInfoLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, InfoLevel)
}

func TestAppendWithFieldsDebugLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, DebugLevel)
}

func testAppendWithFieldsXYZLevel(t *testing.T, lvl Level) {
	lvlStr := lvl.String()
	lvlStr = fmt.Sprintf("%c%s", lvlStr[0], strings.ToLower(lvlStr[1:]))
	name := fmt.Sprintf("TestAppendWithFields%sLevel", lvlStr)
	fields := map[string]interface{}{"size": "large"}
	runLoggerTests(t, name, nil, fields, "Run Barry, run.")
}

func TestAppendWithArgsPanicLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, PanicLevel)
}

func TestAppendWithArgsFatalLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, FatalLevel)
}

func TestAppendWithArgsErrorLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, ErrorLevel)
}

func TestAppendWithArgsWarnLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, WarnLevel)
}

func TestAppendWithArgsInfoLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, InfoLevel)
}

func TestAppendWithArgsDebugLevel(t *testing.T) {
	testAppendWithFieldsXYZLevel(t, DebugLevel)
}

func testAppendWithArgsXYZLevel(t *testing.T, lvl Level) {
	lvlStr := lvl.String()
	lvlStr = fmt.Sprintf("%c%s", lvlStr[0], strings.ToLower(lvlStr[1:]))
	name := fmt.Sprintf("TestAppendWithArgs%sLevel", lvlStr)
	fields := map[string]interface{}{"size": "large"}
	runLoggerTests(t, name, nil, fields, "Run %s, run.", "Barry")
}

func getLogFunction(lvl Level) func(context.Context, string, ...interface{}) {
	switch lvl {
	case PanicLevel:
		return Panic
	case FatalLevel:
		return Fatal
	case ErrorLevel:
		return Error
	case WarnLevel:
		return Warn
	case InfoLevel:
		return Info
	case DebugLevel:
		return Debug
	}
	return nil
}

func runLoggerTests(
	t *testing.T,
	testName string,
	ctx context.Context,
	fields map[string]interface{},
	msg string,
	args ...interface{}) {

	if doTestFatal {
		appndr := NewAppenderWithOptions(os.Stderr)
		if ctx == nil {
			DefaultAppender = appndr
			DefaultLevel = testFatalLvl
		} else {
			ctx = context.WithValue(ctx, AppenderKey(), appndr)
			ctx = context.WithValue(ctx, LevelKey(), testFatalLvl)
		}
		if len(testFatalFields) > 0 {
			fields = map[string]interface{}{}
			json.Unmarshal([]byte(testFatalFields), &fields)
			WithFields(fields).Fatal(ctx, msg, args...)
		} else {
			Fatal(ctx, msg, args...)
		}
		return
	}

	var (
		expStr string
		ctxLvl = getLevel(ctx)
		actBuf = &bytes.Buffer{}
		appndr = NewAppenderWithOptions(actBuf)
		tstArg = fmt.Sprintf("-test.run=%s$", testName)
	)

	if ctx == nil {
		DefaultAppender = appndr
	} else {
		ctx = context.WithValue(ctx, AppenderKey(), appndr)
	}

	for lvl := DebugLevel; lvl >= PanicLevel; lvl-- {

		if ctxLvl >= lvl {
			if len(args) > 0 {
				msg = fmt.Sprintf(msg, args...)
			}

			if len(fields) == 0 {
				expStr = fmt.Sprintf("[%s] %s\n", lvl, msg)
			} else {
				expStr = fmt.Sprintf("[%s] %s %v\n", lvl, msg, fields)
			}
		}

		var (
			panicHandled bool
			fatalHandled bool
		)

		switch lvl {
		case PanicLevel:
			func() {
				defer func() {
					if ctxLvl >= lvl {
						r := recover()
						assert.IsType(t, "", r)
						panicHandled = true
					}
				}()
				if len(fields) > 0 {
					WithFields(fields).Panic(ctx, msg, args...)
				} else {
					Panic(ctx, msg, args...)
				}
			}()
		case FatalLevel:
			cmd := exec.Command(testExe, tstArg)
			cmd.Stderr = actBuf
			envVars := []string{
				"GOURNAL_TEST_FATAL=true",
				fmt.Sprintf("GOURNAL_TEST_FATAL_LEVEL=%s", ctxLvl),
			}
			if len(fields) > 0 {
				fldBuf, _ := json.Marshal(fields)
				envVars = append(envVars, fmt.Sprintf(
					"GOURNAL_TEST_FATAL_FIELDS=%s", string(fldBuf)))
			}
			cmd.Env = append(os.Environ(), envVars...)
			err := cmd.Run()
			if ctxLvl >= lvl {
				assert.IsType(t, exitError, err)
				e := err.(*exec.ExitError)
				assert.False(t, e.Success())
				fatalHandled = true
			}
		case ErrorLevel:
			if len(fields) > 0 {
				WithFields(fields).Error(ctx, msg, args...)
			} else {
				Error(ctx, msg, args...)
			}
		case WarnLevel:
			if len(fields) > 0 {
				WithFields(fields).Warn(ctx, msg, args...)
			} else {
				Warn(ctx, msg, args...)
			}
		case InfoLevel:
			if len(fields) > 0 {
				WithFields(fields).Info(ctx, msg, args...)
			} else {
				Info(ctx, msg, args...)
			}
		case DebugLevel:
			if len(fields) > 0 {
				WithFields(fields).Debug(ctx, msg, args...)
			} else {
				Debug(ctx, msg, args...)
			}
		}

		if ctxLvl >= lvl {
			assert.Equal(t, expStr, actBuf.String())
			switch lvl {
			case PanicLevel:
				assert.True(t, panicHandled)
			case FatalLevel:
				assert.True(t, fatalHandled)
			}
			t.Logf("test success - %s - %s", testName, lvl)
		} else {
			assert.Equal(t, "", actBuf.String())
			switch lvl {
			case PanicLevel:
				assert.False(t, panicHandled)
			case FatalLevel:
				assert.False(t, fatalHandled)
			}
			t.Logf("test success - %s - %s - noop", testName, lvl)
		}
		actBuf.Reset()
	}
}

func newTestContext() (*bytes.Buffer, context.Context) {
	w := &bytes.Buffer{}
	a := NewAppenderWithOptions(w)
	return w, context.WithValue(context.Background(), AppenderKey(), a)
}
