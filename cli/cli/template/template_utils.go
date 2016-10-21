package template

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// indirect is taken from 'text/template/exec.go'
func indirect(v reflect.Value) (rv reflect.Value, isNil bool) {
	for ; v.Kind() == reflect.Ptr ||
		v.Kind() == reflect.Interface; v = v.Elem() {
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

func (p pairList) Swap(i, j int) {
	p.Pairs[i], p.Pairs[j] = p.Pairs[j], p.Pairs[i]
}
func (p pairList) Len() int { return len(p.Pairs) }
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

func evaluateSubElem(
	obj reflect.Value,
	elemName string) (reflect.Value, error) {

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
			return zero, fmt.Errorf(
				"%s is an unexported method of type %s", elemName, typ)
		}
		// struct pointer has one receiver argument and interface doesn't have
		// an argument
		if mt.Type.NumIn() > 1 ||
			mt.Type.NumOut() == 0 || mt.Type.NumOut() > 2 {
			return zero, fmt.Errorf(
				"%s is a method of type %s but doesn't satisfy requirements",
				elemName, typ)
		}
		if mt.Type.NumOut() == 1 && mt.Type.Out(0).Implements(errorType) {
			return zero, fmt.Errorf(
				"%s is a method of type %s but doesn't satisfy requirements",
				elemName, typ)
		}
		if mt.Type.NumOut() == 2 && !mt.Type.Out(1).Implements(errorType) {
			return zero, fmt.Errorf(
				"%s is a method of type %s but doesn't satisfy requirements",
				elemName, typ)
		}
		res := objPtr.Method(mt.Index).Call([]reflect.Value{})
		if len(res) == 2 && !res[1].IsNil() {
			return zero, fmt.Errorf(
				"error at calling a method %s of type %s: %s",
				elemName, typ, res[1].Interface().(error))
		}
		return res[0], nil
	}

	// elemName isn't a method so next start to check whether it is
	// a struct field or a map value. In both cases, it mustn't be
	// a nil value
	if isNil {
		return zero, fmt.Errorf(
			"can't evaluate a nil pointer of type %s by "+
				"a struct field or map key name %s", typ, elemName)
	}
	switch obj.Kind() {
	case reflect.Struct:
		ft, ok := obj.Type().FieldByName(elemName)
		if ok {
			if ft.PkgPath != "" && !ft.Anonymous {
				return zero, fmt.Errorf(
					"%s is an unexported field of struct type %s", elemName, typ)
			}
			return obj.FieldByIndex(ft.Index), nil
		}
		return zero, fmt.Errorf(
			"%s isn't a field of struct type %s", elemName, typ)
	case reflect.Map:
		kv := reflect.ValueOf(elemName)
		if kv.Type().AssignableTo(obj.Type().Key()) {
			return obj.MapIndex(kv), nil
		}
		return zero, fmt.Errorf("%s isn't a key of map type %s", elemName, typ)
	}
	return zero, fmt.Errorf(
		"%s is neither a struct field, a method nor "+
			"a map element of type %s", elemName, typ)
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

// parseWhereArgs parses the end arguments to the where function.  Return a
// match value and an operator, if one is defined.
func parseWhereArgs(
	args ...interface{}) (mv reflect.Value, op string, err error) {

	switch len(args) {
	case 1:
		mv = reflect.ValueOf(args[0])
	case 2:
		var ok bool
		if op, ok = args[0].(string); !ok {
			err = errors.New("operator argument must be string type")
			return
		}
		op = strings.TrimSpace(strings.ToLower(op))
		mv = reflect.ValueOf(args[1])
	default:
		err = errors.New("can't evaluate the array by no match argument " +
			"or more than or equal to two arguments")
	}
	return
}

// checkWhereArray handles the where-matching logic when the seqv value is an
// Array or Slice.
func checkWhereArray(
	seqv, kv, mv reflect.Value,
	path []string, op string) (interface{}, error) {

	rv := reflect.MakeSlice(seqv.Type(), 0, 0)
	for i := 0; i < seqv.Len(); i++ {
		var vvv reflect.Value
		rvv := seqv.Index(i)
		if kv.Kind() == reflect.String {
			vvv = rvv
			for _, elemName := range path {
				var err error
				vvv, err = evaluateSubElem(vvv, elemName)
				if err != nil {
					return nil, err
				}
			}
		} else {
			vv, _ := indirect(rvv)
			if vv.Kind() == reflect.Map &&
				kv.Type().AssignableTo(vv.Type().Key()) {
				vvv = vv.MapIndex(kv)
			}
		}

		if ok, err := checkCondition(vvv, mv, op); ok {
			rv = reflect.Append(rv, rvv)
		} else if err != nil {
			return nil, err
		}
	}
	return rv.Interface(), nil
}

// checkWhereMap handles the where-matching logic when the seqv value is a Map.
func checkWhereMap(seqv, kv, mv reflect.Value,
	path []string, op string) (interface{}, error) {

	rv := reflect.MakeMap(seqv.Type())
	keys := seqv.MapKeys()
	for _, k := range keys {
		elemv := seqv.MapIndex(k)
		switch elemv.Kind() {
		case reflect.Array, reflect.Slice:
			r, err := checkWhereArray(elemv, kv, mv, path, op)
			if err != nil {
				return nil, err
			}

			switch rr := reflect.ValueOf(r); rr.Kind() {
			case reflect.Slice:
				if rr.Len() > 0 {
					rv.SetMapIndex(k, elemv)
				}
			}
		case reflect.Interface:
			elemvv, isNil := indirect(elemv)
			if isNil {
				continue
			}

			switch elemvv.Kind() {
			case reflect.Array, reflect.Slice:
				r, err := checkWhereArray(elemvv, kv, mv, path, op)
				if err != nil {
					return nil, err
				}

				switch rr := reflect.ValueOf(r); rr.Kind() {
				case reflect.Slice:
					if rr.Len() > 0 {
						rv.SetMapIndex(k, elemv)
					}
				}
			}
		}
	}
	return rv.Interface(), nil
}

func checkCondition(v, mv reflect.Value, op string) (bool, error) {
	v, vIsNil := indirect(v)
	if !v.IsValid() {
		vIsNil = true
	}
	mv, mvIsNil := indirect(mv)
	if !mv.IsValid() {
		mvIsNil = true
	}
	if vIsNil || mvIsNil {
		switch op {
		case "", "=", "==", "eq":
			return vIsNil == mvIsNil, nil
		case "!=", "<>", "ne":
			return vIsNil != mvIsNil, nil
		}
		return false, nil
	}

	if v.Kind() == reflect.Bool && mv.Kind() == reflect.Bool {
		switch op {
		case "", "=", "==", "eq":
			return v.Bool() == mv.Bool(), nil
		case "!=", "<>", "ne":
			return v.Bool() != mv.Bool(), nil
		}
		return false, nil
	}

	var ivp, imvp *int64
	var svp, smvp *string
	var slv, slmv interface{}
	var ima []int64
	var sma []string
	if mv.Type() == v.Type() {
		switch v.Kind() {
		case reflect.Int, reflect.Int8,
			reflect.Int16, reflect.Int32, reflect.Int64:
			iv := v.Int()
			ivp = &iv
			imv := mv.Int()
			imvp = &imv
		case reflect.String:
			sv := v.String()
			svp = &sv
			smv := mv.String()
			smvp = &smv
		case reflect.Struct:
			switch v.Type() {
			case timeType:
				iv := toTimeUnix(v)
				ivp = &iv
				imv := toTimeUnix(mv)
				imvp = &imv
			}
		case reflect.Array, reflect.Slice:
			slv = v.Interface()
			slmv = mv.Interface()
		}
	} else {
		if mv.Kind() != reflect.Array && mv.Kind() != reflect.Slice {
			return false, nil
		}

		if mv.Len() == 0 {
			return false, nil
		}

		if v.Kind() != reflect.Interface &&
			mv.Type().Elem().Kind() != reflect.Interface &&
			mv.Type().Elem() != v.Type() {
			return false, nil
		}
		switch v.Kind() {
		case reflect.Int, reflect.Int8,
			reflect.Int16, reflect.Int32, reflect.Int64:
			iv := v.Int()
			ivp = &iv
			for i := 0; i < mv.Len(); i++ {
				if anInt := toInt(mv.Index(i)); anInt != -1 {
					ima = append(ima, anInt)
				}

			}
		case reflect.String:
			sv := v.String()
			svp = &sv
			for i := 0; i < mv.Len(); i++ {
				if aString := toString(mv.Index(i)); aString != "" {
					sma = append(sma, aString)
				}
			}
		case reflect.Struct:
			switch v.Type() {
			case timeType:
				iv := toTimeUnix(v)
				ivp = &iv
				for i := 0; i < mv.Len(); i++ {
					ima = append(ima, toTimeUnix(mv.Index(i)))
				}
			}
		}
	}

	switch op {
	case "", "=", "==", "eq":
		if ivp != nil && imvp != nil {
			return *ivp == *imvp, nil
		} else if svp != nil && smvp != nil {
			return *svp == *smvp, nil
		}
	case "!=", "<>", "ne":
		if ivp != nil && imvp != nil {
			return *ivp != *imvp, nil
		} else if svp != nil && smvp != nil {
			return *svp != *smvp, nil
		}
	case ">=", "ge":
		if ivp != nil && imvp != nil {
			return *ivp >= *imvp, nil
		} else if svp != nil && smvp != nil {
			return *svp >= *smvp, nil
		}
	case ">", "gt":
		if ivp != nil && imvp != nil {
			return *ivp > *imvp, nil
		} else if svp != nil && smvp != nil {
			return *svp > *smvp, nil
		}
	case "<=", "le":
		if ivp != nil && imvp != nil {
			return *ivp <= *imvp, nil
		} else if svp != nil && smvp != nil {
			return *svp <= *smvp, nil
		}
	case "<", "lt":
		if ivp != nil && imvp != nil {
			return *ivp < *imvp, nil
		} else if svp != nil && smvp != nil {
			return *svp < *smvp, nil
		}
	case "in", "not in":
		var r bool
		if ivp != nil && len(ima) > 0 {
			r = in(ima, *ivp)
		} else if svp != nil {
			if len(sma) > 0 {
				r = in(sma, *svp)
			} else if smvp != nil {
				r = in(*smvp, *svp)
			}
		} else {
			return false, nil
		}
		if op == "not in" {
			return !r, nil
		}
		return r, nil
	case "intersect":
		r, err := intersect(slv, slmv)
		if err != nil {
			return false, err
		}

		if reflect.TypeOf(r).Kind() == reflect.Slice {
			s := reflect.ValueOf(r)
			if s.Len() > 0 {
				return true, nil
			}
			return false, nil
		}
		return false, errors.New("invalid intersect values")
	default:
		return false, errors.New("no such operator")
	}
	return false, nil
}

// toInt returns the int value if possible, -1 if not.
func toInt(v reflect.Value) int64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int8,
		reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Interface:
		return toInt(v.Elem())
	}
	return -1
}

// toString returns the string value if possible, "" if not.
func toString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Interface:
		return toString(v.Elem())
	}
	return ""
}

// in returns whether v is in the set l.  l may be an array or slice.
func in(l interface{}, v interface{}) bool {
	lv := reflect.ValueOf(l)
	vv := reflect.ValueOf(v)

	switch lv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < lv.Len(); i++ {
			lvv := lv.Index(i)
			lvv, isNil := indirect(lvv)
			if isNil {
				continue
			}
			switch lvv.Kind() {
			case reflect.String:
				if vv.Type() == lvv.Type() && vv.String() == lvv.String() {
					return true
				}
			case reflect.Int, reflect.Int8,
				reflect.Int16, reflect.Int32, reflect.Int64:
				switch vv.Kind() {
				case reflect.Int, reflect.Int8,
					reflect.Int16, reflect.Int32, reflect.Int64:
					if vv.Int() == lvv.Int() {
						return true
					}
				}
			case reflect.Float32, reflect.Float64:
				switch vv.Kind() {
				case reflect.Float32, reflect.Float64:
					if vv.Float() == lvv.Float() {
						return true
					}
				}
			}
		}
	case reflect.String:
		if vv.Type() == lv.Type() &&
			strings.Contains(lv.String(), vv.String()) {
			return true
		}
	}
	return false
}

// intersect returns the common elements in the given sets, l1 and l2.  l1 and
// l2 must be of the same type and may be either arrays or slices.
func intersect(l1, l2 interface{}) (interface{}, error) {
	if l1 == nil || l2 == nil {
		return make([]interface{}, 0), nil
	}

	l1v := reflect.ValueOf(l1)
	l2v := reflect.ValueOf(l2)

	switch l1v.Kind() {
	case reflect.Array, reflect.Slice:
		switch l2v.Kind() {
		case reflect.Array, reflect.Slice:
			r := reflect.MakeSlice(l1v.Type(), 0, 0)
			for i := 0; i < l1v.Len(); i++ {
				l1vv := l1v.Index(i)
				for j := 0; j < l2v.Len(); j++ {
					l2vv := l2v.Index(j)
					switch l1vv.Kind() {
					case reflect.String:
						if l1vv.Type() == l2vv.Type() &&
							l1vv.String() == l2vv.String() &&
							!in(r.Interface(), l2vv.Interface()) {
							r = reflect.Append(r, l2vv)
						}
					case reflect.Int, reflect.Int8,
						reflect.Int16, reflect.Int32, reflect.Int64:
						switch l2vv.Kind() {
						case reflect.Int, reflect.Int8,
							reflect.Int16, reflect.Int32, reflect.Int64:
							if l1vv.Int() == l2vv.Int() &&
								!in(r.Interface(), l2vv.Interface()) {
								r = reflect.Append(r, l2vv)
							}
						}
					case reflect.Float32, reflect.Float64:
						switch l2vv.Kind() {
						case reflect.Float32, reflect.Float64:
							if l1vv.Float() == l2vv.Float() &&
								!in(r.Interface(), l2vv.Interface()) {
								r = reflect.Append(r, l2vv)
							}
						}
					}
				}
			}
			return r.Interface(), nil
		default:
			return nil, errors.New(
				"can't iterate over " + reflect.ValueOf(l2).Type().String())
		}
	default:
		return nil, errors.New(
			"can't iterate over " + reflect.ValueOf(l1).Type().String())
	}
}
