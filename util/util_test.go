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
