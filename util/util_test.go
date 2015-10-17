package util

import (
	"testing"
)

func TestStringInSlice(t *testing.T) {

	var r bool

	r = StringInSlice("hi", []string{"hello", "world"})
	if r {
		t.Fatal("hi there!")
	}

	r = StringInSlice("hi", []string{"hi", "world"})
	if !r {
		t.Fatal("hi where?")
	}
}

func TestTrimSingleWord(t *testing.T) {

	s := Trim(`

						hi


     		    

    `)

	if s != "hi" {
		t.Fatalf("trim failed '%v'", s)
	}
}

func TestTrimMultipleWords(t *testing.T) {

	s := Trim(`

						hi

		there

		     you
    `)

	if s != `hi

		there

		     you` {
		t.Fatalf("trim failed '%v'", s)
	}
}

func TestFileExists(t *testing.T) {
	if !FileExists("/bin/sh") {
		t.Fail()
	}
}

func TestFileExistsInPath(t *testing.T) {
	if !FileExistsInPath("sh") {
		t.Fail()
	}
}
