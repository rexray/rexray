package golf

import (
	"fmt"
	"testing"
)

func TestInt(t *testing.T) {
	f := Fore("hero", 3)
	assertMapLen(t, f, 1)
	assertKeyEquals(t, f, "hero", 3)
}

func TestString(t *testing.T) {
	f := Fore("hero", "three")
	assertMapLen(t, f, 1)
	assertKeyEquals(t, f, "hero", "three")
}

func TestStringThatGolfs(t *testing.T) {
	s := StringThatGolfs("three")
	f := Fore("hero", &s)
	t.Log(f)
	assertMapLen(t, f, 1)
	assertKeyEquals(t, f, "hero.golfer", "three three")
}

func TestNil(t *testing.T) {
	f := Fore("hero", nil)
	if f != nil {
		t.Fatal("not nil")
	}
}

func TestStructPointer(t *testing.T) {
	f := Fore("test", &StructWithPointer{Foo: &FooStruct{Bar: "value"}})
	assertMapLen(t, f, 1)
	assertKeyEquals(t, f, "test.foo.Bar", "value")
}

func TestStructNonPointer(t *testing.T) {
	f := Fore("test", &StructWithoutPointer{Foo: FooStruct{Bar: "value"}})
	assertMapLen(t, f, 1)
	assertKeyEquals(t, f, "test.foo.Bar", "value")
}

type StringThatGolfs string

func (s *StringThatGolfs) GolfExportedFields() map[string]interface{} {
	return map[string]interface{}{"golfer": fmt.Sprintf("%s %s", *s, *s)}
}

type FooStruct struct {
	Bar string
}

type StructWithPointer struct {
	Foo *FooStruct `golf:"foo"`
}

type StructWithoutPointer struct {
	Foo FooStruct `golf:"foo"`
}
