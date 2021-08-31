/*
Package filters is a piece of thievery as the LDAP filter parsing code was
lifted serepticiously from
https://github.com/tonnerre/go-ldap/blob/master/ldap.go. The code was reused
this way as opposed to imports in order to modify it heavily for the needs of
the project.
*/
package filters

import (
	"bytes"
	"net/url"

	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

const (
	filterAnd               = types.FilterAnd
	filterOr                = types.FilterOr
	filterNot               = types.FilterNot
	filterPresent           = types.FilterPresent
	filterEqualityMatch     = types.FilterEqualityMatch
	filterSubstrings        = types.FilterSubstrings
	filterSubstringsPrefix  = types.FilterSubstringsPrefix
	filterSubstringsPostfix = types.FilterSubstringsPostfix
	filterGreaterOrEqual    = types.FilterGreaterOrEqual
	filterLessOrEqual       = types.FilterLessOrEqual
	filterApproxMatch       = types.FilterApproxMatch
)

var filterMap = map[types.FilterOperator]string{
	filterAnd:            "And",
	filterOr:             "Or",
	filterNot:            "Not",
	filterEqualityMatch:  "Equality Match",
	filterSubstrings:     "Substrings",
	filterGreaterOrEqual: "Greater Or Equal",
	filterLessOrEqual:    "Less Or Equal",
	filterPresent:        "Present",
	filterApproxMatch:    "Approx Match",
}

var (
	errCharZeroNotLParen = goof.New("filter does not start with an '()'")
	errUnexpectedEOF     = goof.New("unexpected end of filter")
	errCompile           = goof.New("error compiling filter")
	errParse             = goof.New("error parsing filter")
)

// CompileFilter compiles a filter string.
func CompileFilter(s string) (*types.Filter, error) {

	es, err := url.QueryUnescape(s)
	if err != nil {
		return nil, err
	}
	s = es

	if len(s) == 0 || s[0] != '(' {
		return nil, errCharZeroNotLParen
	}

	f, pos, err := compileFilter(s, 1)
	if err != nil {
		return nil, err
	}

	if pos != len(s) {
		return nil, goof.WithField(
			"extra", s[pos:],
			"finished compiling filter with extra at end")
	}

	return f, nil
}

func compileFilterSet(
	s string, pos int, parent *types.Filter) (int, error) {

	for pos < len(s) && s[pos] == '(' {
		child, newPos, err := compileFilter(s, pos+1)
		if err != nil {
			return pos, err
		}
		pos = newPos
		parent.Children = append(parent.Children, child)
	}

	if pos == len(s) {
		return pos, errUnexpectedEOF
	}

	return pos + 1, nil
}

func compileFilter(s string, pos int) (*types.Filter, int, error) {

	switch s[pos] {
	case '(':
		f, newPos, err := compileFilter(s, pos+1)
		newPos++
		return f, newPos, err

	case '&':
		f := &types.Filter{Op: filterAnd}
		newPos, err := compileFilterSet(s, pos+1, f)
		return f, newPos, err

	case '|':
		f := &types.Filter{Op: filterOr}
		newPos, err := compileFilterSet(s, pos+1, f)
		return f, newPos, err

	case '!':
		f := &types.Filter{Op: filterNot}
		child, newPos, err := compileFilter(s, pos+1)
		f.Children = append(f.Children, child)
		return f, newPos, err

	default:

		var (
			f          *types.Filter
			abuf, cbuf bytes.Buffer
			newPos     = pos
		)

		for newPos < len(s) && s[newPos] != ')' {
			switch {
			case f != nil:
				if err := cbuf.WriteByte(s[newPos]); err != nil {
					return nil, 0, err
				}

			case s[newPos] == '=':
				f = &types.Filter{Op: filterEqualityMatch}

			case s[newPos] == '>' && s[newPos+1] == '=':
				f = &types.Filter{Op: filterGreaterOrEqual}
				newPos++

			case s[newPos] == '<' && s[newPos+1] == '=':
				f = &types.Filter{Op: filterLessOrEqual}
				newPos++

			case s[newPos] == '~' && s[newPos+1] == '=':
				f = &types.Filter{Op: filterApproxMatch}
				newPos++

			case f == nil:
				if err := abuf.WriteByte(s[newPos]); err != nil {
					return nil, 0, err
				}
			}
			newPos++
		}

		if newPos == len(s) {
			return f, newPos, errUnexpectedEOF
		}

		if f == nil {
			return nil, 0, errParse
		}

		f.Left = abuf.String()

		var (
			cbyt = cbuf.Bytes()
			cstr = cbuf.String()
			clen = len(cbyt)
			cfch = cbyt[clen-1]
		)

		switch {
		case f.Op == filterEqualityMatch && cstr == "*":
			f.Op = filterPresent

		case f.Op == filterEqualityMatch &&
			cbyt[0] == '*' && cfch == '*' && clen > 2:
			f.Op = filterSubstrings
			f.Right = cstr[1 : clen-1]

		case f.Op == filterEqualityMatch && cbyt[0] == '*' && clen > 1:
			f.Op = filterSubstringsPrefix
			f.Right = cstr[1:]

		case f.Op == filterEqualityMatch && cfch == '*' && clen > 1:
			f.Op = filterSubstringsPostfix
			f.Right = cstr[:clen-1]

		default:
			f.Right = cstr

		}
		newPos++
		return f, newPos, nil
	}
}
