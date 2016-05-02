package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompilePresent(t *testing.T) {
	f, err := CompileFilter(`(datacenter=*)`)
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, filterPresent, f.Op)
	assert.EqualValues(t, "datacenter", f.Left)
}

func TestCompileSubstring(t *testing.T) {
	f, err := CompileFilter(`(datacenter=*Texas*)`)
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, filterSubstrings, f.Op)
	assert.EqualValues(t, "datacenter", f.Left)
	assert.EqualValues(t, "Texas", f.Right)

}

func TestCompileSubstringPrefix(t *testing.T) {
	f, err := CompileFilter(`(datacenter=*Texas)`)
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, filterSubstringsPrefix, f.Op)
	assert.EqualValues(t, "datacenter", f.Left)
	assert.EqualValues(t, "Texas", f.Right)
}

func TestCompileSubstringPostfix(t *testing.T) {
	f, err := CompileFilter(`(datacenter=Texas*)`)
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, filterSubstringsPostfix, f.Op)
	assert.EqualValues(t, "datacenter", f.Left)
	assert.EqualValues(t, "Texas", f.Right)
}

func TestCompileAndEqualityFilter(t *testing.T) {
	f, err := CompileFilter(`(&(datacenter=irvine)(department=finance))`)
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, filterAnd, f.Op)

	assert.EqualValues(t, filterEqualityMatch, f.Children[0].Op)
	assert.EqualValues(t, "datacenter", f.Children[0].Left)
	assert.EqualValues(t, "irvine", f.Children[0].Right)

	assert.EqualValues(t, filterEqualityMatch, f.Children[1].Op)
	assert.EqualValues(t, "department", f.Children[1].Left)
	assert.EqualValues(t, "finance", f.Children[1].Right)
}

func TestCompileAndEqualityURLEscapedFilter(t *testing.T) {
	f, err := CompileFilter(`(%26(datacenter=irvine)(department=finance))`)
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, filterAnd, f.Op)

	assert.EqualValues(t, filterEqualityMatch, f.Children[0].Op)
	assert.EqualValues(t, "datacenter", f.Children[0].Left)
	assert.EqualValues(t, "irvine", f.Children[0].Right)

	assert.EqualValues(t, filterEqualityMatch, f.Children[1].Op)
	assert.EqualValues(t, "department", f.Children[1].Left)
	assert.EqualValues(t, "finance", f.Children[1].Right)
}

func TestCompileAndOrEqualityFilter(t *testing.T) {
	f, err := CompileFilter(
		`(&(|(datacenter=irvine)(datacenter=houston))(department=finance))`)
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, filterAnd, f.Op)

	assert.EqualValues(t, filterOr, f.Children[0].Op)

	assert.EqualValues(t, filterEqualityMatch,
		f.Children[0].Children[0].Op)
	assert.EqualValues(t, "datacenter",
		f.Children[0].Children[0].Left)
	assert.EqualValues(t, "irvine",
		f.Children[0].Children[0].Right)

	assert.EqualValues(t, filterEqualityMatch,
		f.Children[0].Children[1].Op)
	assert.EqualValues(t, "datacenter",
		f.Children[0].Children[1].Left)
	assert.EqualValues(t, "houston",
		f.Children[0].Children[1].Right)

	assert.EqualValues(t, filterEqualityMatch, f.Children[1].Op)
	assert.EqualValues(t, "department", f.Children[1].Left)
	assert.EqualValues(t, "finance", f.Children[1].Right)
}

func TestCompileInequalityMatch(t *testing.T) {
	f, err := CompileFilter(
		`(&(|(datacenter=irvine)(!(datacenter=houston)))(department=finance))`)
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, filterAnd, f.Op)

	assert.EqualValues(t, filterOr, f.Children[0].Op)

	assert.EqualValues(t, filterEqualityMatch,
		f.Children[0].Children[0].Op)
	assert.EqualValues(t, "datacenter",
		f.Children[0].Children[0].Left)
	assert.EqualValues(t, "irvine",
		f.Children[0].Children[0].Right)

	assert.EqualValues(t, filterNot, f.Children[0].Children[1].Op)

	assert.EqualValues(t, filterEqualityMatch,
		f.Children[0].Children[1].Children[0].Op)
	assert.EqualValues(t, "datacenter",
		f.Children[0].Children[1].Children[0].Left)
	assert.EqualValues(t, "houston",
		f.Children[0].Children[1].Children[0].Right)

	assert.EqualValues(t, filterEqualityMatch, f.Children[1].Op)
	assert.EqualValues(t, "department", f.Children[1].Left)
	assert.EqualValues(t, "finance", f.Children[1].Right)
}
