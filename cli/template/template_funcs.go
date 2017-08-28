package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"strings"
)

// sortSeq returns a sorted sequence.
func sortSeq(seq interface{}, args ...interface{}) (interface{}, error) {
	if seq == nil {
		return nil, errors.New("sequence must be provided")
	}

	seqv := reflect.ValueOf(seq)
	seqv, isNil := indirect(seqv)
	if isNil {
		return nil, errors.New("can't iterate over a nil value")
	}

	switch seqv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		// ok
	default:
		return nil, errors.New(
			"can't sort " + reflect.ValueOf(seq).Type().String())
	}

	// Create a list of pairs that will be used to do the sort
	p := pairList{SortAsc: true, SliceType: reflect.SliceOf(seqv.Type().Elem())}
	p.Pairs = make([]pair, seqv.Len())

	var sortByField string
	for i, l := range args {
		dStr, err := ToStringE(l)
		switch {
		case i == 0 && err != nil:
			sortByField = ""
		case i == 0 && err == nil:
			sortByField = dStr
		case i == 1 && err == nil && dStr == "desc":
			p.SortAsc = false
		case i == 1:
			p.SortAsc = true
		}
	}
	path := strings.Split(strings.Trim(sortByField, "."), ".")

	switch seqv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < seqv.Len(); i++ {
			p.Pairs[i].Value = seqv.Index(i)
			if sortByField == "" || sortByField == "value" {
				p.Pairs[i].Key = p.Pairs[i].Value
			} else {
				v := p.Pairs[i].Value
				var err error
				for _, elemName := range path {
					v, err = evaluateSubElem(v, elemName)
					if err != nil {
						return nil, err
					}
				}
				p.Pairs[i].Key = v
			}
		}

	case reflect.Map:
		keys := seqv.MapKeys()
		for i := 0; i < seqv.Len(); i++ {
			p.Pairs[i].Value = seqv.MapIndex(keys[i])
			if sortByField == "" {
				p.Pairs[i].Key = keys[i]
			} else if sortByField == "value" {
				p.Pairs[i].Key = p.Pairs[i].Value
			} else {
				v := p.Pairs[i].Value
				var err error
				for _, elemName := range path {
					v, err = evaluateSubElem(v, elemName)
					if err != nil {
						return nil, err
					}
				}
				p.Pairs[i].Key = v
			}
		}
	}
	return p.sort(), nil
}

// where returns a filtered subset of a given data type.
func where(seq, key interface{}, args ...interface{}) (interface{}, error) {
	seqv, isNil := indirect(reflect.ValueOf(seq))
	if isNil {
		return nil, errors.New("can't iterate over a nil value of type " +
			reflect.ValueOf(seq).Type().String())
	}

	mv, op, err := parseWhereArgs(args...)
	if err != nil {
		return nil, err
	}

	var path []string
	kv := reflect.ValueOf(key)
	if kv.Kind() == reflect.String {
		path = strings.Split(strings.Trim(kv.String(), "."), ".")
	}

	switch seqv.Kind() {
	case reflect.Array, reflect.Slice:
		return checkWhereArray(seqv, kv, mv, path, op)
	case reflect.Map:
		return checkWhereMap(seqv, kv, mv, path, op)
	default:
		return nil, fmt.Errorf("can't iterate over %v", seq)
	}
}

// jsonify encodes a given object to JSON.
func jsonify(v interface{}) (template.HTML, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return template.HTML(b), nil
}

// jsonpify encodes a given object to pretty JSON.
func jsonpify(v interface{}) (template.HTML, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return template.HTML(b), nil
}
