// The deep package provides deep comparisons with options.
package deep

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"unsafe"
)

// Comparer configures how comparisons are made.
type Comparer struct {
	// CompareUnexportedFields, when true, causes unexported fields to be
	// compared.
	CompareUnexportedFields bool
	// FloatPrecision sets the amount of mantissa precision, in bits, used when
	// comparing floats. A value of zero or less uses an exact equality
	// comparison.
	FloatPrecision int
	// MaxDepth specifies the maximum depth below which values will be
	// automatically be considered equal. A value of zero or less indicates an
	// infinite maxmimum depth.
	MaxDepth int
	// MaxDiffs specifies the maximum number of diffs to be returned. A value of
	// zero or less indicates an infinite maximum amount.
	MaxDiffs int
	// NilMapsAreEmpty, when true, causes nil maps to be equal to maps with zero
	// elements.
	NilMapsAreEmpty bool
	// NilSlicesAreEmpty, when true, causes nil slices to be equal to slices
	// with zero elements.
	NilSlicesAreEmpty bool
}

// NewComparer returns a new Comparer with a sensible default configuration.
//
// For the purposes of testing, NewComparer should be avoided. A Comparer with
// the desired options should be created manually.
func NewComparer() *Comparer {
	return &Comparer{
		CompareUnexportedFields: false,
		FloatPrecision:          34, // Close to 1e-10.
		MaxDepth:                0,
		MaxDiffs:                10,
		NilMapsAreEmpty:         false,
		NilSlicesAreEmpty:       false,
	}
}

// Diff represents a single unit of difference between two values.
type Diff struct {
	left  string
	right string
	stack string
}

// String returns a string representation of the diff. The returned string is
// meant to be read by humans, so it is not guaranteed to be consistent.
func (d Diff) String() string {
	if d.stack != "" {
		return d.stack + ": " + d.left + " != " + d.right
	}
	return d.left + " != " + d.right
}

type compareState struct {
	Comparer
	result  []Diff
	stack   []string
	visited map[visit]struct{}
	floatx  *big.Float
	floaty  *big.Float
}

func (s *compareState) push(v, i string) {
	if len(s.stack) == 0 {
		s.stack = append(s.stack, v+i)
		return
	}
	s.stack = append(s.stack, i)
}

func (s *compareState) pop() {
	s.stack = s.stack[:len(s.stack)-1]
}

func (s *compareState) append(x, y interface{}) {
	if i, ok := x.(reflect.Value); ok && i.IsValid() {
		x = i.Interface()
	}
	if i, ok := y.(reflect.Value); ok && i.IsValid() {
		y = i.Interface()
	}
	s.result = append(s.result, Diff{
		left:  fmt.Sprintf("%v", x),
		right: fmt.Sprintf("%v", y),
		stack: strings.Join(s.stack, ""),
	})
}

// Equal makes a comparison between two values, and returns a list of
// differences between them. If the values are equivalent according to the
// current configuration, then nil is returned.
//
// Equal makes comparisons based on equivalency rather than equality. Values may
// or may not be equivalent depending on the tolerances configured by the
// Comparer. Generally, Equal is similar to reflect.Equal. Aside from the
// configurable options, there are several other differences:
//
//     - NaN is equivalent to NaN.
//     - Because of quirks with maps, maps containing NaN keys can be reported
//       incorrectly.
func (c Comparer) Equal(x, y interface{}) []Diff {
	if x == nil && y == nil {
		return nil
	}

	state := compareState{
		Comparer: c,
		visited:  make(map[visit]struct{}),
	}
	if state.FloatPrecision > 0 {
		state.floatx = new(big.Float).SetPrec(uint(state.FloatPrecision))
		state.floaty = new(big.Float).SetPrec(uint(state.FloatPrecision))
	}

	if x == nil && y != nil {
		state.append("<nil>", y)
		return state.result
	} else if x != nil && y == nil {
		state.append(x, "<nil>")
		return state.result
	}

	state.deepValueEqual(reflect.ValueOf(x), reflect.ValueOf(y), 0)
	if len(state.result) == 0 {
		return nil
	}
	return state.result
}

type visit struct {
	a1  unsafe.Pointer
	a2  unsafe.Pointer
	typ reflect.Type
}

func (s *compareState) deepValueEqual(x, y reflect.Value, depth int) bool {
	if s.MaxDepth > 0 && depth > s.MaxDepth {
		return true
	}

	if !x.IsValid() || !y.IsValid() {
		if !x.IsValid() && y.IsValid() {
			s.append("<nil>", y.Type())
		} else if x.IsValid() && !y.IsValid() {
			s.append(x.Type(), "<nil>")
		}
		return x.IsValid() == y.IsValid()
	}

	if x.Type() != y.Type() {
		s.append(x.Type(), y.Type())
		return false
	}

	hard := func(k reflect.Kind) bool {
		switch k {
		case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Interface:
			return true
		}
		return false
	}

	if x.CanAddr() && y.CanAddr() && hard(x.Kind()) {
		addr1 := unsafe.Pointer(x.UnsafeAddr())
		addr2 := unsafe.Pointer(y.UnsafeAddr())
		if uintptr(addr1) > uintptr(addr2) {
			addr1, addr2 = addr2, addr1
		}
		v := visit{addr1, addr2, x.Type()}
		if _, ok := s.visited[v]; ok {
			return true
		}
		s.visited[v] = struct{}{}
	}

	switch x.Kind() {
	case reflect.Array:
		for i := 0; i < x.Len(); i++ {
			s.push("array", fmt.Sprintf("[%d]", i))
			s.deepValueEqual(x.Index(i), y.Index(i), depth+1)
			s.pop()
			if len(s.result) >= s.MaxDiffs {
				return false
			}
		}
		return true
	case reflect.Slice:
		if s.NilSlicesAreEmpty {
			if x.IsNil() && y.Len() != 0 {
				s.append("<nil slice>", y)
				return false
			} else if x.Len() != 0 && y.IsNil() {
				s.append(x, "<nil slice>")
				return false
			}
		} else {
			if x.IsNil() && !y.IsNil() {
				s.append("<nil slice>", y)
				return false
			} else if !x.IsNil() && y.IsNil() {
				s.append(x, "<nil slice>")
				return false
			}
		}
		if x.Pointer() == y.Pointer() && x.Len() != y.Len() {
			return true
		}
		n := x.Len()
		if y.Len() > n {
			n = y.Len()
		}
		for i := 0; i < n; i++ {
			s.push("slice", fmt.Sprintf("[%d]", i))
			if i < x.Len() {
				if i < y.Len() {
					s.deepValueEqual(x.Index(i), y.Index(i), depth+1)
				} else {
					s.append(x.Index(i), "<no value>")
				}
			} else {
				s.append("<no value>", y.Index(i))
			}
			s.pop()
			if len(s.result) >= s.MaxDiffs {
				return false
			}
		}
		return true
	case reflect.Interface:
		if x.IsNil() || y.IsNil() {
			if x.IsNil() && !y.IsNil() {
				s.append(fmt.Sprintf("<nil %s>", x.Type()), y)
			} else if !x.IsNil() && y.IsNil() {
				s.append(x, fmt.Sprintf("<nil %s>", y.Type()))
			}
			return x.IsNil() == y.IsNil()
		}
		return s.deepValueEqual(x.Elem(), y.Elem(), depth)
	case reflect.Ptr:
		if x.Pointer() == y.Pointer() {
			return true
		}
		return s.deepValueEqual(x.Elem(), y.Elem(), depth)
	case reflect.Struct:
		for i, n := 0, x.NumField(); i < n; i++ {
			if !s.CompareUnexportedFields && x.Type().Field(i).PkgPath != "" {
				continue
			}
			s.push("struct", "."+x.Type().Field(i).Name)
			s.deepValueEqual(x.Field(i), y.Field(i), depth+1)
			s.pop()
			if len(s.result) >= s.MaxDiffs {
				return false
			}
		}
		return true
	case reflect.Map:
		if s.NilMapsAreEmpty {
			if x.IsNil() && y.Len() != 0 {
				s.append("<nil map>", y)
				return false
			} else if x.Len() != 0 && y.IsNil() {
				s.append(x, "<nil map>")
				return false
			}
		} else {
			if x.IsNil() && !y.IsNil() {
				s.append("<nil map>", y)
				return false
			} else if !x.IsNil() && y.IsNil() {
				s.append(x, "<nil map>")
				return false
			}
		}
		if x.Pointer() == y.Pointer() {
			return true
		}

		for _, k := range x.MapKeys() {
			s.push("map", fmt.Sprintf("[%v]", k))
			if y.MapIndex(k).IsValid() {
				s.deepValueEqual(x.MapIndex(k), y.MapIndex(k), depth+1)
			} else if x.MapIndex(k).IsValid() {
				s.append(x.MapIndex(k), "<no key>")
			} else {
				s.append("<invalid key>", "<no key>")
			}
			s.pop()
			if len(s.result) >= s.MaxDiffs {
				return false
			}
		}
		for _, k := range y.MapKeys() {
			if x.MapIndex(k).IsValid() {
				continue
			}
			s.push("map", fmt.Sprintf("[%v]", k))
			s.append("<no key>", y.MapIndex(k))
			s.pop()
			if len(s.result) >= s.MaxDiffs {
				return false
			}
		}

		return true
	case reflect.Func:
		if !x.IsNil() || !y.IsNil() {
			s.append(x, y)
			return false
		}
		return true
	case reflect.Bool:
		if x.Bool() != y.Bool() {
			s.append(x.Bool(), y.Bool())
			return false
		}
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if x.Int() != y.Int() {
			s.append(x.Int(), y.Int())
			return false
		}
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if x.Uint() != y.Uint() {
			s.append(x.Uint(), y.Uint())
			return false
		}
		return true
	case reflect.Float32, reflect.Float64:
		vx := x.Float()
		vy := y.Float()
		// Compare NaN. Both being NaN is considered equivalent.
		if vx != vx && vy != vy {
			return true
		} else if vx != vx || vy != vy {
			s.append(vx, vy)
			return false
		}
		if s.FloatPrecision <= 0 {
			if vx != vy {
				s.append(vx, vy)
				return false
			}
			return true
		}
		s.floatx.SetFloat64(vx)
		s.floaty.SetFloat64(vy)
		if s.floatx.Cmp(s.floaty) != 0 {
			s.append(vx, vy)
			return false
		}
		return true
	case reflect.Complex64, reflect.Complex128:
		vx := x.Complex()
		vy := y.Complex()
		if s.FloatPrecision <= 0 {
			if vx != vy {
				s.append(vx, vy)
				return false
			}
			return true
		}
		s.floatx.SetFloat64(real(vx))
		s.floaty.SetFloat64(real(vy))
		if s.floatx.Cmp(s.floaty) != 0 {
			s.append(vx, vy)
			return false
		}
		s.floatx.SetFloat64(imag(vx))
		s.floaty.SetFloat64(imag(vy))
		if s.floatx.Cmp(s.floaty) != 0 {
			s.append(vx, vy)
			return false
		}
		return true
	case reflect.String:
		if x.String() != y.String() {
			s.append(x.String(), y.String())
			return false
		}
		return true
	case reflect.Uintptr:
		fallthrough
	case reflect.Chan:
		fallthrough
	case reflect.UnsafePointer:
		fallthrough
	default:
		if x.Interface() != y.Interface() {
			s.append(x.Interface(), y.Interface())
			return false
		}
		return true
	}
}

// Config is the Comparer used by Equal.
var Config = NewComparer()

// Equal makes a comparison between two values, and returns a list of
// differences between them. If the values are equivalent according to the
// global configuration, then nil is returned.
func Equal(x, y interface{}) []Diff {
	return Config.Equal(x, y)
}
