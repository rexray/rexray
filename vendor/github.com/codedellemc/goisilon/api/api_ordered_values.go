package api

import (
	"bytes"
	"io"
	"net/url"
	"strings"
)

var chrs = []byte(`,=&+%`)

const (
	chComa = iota
	chEqul
	chAmps
	chPlus
	chPrct
)

// OrderedValues maps a string key to a list of values and preserves insertion
// order of the keys. It is typically used for query parameters and form values.
// Unlike in the http.Header map, the keys in a Values map are case-sensitive.
type OrderedValues [][][]byte

// NewOrderedValues returns a new OrderedValues object.
func NewOrderedValues(vals [][]string) OrderedValues {
	if len(vals) == 0 {
		return nil
	}
	var nov OrderedValues
	for i := range vals {
		var a [][]byte
		for j := range vals[i] {
			a = append(a, []byte(vals[i][j]))
		}
		nov = append(nov, a)
	}
	return nov
}

// StringAdd adds the value to key. It appends to any existing values
// associated with key.
func (v *OrderedValues) StringAdd(key, val string) {
	if len(key) == 0 {
		return
	}
	if len(val) == 0 {
		v.Add([]byte(key), chrs[0:0])
	} else {
		v.Add([]byte(key), []byte(val))
	}
}

// Add adds the value to key. It appends to any existing values associated with
// key.
func (v *OrderedValues) Add(key, val []byte) {
	for i, j := range *v {
		if len(j) > 0 && bytes.Equal(j[0], key) {
			if len(val) > 0 {
				(*v)[i] = append((*v)[i], val)
			}
			return
		}
	}
	if len(val) == 0 {
		*v = append(*v, [][]byte{key})
	} else {
		*v = append(*v, [][]byte{key, val})
	}
}

// StringSet sets the key to value. It replaces any existing values.
func (v *OrderedValues) StringSet(key, val string) {
	if len(key) == 0 {
		return
	}
	if len(val) == 0 {
		v.Set([]byte(key), chrs[0:0])
	} else {
		v.Set([]byte(key), []byte(val))
	}
}

// Set sets the key to value. It replaces any existing values.
func (v *OrderedValues) Set(key, val []byte) {
	for i, j := range *v {
		if len(j) > 0 && bytes.Equal(j[0], key) {
			if len(val) == 0 {
				(*v)[i] = [][]byte{j[0]}
			} else {
				(*v)[i] = [][]byte{j[0], val}
			}
			return
		}
	}
	if len(val) == 0 {
		*v = append(*v, [][]byte{key})
	} else {
		*v = append(*v, [][]byte{key, val})
	}
}

// StringGet gets the first value associated with the given key. If there are no
// values associated with the key, Get returns the empty string. To access
// multiple values, use the array directly.
func (v *OrderedValues) StringGet(key string) string {
	if len(key) == 0 {
		return ""
	}
	return string(v.Get([]byte(key)))
}

// Get gets the first value associated with the given key. If there are no
// values associated with the key, Get returns the empty string. To access
// multiple values, use the array directly.
func (v *OrderedValues) Get(key []byte) []byte {
	for _, j := range *v {
		if len(j) > 0 && bytes.Equal(j[0], key) {
			if len(j) > 1 {
				return j[1]
			}
			return nil
		}
	}
	return nil
}

// StringGetOk returns a flag indicating whether or not the key exists. The
// StringGet method can return an empty value for keys that do not have values,
// so it cannot be trusted to indicate the existence of a key.
func (v *OrderedValues) StringGetOk(key string) (string, bool) {
	if len(key) == 0 {
		return "", false
	}
	val, ok := v.GetOk([]byte(key))
	if !ok {
		return "", false
	}
	return string(val), true
}

// GetOk returns a flag indicating whether or not the key exists. The Get
// method can return an empty value for keys that do not have values, so it
// cannot be trusted to indicate the existence of a key.
func (v *OrderedValues) GetOk(key []byte) ([]byte, bool) {
	for _, j := range *v {
		if len(j) > 0 && bytes.Equal(j[0], key) {
			if len(j) > 1 {
				return j[1], true
			}
			return nil, true
		}
	}
	return nil, false
}

// StringDel deletes the values associated with key.
func (v *OrderedValues) StringDel(key string) {
	v.Del([]byte(key))
}

// Del deletes the values associated with key.
func (v *OrderedValues) Del(key []byte) {
	var (
		i  int
		ok bool
		j  [][]byte
	)
	for i, j = range *v {
		if len(j) > 0 && bytes.Equal(j[0], key) {
			ok = true
			break
		}
	}
	if !ok {
		return
	}
	copy((*v)[i:], (*v)[i+1:])
	(*v)[len(*v)-1] = nil
	*v = (*v)[:len(*v)-1]
}

// Encode encodes the values into “URL encoded” form ("bar=baz&foo=quux")
// using insertion order.
func (v *OrderedValues) Encode() string {
	buf := &bytes.Buffer{}
	v.EncodeTo(buf)
	return buf.String()
}

// EncodeTo encodes the values into “URL encoded” form ("bar=baz&foo=quux")
// using insertion order.
func (v *OrderedValues) EncodeTo(w io.Writer) error {
	first := true
	for _, j := range *v {
		if len(j) == 0 {
			continue
		}
		if !first {
			if _, err := w.Write(chrs[chAmps : chAmps+1]); err != nil {
				return err
			}
		} else {
			first = false
		}
		if _, err := w.Write(j[0]); err != nil {
			return err
		}
		if len(j) == 1 {
			continue
		}
		if _, err := w.Write(chrs[chEqul : chEqul+1]); err != nil {
			return err
		}
		for e := 1; e < len(j); e++ {
			if e > 1 {
				if _, err := w.Write(chrs[chComa : chComa+1]); err != nil {
					return err
				}
			}
			if err := escapeTo(w, j[e]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *OrderedValues) String() string {
	return v.Encode()
}

// ParseQuery parses the URL-encoded query string and returns an OrderedValues
// object. ParseQuery may return nil if no valid query parameters are found. The
// return error object is set to the first encountered decoding error, if any.
func ParseQuery(query string) (OrderedValues, error) {
	ov := OrderedValues{}
	for query != "" {
		if err := parseQuery(&ov, &query); err != nil {
			return nil, err
		}
	}
	return ov, nil
}

func parseQuery(m *OrderedValues, query *string) error {
	var (
		err   error
		value string
		key   = *query
	)
	if i := strings.IndexAny(key, "&;"); i >= 0 {
		key, *query = key[:i], key[i+1:]
	} else {
		*query = ""
	}
	if key == "" {
		return nil
	}
	if i := strings.Index(key, "="); i >= 0 {
		key, value = key[:i], key[i+1:]
	}
	if key, err = url.QueryUnescape(key); err != nil {
		return err
	}
	if value, err = url.QueryUnescape(value); err != nil {
		return err
	}
	(*m).StringAdd(key, value)
	return nil
}

var hexChars = []byte{
	'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
}

func escapeTo(w io.Writer, s []byte) error {

	var (
		hexCount   = 0
		spaceCount = 0
	)

	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			if c == ' ' {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}
	if spaceCount == 0 && hexCount == 0 {
		if _, err := w.Write(s); err != nil {
			return err
		}
		return nil
	}
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ':
			if _, err := w.Write(chrs[chPlus : chPlus+1]); err != nil {
				return err
			}
		case shouldEscape(c):
			if _, err := w.Write(chrs[chPrct : chPrct+1]); err != nil {
				return err
			}
			c4, c15 := c>>4, c&15
			if _, err := w.Write(hexChars[c4 : c4+1]); err != nil {
				return err
			}
			if _, err := w.Write(hexChars[c15 : c15+1]); err != nil {
				return err
			}
		default:
			if _, err := w.Write(s[i : i+1]); err != nil {
				return err
			}
		}
	}

	return nil
}

// shouldEscape returns true if the specified character should be escaped when
// appearing in a URL string, according to RFC 3986.
//
// Please be informed that for now shouldEscape does not check all
// reserved characters correctly. See golang.org/issue/5684.
func shouldEscape(c byte) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}
	switch c {
	case '-', '_', '.', '~':
		// §2.3 Unreserved characters (mark)
		return false
	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@':
		return true
	}
	// Everything else must be escaped.
	return true
}
