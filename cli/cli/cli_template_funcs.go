package cli

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
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
		return nil, errors.New("can't sort " + reflect.ValueOf(seq).Type().String())
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

// indirect is taken from 'text/template/exec.go'
func indirect(v reflect.Value) (rv reflect.Value, isNil bool) {
	for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return v, true
		}
		if v.Kind() == reflect.Interface && v.NumMethod() > 0 {
			break
		}
	}
	return v, false
}

// Credit for pair sorting method goes to Andrew Gerrand
// https://groups.google.com/forum/#!topic/golang-nuts/FT7cjmcL7gw
// A data structure to hold a key/value pair.
type pair struct {
	Key   reflect.Value
	Value reflect.Value
}

// A slice of pairs that implements sort.Interface to sort by Value.
type pairList struct {
	Pairs     []pair
	SortAsc   bool
	SliceType reflect.Type
}

func (p pairList) Swap(i, j int) { p.Pairs[i], p.Pairs[j] = p.Pairs[j], p.Pairs[i] }
func (p pairList) Len() int      { return len(p.Pairs) }
func (p pairList) Less(i, j int) bool {
	iv := p.Pairs[i].Key
	jv := p.Pairs[j].Key

	if iv.IsValid() {
		if jv.IsValid() {
			// can only call Interface() on valid reflect Values
			return lt(iv.Interface(), jv.Interface())
		}
		// if j is invalid, test i against i's zero value
		return lt(iv.Interface(), reflect.Zero(iv.Type()))
	}

	if jv.IsValid() {
		// if i is invalid, test j against j's zero value
		return lt(reflect.Zero(jv.Type()), jv.Interface())
	}

	return false
}

// sorts a pairList and returns a slice of sorted values
func (p pairList) sort() interface{} {
	if p.SortAsc {
		sort.Sort(p)
	} else {
		sort.Sort(sort.Reverse(p))
	}
	sorted := reflect.MakeSlice(p.SliceType, len(p.Pairs), len(p.Pairs))
	for i, v := range p.Pairs {
		sorted.Index(i).Set(v.Value)
	}

	return sorted.Interface()
}

func evaluateSubElem(obj reflect.Value, elemName string) (reflect.Value, error) {
	if !obj.IsValid() {
		return zero, errors.New("can't evaluate an invalid value")
	}
	typ := obj.Type()
	obj, isNil := indirect(obj)

	// first, check whether obj has a method. In this case, obj is
	// an interface, a struct or its pointer. If obj is a struct,
	// to check all T and *T method, use obj pointer type Value
	objPtr := obj
	if objPtr.Kind() != reflect.Interface && objPtr.CanAddr() {
		objPtr = objPtr.Addr()
	}
	mt, ok := objPtr.Type().MethodByName(elemName)
	if ok {
		if mt.PkgPath != "" {
			return zero, fmt.Errorf("%s is an unexported method of type %s", elemName, typ)
		}
		// struct pointer has one receiver argument and interface doesn't have an argument
		if mt.Type.NumIn() > 1 || mt.Type.NumOut() == 0 || mt.Type.NumOut() > 2 {
			return zero, fmt.Errorf("%s is a method of type %s but doesn't satisfy requirements", elemName, typ)
		}
		if mt.Type.NumOut() == 1 && mt.Type.Out(0).Implements(errorType) {
			return zero, fmt.Errorf("%s is a method of type %s but doesn't satisfy requirements", elemName, typ)
		}
		if mt.Type.NumOut() == 2 && !mt.Type.Out(1).Implements(errorType) {
			return zero, fmt.Errorf("%s is a method of type %s but doesn't satisfy requirements", elemName, typ)
		}
		res := objPtr.Method(mt.Index).Call([]reflect.Value{})
		if len(res) == 2 && !res[1].IsNil() {
			return zero, fmt.Errorf("error at calling a method %s of type %s: %s", elemName, typ, res[1].Interface().(error))
		}
		return res[0], nil
	}

	// elemName isn't a method so next start to check whether it is
	// a struct field or a map value. In both cases, it mustn't be
	// a nil value
	if isNil {
		return zero, fmt.Errorf("can't evaluate a nil pointer of type %s by a struct field or map key name %s", typ, elemName)
	}
	switch obj.Kind() {
	case reflect.Struct:
		ft, ok := obj.Type().FieldByName(elemName)
		if ok {
			if ft.PkgPath != "" && !ft.Anonymous {
				return zero, fmt.Errorf("%s is an unexported field of struct type %s", elemName, typ)
			}
			return obj.FieldByIndex(ft.Index), nil
		}
		return zero, fmt.Errorf("%s isn't a field of struct type %s", elemName, typ)
	case reflect.Map:
		kv := reflect.ValueOf(elemName)
		if kv.Type().AssignableTo(obj.Type().Key()) {
			return obj.MapIndex(kv), nil
		}
		return zero, fmt.Errorf("%s isn't a key of map type %s", elemName, typ)
	}
	return zero, fmt.Errorf("%s is neither a struct field, a method nor a map element of type %s", elemName, typ)
}

// lt returns the boolean truth of arg1 < arg2.
func lt(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left < right
}

func compareGetFloat(a interface{}, b interface{}) (float64, float64) {
	var left, right float64
	var leftStr, rightStr *string
	av := reflect.ValueOf(a)

	switch av.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		left = float64(av.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		left = float64(av.Int())
	case reflect.Float32, reflect.Float64:
		left = av.Float()
	case reflect.String:
		var err error
		left, err = strconv.ParseFloat(av.String(), 64)
		if err != nil {
			str := av.String()
			leftStr = &str
		}
	case reflect.Struct:
		switch av.Type() {
		case timeType:
			left = float64(toTimeUnix(av))
		}
	}

	bv := reflect.ValueOf(b)

	switch bv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		right = float64(bv.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		right = float64(bv.Int())
	case reflect.Float32, reflect.Float64:
		right = bv.Float()
	case reflect.String:
		var err error
		right, err = strconv.ParseFloat(bv.String(), 64)
		if err != nil {
			str := bv.String()
			rightStr = &str
		}
	case reflect.Struct:
		switch bv.Type() {
		case timeType:
			right = float64(toTimeUnix(bv))
		}
	}

	switch {
	case leftStr == nil || rightStr == nil:
	case *leftStr < *rightStr:
		return 0, 1
	case *leftStr > *rightStr:
		return 1, 0
	default:
		return 0, 0
	}

	return left, right
}

var (
	zero      reflect.Value
	errorType = reflect.TypeOf((*error)(nil)).Elem()
	timeType  = reflect.TypeOf((*time.Time)(nil)).Elem()
)

func toTimeUnix(v reflect.Value) int64 {
	if v.Kind() == reflect.Interface {
		return toTimeUnix(v.Elem())
	}
	if v.Type() != timeType {
		panic("coding error: argument must be time.Time type reflect Value")
	}
	return v.MethodByName("Unix").Call([]reflect.Value{})[0].Int()
}

// ToStringE casts an empty interface to a string.
func ToStringE(i interface{}) (string, error) {
	i = indirectToStringerOrError(i)

	switch s := i.(type) {
	case string:
		return s, nil
	case bool:
		return strconv.FormatBool(s), nil
	case float64:
		return strconv.FormatFloat(i.(float64), 'f', -1, 64), nil
	case int64:
		return strconv.FormatInt(i.(int64), 10), nil
	case int:
		return strconv.FormatInt(int64(i.(int)), 10), nil
	case []byte:
		return string(s), nil
	case nil:
		return "", nil
	case fmt.Stringer:
		return s.String(), nil
	case error:
		return s.Error(), nil
	default:
		return "", fmt.Errorf("Unable to Cast %#v to string", i)
	}
}

// From html/template/content.go
// Copyright 2011 The Go Authors. All rights reserved.
// indirectToStringerOrError returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil) or an implementation of fmt.Stringer
// or error,
func indirectToStringerOrError(a interface{}) interface{} {
	if a == nil {
		return nil
	}

	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	var fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	v := reflect.ValueOf(a)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

var funcMap = template.FuncMap{
	"sort": sortSeq,
}
