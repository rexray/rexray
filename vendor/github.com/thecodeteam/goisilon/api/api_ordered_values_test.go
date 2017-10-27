package api

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	b0 []byte
	b1 []byte
	b2 []byte
	b3 []byte
)

func BenchmarkByteArraysSingle(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b0 = []byte{'a'}
		b1 = []byte{'b'}
		b2 = []byte{'c'}
	}
}

func BenchmarkByteArraysMulti(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b3 = []byte{'a', 'b', 'c'}
	}
}

func BenchmarkOrderedValuesEncode(b *testing.B) {
	v := OrderedValues{
		{[]byte("query")},
		{[]byte("size"), []byte("2")},
		{[]byte("detail"), []byte("owner"), []byte("group")},
		{[]byte("info"), []byte("?")},
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.EncodeTo(os.Stderr)
		}
	})
}

func BenchmarkStringOrderedValuesEncode(b *testing.B) {
	v := NewOrderedValues([][]string{
		{"query"},
		{"size", "2"},
		{"detail", "owner", "group"},
		{"info", "?"},
	})
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.EncodeTo(os.Stderr)
		}
	})
}

func TestOrderedValuesEncode(t *testing.T) {
	v := OrderedValues{
		{[]byte("query")},
		{[]byte("size"), []byte("2")},
		{[]byte("detail"), []byte("owner"), []byte("group")},
		{[]byte("info"), []byte("?")},
	}
	assert.Equal(t, "query&size=2&detail=owner,group&info=%3F", v.Encode())
}

func TestStringOrderedValuesEncode(t *testing.T) {
	v := NewOrderedValues([][]string{
		{"query"},
		{"size", "2"},
		{"detail", "owner", "group"},
		{"info", "?"},
	})
	assert.Equal(t, "query&size=2&detail=owner,group&info=%3F", v.Encode())
}

func TestOrderedValuesGet(t *testing.T) {
	v := OrderedValues{
		{[]byte("query")},
		{[]byte("size"), []byte("2")},
		{[]byte("detail"), []byte("owner"), []byte("group")},
	}
	assert.Equal(t, "query&size=2&detail=owner,group", v.Encode())
	assert.Equal(t, []byte(nil), v.Get([]byte("query")))
	assert.Equal(t, []byte("2"), v.Get([]byte("size")))
	assert.Equal(t, []byte("owner"), v.Get([]byte("detail")))
}

func TestStringOrderedValuesGet(t *testing.T) {
	v := NewOrderedValues([][]string{
		{"query"},
		{"size", "2"},
		{"detail", "owner", "group"},
	})
	assert.Equal(t, "query&size=2&detail=owner,group", v.Encode())
	assert.Equal(t, "", v.StringGet("query"))
	assert.Equal(t, "2", v.StringGet("size"))
	assert.Equal(t, "owner", v.StringGet("detail"))
}

func TestOrderedValuesGetOk(t *testing.T) {
	v := OrderedValues{
		{[]byte("query")},
		{[]byte("size"), []byte("2")},
		{[]byte("detail"), []byte("owner"), []byte("group")},
	}
	assert.Equal(t, "query&size=2&detail=owner,group", v.Encode())
	val, ok := v.GetOk([]byte("query"))
	assert.Equal(t, []byte(nil), val)
	assert.True(t, ok)
	val, ok = v.GetOk([]byte("size"))
	assert.Equal(t, []byte("2"), val)
	assert.True(t, ok)
	val, ok = v.GetOk([]byte("detail"))
	assert.Equal(t, []byte("owner"), val)
	assert.True(t, ok)
	val, ok = v.GetOk([]byte("depth"))
	assert.Equal(t, []byte(nil), val)
	assert.False(t, ok)
}

func TestStringOrderedValuesGetOk(t *testing.T) {
	v := NewOrderedValues([][]string{
		{"query"},
		{"size", "2"},
		{"detail", "owner", "group"},
	})
	assert.Equal(t, "query&size=2&detail=owner,group", v.Encode())
	val, ok := v.StringGetOk("query")
	assert.Equal(t, "", val)
	assert.True(t, ok)
	val, ok = v.StringGetOk("size")
	assert.Equal(t, "2", val)
	assert.True(t, ok)
	val, ok = v.StringGetOk("detail")
	assert.Equal(t, "owner", val)
	assert.True(t, ok)
	val, ok = v.StringGetOk("depth")
	assert.Equal(t, "", val)
	assert.False(t, ok)
}

func TestOrderedValuesAdd(t *testing.T) {
	var v OrderedValues
	v.Add([]byte("query"), nil)
	assert.Equal(t, "query", v.Encode())
	v.Add([]byte("size"), []byte("2"))
	assert.Equal(t, "query&size=2", v.Encode())
	v.Add([]byte("detail"), []byte("owner"))
	assert.Equal(t, "query&size=2&detail=owner", v.Encode())
	v.Add([]byte("detail"), []byte("group"))
	assert.Equal(t, "query&size=2&detail=owner,group", v.Encode())
}

func TestOrderedValuesSet(t *testing.T) {
	var v OrderedValues
	v.Set([]byte("query"), nil)
	assert.Equal(t, "query", v.Encode())
	v.Set([]byte("size"), []byte("2"))
	assert.Equal(t, "query&size=2", v.Encode())
	v.Set([]byte("detail"), []byte("owner"))
	assert.Equal(t, "query&size=2&detail=owner", v.Encode())
	v.Set([]byte("detail"), []byte("group"))
	assert.Equal(t, "query&size=2&detail=group", v.Encode())
	v.Add([]byte("detail"), []byte("owner"))
	assert.Equal(t, "query&size=2&detail=group,owner", v.Encode())
}

func TestOrderedValuesDel(t *testing.T) {
	v := OrderedValues{
		{[]byte("query")},
		{[]byte("size"), []byte("2")},
		{[]byte("detail"), []byte("owner"), []byte("group")},
	}
	assert.Equal(t, "query&size=2&detail=owner,group", v.Encode())
	v.Del([]byte("size"))
	assert.Equal(t, "query&detail=owner,group", v.Encode())
	v.Del([]byte("query"))
	assert.Equal(t, "detail=owner,group", v.Encode())
	v.Del([]byte("detail"))
	assert.Equal(t, "", v.Encode())
}

func TestParseQuery(t *testing.T) {
	tf := func(qs string) {
		ov, err := ParseQuery(qs)
		assertNoError(t, err)
		assertLen(t, ov, 3)
		assertLen(t, ov[0], 2)
		assertLen(t, ov[1], 1)
		assertLen(t, ov[2], 2)
		assert.Equal(t, "Andrew Kutz", ov.StringGet("name"))
		assert.Equal(t, "", ov.StringGet("active"))
		_, ok := ov.StringGetOk("active")
		assert.True(t, ok)
		assert.Equal(t, "akutz", ov.StringGet("alias"))
		assert.Equal(t, "name=Andrew+Kutz&active&alias=akutz", ov.Encode())
	}
	tf("name=Andrew%20Kutz&active&alias=akutz")
	tf("name=Andrew+Kutz&active&alias=akutz")
}
