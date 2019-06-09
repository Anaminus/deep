package deep

import (
	"math"
	"reflect"
	"testing"
)

type basic struct {
	X int
	Y float32
}

type notBasic basic

type unexported struct {
	E int
	u int
}

var (
	fn1 func()             // nil.
	fn2 func()             // nil.
	fn3 = func() { fn1() } // Not nil.
)

type fnType func()

type typeLoop *typeLoop
type interfaceLoop interface{}

var tLoop1, tLoop2 typeLoop
var iLoop1, iLoop2 interfaceLoop

func init() {
	tLoop1 = &tLoop2
	tLoop2 = &tLoop1
	iLoop1 = &iLoop2
	iLoop2 = &iLoop1
}

func newComparer(f ...interface{}) Comparer {
	c := &Comparer{
		CompareUnexportedFields: false,
		FloatPrecision:          34,
		MaxDepth:                0,
		MaxDiffs:                10,
		NilMapsAreEmpty:         false,
		NilSlicesAreEmpty:       false,
	}
	v := reflect.ValueOf(c).Elem()
	for i := 0; i < len(f); i += 2 {
		v.FieldByName(f[i].(string)).Set(reflect.ValueOf(f[i+1]))
	}
	return *c
}

var equalConfigs = [...]Comparer{
	0: newComparer(),
	1: newComparer("CompareUnexportedFields", true),
	2: newComparer("FloatPrecision", 21),
	3: newComparer("FloatPrecision", 4),
	4: newComparer("FloatPrecision", 1),
	5: newComparer("FloatPrecision", 0),
	6: newComparer("MaxDepth", 2),
	7: newComparer("MaxDiffs", 1),
	8: newComparer("NilMapsAreEmpty", true),
	9: newComparer("NilSlicesAreEmpty", true),
}

type r [len(equalConfigs)]int

type equalTest struct {
	eq   r
	x, y interface{}
}

// Compares first operand with itself.
type self struct{}

// Repeats first result.
const x = -1

var equalTests = []equalTest{
	/*            0  1  2  3  4  5  6  7  8  9 */
	/*#   0 */ {r{0, x, x, x, x, x, x, x, x, x}, nil, nil},
	/*#   1 */ {r{1, x, x, x, x, x, x, x, x, x}, 0, nil},
	/*#   2 */ {r{0, x, x, x, x, x, x, x, x, x}, 0, 0},

	/*#   3 */ {r{1, x, x, x, x, x, x, x, x, x}, false, nil},
	/*#   4 */ {r{0, x, x, x, x, x, x, x, x, x}, false, false},

	/*#   5 */ {r{1, x, x, x, x, x, x, x, x, x}, "", nil},
	/*#   6 */ {r{0, x, x, x, x, x, x, x, x, x}, "", ""},

	/*#   7 */ {r{1, x, x, x, x, x, x, x, x, x}, 1, 0},
	/*#   8 */ {r{0, x, x, x, x, x, x, x, x, x}, 1, 1},

	/*#   9 */ {r{0, x, x, x, x, x, x, x, x, x}, int32(0), int32(0)},
	/*#  10 */ {r{1, x, x, x, x, x, x, x, x, x}, int32(1), int32(0)},

	/*#  11 */ {r{0, x, x, x, x, x, x, x, x, x}, uint(0), uint(0)},
	/*#  12 */ {r{1, x, x, x, x, x, x, x, x, x}, uint(1), uint(0)},

	/*#  13 */ {r{0, x, x, x, x, x, x, x, x, x}, float64(0.5), float64(0.5)},
	/*#  14 */ {r{1, x, x, x, 0, x, x, x, x, x}, float64(0.6), float64(0.5)},

	/*#  15 */ {r{0, x, x, x, x, x, x, x, x, x}, float32(0.5), float32(0.5)},
	/*#  16 */ {r{1, x, x, x, 0, x, x, x, x, x}, float32(0.6), float32(0.5)},

	/*#  17 */ {r{0, x, x, x, x, x, x, x, x, x}, "foo", "foo"},
	/*#  18 */ {r{1, x, x, x, x, x, x, x, x, x}, "foo", "bar"},
	/*#  19 */ {r{1, x, x, x, x, x, x, x, x, x}, "foobar", "bar"},

	/*#  20 */ {r{1, x, x, x, x, x, x, x, x, x}, float64(0.1), float64(0.2)},
	/*#  21 */ {r{1, x, x, x, 0, x, x, x, x, x}, float64(0.11), float64(0.12)},
	/*#  22 */ {r{1, x, x, x, 0, x, x, x, x, x}, float64(0.121), float64(0.122)},
	/*#  23 */ {r{1, x, x, 0, 0, x, x, x, x, x}, float64(0.1231), float64(0.1232)},
	/*#  24 */ {r{1, x, x, 0, 0, x, x, x, x, x}, float64(0.12341), float64(0.12342)},
	/*#  25 */ {r{1, x, x, 0, 0, x, x, x, x, x}, float64(0.123451), float64(0.123452)},
	/*#  26 */ {r{1, x, x, 0, 0, x, x, x, x, x}, float64(0.1234561), float64(0.1234562)},
	/*#  27 */ {r{1, x, 0, 0, 0, x, x, x, x, x}, float64(0.12345671), float64(0.12345672)},
	/*#  28 */ {r{1, x, 0, 0, 0, x, x, x, x, x}, float64(0.123456781), float64(0.123456782)},
	/*#  29 */ {r{1, x, 0, 0, 0, x, x, x, x, x}, float64(0.1234567891), float64(0.1234567892)},
	/*#  30 */ {r{1, x, 0, 0, 0, x, x, x, x, x}, float64(0.12345678901), float64(0.12345678902)},

	/*#  31 */ {r{1, x, x, x, x, x, x, x, x, x}, float32(0.1), float32(0.2)},
	/*#  32 */ {r{1, x, x, x, 0, x, x, x, x, x}, float32(0.11), float32(0.12)},
	/*#  33 */ {r{1, x, x, x, 0, x, x, x, x, x}, float32(0.121), float32(0.122)},
	/*#  34 */ {r{1, x, x, 0, 0, x, x, x, x, x}, float32(0.1231), float32(0.1232)},
	/*#  35 */ {r{1, x, x, 0, 0, x, x, x, x, x}, float32(0.12341), float32(0.12342)},
	/*#  36 */ {r{1, x, x, 0, 0, x, x, x, x, x}, float32(0.123451), float32(0.123452)},
	/*#  37 */ {r{1, x, x, 0, 0, x, x, x, x, x}, float32(0.1234561), float32(0.1234562)},
	/*#  38 */ {r{1, x, 0, 0, 0, x, x, x, x, x}, float32(0.12345671), float32(0.12345672)},
	/*#  39 */ {r{0, x, x, x, x, x, x, x, x, x}, float32(0.123456781), float32(0.123456782)},
	/*#  40 */ {r{0, x, x, x, x, x, x, x, x, x}, float32(0.1234567891), float32(0.1234567892)},
	/*#  41 */ {r{0, x, x, x, x, x, x, x, x, x}, float32(0.12345678901), float32(0.12345678902)},

	/*#  42 */ {r{0, x, x, x, x, x, x, x, x, x}, [0]int{}, [0]int{}},
	/*#  43 */ {r{1, x, x, x, x, x, x, x, x, x}, [0]int{}, [3]int{}},
	/*#  44 */ {r{0, x, x, x, x, x, x, x, x, x}, [3]int{}, [3]int{}},
	/*#  45 */ {r{3, x, x, x, x, x, x, 1, x, x}, [3]int{1, 2, 3}, [3]int{}},
	/*#  46 */ {r{0, x, x, x, x, x, x, x, x, x}, [3]int{1, 2, 3}, [3]int{1, 2, 3}},
	/*#  47 */ {r{1, x, x, x, x, x, x, x, x, x}, [3]int{1, 2, 3}, [3]int{1, 2, 4}},
	/*#  48 */ {r{0, x, x, x, x, x, x, x, x, x}, &[3]int{1, 2, 3}, &[3]int{1, 2, 3}},
	/*#  49 */ {r{1, x, x, x, x, x, x, x, x, x}, &[3]int{1, 2, 3}, &[3]int{1, 2, 4}},
	/*#  50 */ {r{0, x, x, x, x, x, x, x, x, x}, &[3]int{1, 2, 3}, self{}},

	/*#  51 */ {r{0, x, x, x, x, x, x, x, x, x}, make([]int, 3), make([]int, 3)},
	/*#  52 */ {r{1, x, x, x, x, x, x, x, x, x}, make([]int, 3), make([]int, 4)},
	/*#  53 */ {r{0, x, x, x, x, x, x, x, x, x}, make([]int, 3), self{}},

	/*#  54 */ {r{0, x, x, x, x, x, x, x, x, x}, basic{1, 0.5}, basic{1, 0.5}},
	/*#  55 */ {r{1, x, x, x, 0, x, x, x, x, x}, basic{1, 0.5}, basic{1, 0.6}},
	/*#  56 */ {r{1, x, x, x, x, x, x, x, x, x}, basic{1, 0}, basic{2, 0}},
	/*#  57 */ {r{1, x, x, x, x, x, x, x, x, x}, basic{1, 0.5}, notBasic{1, 0.5}},
	/*#  58 */ {r{0, x, x, x, x, x, x, x, x, x}, notBasic{1, 0.5}, notBasic{1, 0.5}},

	/*#  59 */ {r{0, x, x, x, x, x, x, x, x, x}, unexported{E: 1, u: 1}, unexported{E: 1, u: 1}},
	/*#  60 */ {r{1, x, x, x, x, x, x, x, x, x}, unexported{E: 1, u: 1}, unexported{E: 2, u: 1}},
	/*#  61 */ {r{0, 1, x, x, x, x, x, x, x, x}, unexported{E: 1, u: 1}, unexported{E: 1, u: 2}},
	/*#  62 */ {r{1, 2, x, x, x, x, x, x, x, x}, unexported{E: 1, u: 1}, unexported{E: 2, u: 2}},

	/*#  63 */ {r{0, x, x, x, x, x, x, x, x, x}, &unexported{E: 1, u: 1}, self{}},
	/*#  64 */ {r{0, x, x, x, x, x, x, x, x, x}, &unexported{E: 2, u: 1}, self{}},
	/*#  65 */ {r{0, x, x, x, x, x, x, x, x, x}, &unexported{E: 1, u: 2}, self{}},

	/*#  66 */ {r{0, x, x, x, x, x, x, x, x, x}, error(nil), error(nil)},

	/*#  67 */ {r{0, x, x, x, x, x, x, x, x, x}, map[int]string{1: "one", 2: "two"}, self{}},
	/*#  68 */ {r{0, x, x, x, x, x, x, x, x, x}, map[int]string{1: "one", 2: "two"}, map[int]string{2: "two", 1: "one"}},
	/*#  69 */ {r{2, x, x, x, x, x, x, 1, x, x}, map[int]string{1: "one", 3: "two"}, map[int]string{2: "two", 1: "one"}},
	/*#  70 */ {r{1, x, x, x, x, x, x, x, x, x}, map[int]string{1: "one", 2: "txo"}, map[int]string{2: "two", 1: "one"}},
	/*#  71 */ {r{1, x, x, x, x, x, x, x, x, x}, map[int]string{1: "one"}, map[int]string{2: "two", 1: "one"}},
	/*#  72 */ {r{1, x, x, x, x, x, x, x, x, x}, map[int]string{2: "two", 1: "one"}, map[int]string{1: "one"}},

	/*#  73 */ {r{0, x, x, x, x, x, x, x, x, x}, fn1, fn1},
	/*#  74 */ {r{0, x, x, x, x, x, x, x, x, x}, fn1, fn2},
	/*#  75 */ {r{1, x, x, x, x, x, x, x, x, x}, fn1, fn3},
	/*#  76 */ {r{0, x, x, x, x, x, x, x, x, x}, fn2, fn2},
	/*#  77 */ {r{1, x, x, x, x, x, x, x, x, x}, fn2, fn3},
	/*#  78 */ {r{1, x, x, x, x, x, x, x, x, x}, fn3, fn3},

	/*#  79 */ {r{0, x, x, x, x, x, x, x, x, x}, fnType(nil), fnType(nil)},
	/*#  80 */ {r{1, x, x, x, x, x, x, x, x, x}, fnType(nil), fnType(func() {})},
	/*#  81 */ {r{1, x, x, x, x, x, x, x, x, x}, fnType(func() {}), fnType(func() {})},

	/*#  82 */ {r{0, x, x, x, x, x, x, x, x, x}, [][]int{{1}}, [][]int{{1}}},
	/*#  83 */ {r{1, x, x, x, x, x, x, x, x, x}, [][]int{{1}}, [][]int{{2}}},
	/*#  84 */ {r{0, x, x, x, x, x, x, x, x, x}, [][]int{{1}}, self{}},
	/*#  85 */ {r{0, x, x, x, x, x, x, x, x, x}, [][][]int{{{1}}}, [][][]int{{{1}}}},
	/*#  86 */ {r{1, x, x, x, x, x, 0, x, x, x}, [][][]int{{{1}}}, [][][]int{{{2}}}},
	/*#  87 */ {r{0, x, x, x, x, x, x, x, x, x}, [][][]int{{{1}}}, self{}},

	/*#  88 */ {r{0, x, x, x, x, x, x, x, x, x}, math.NaN(), math.NaN()},
	/*#  89 */ {r{1, x, x, x, x, x, x, x, x, x}, math.NaN(), 0.5},
	/*#  90 */ {r{0, x, x, x, x, x, x, x, x, x}, float32(math.NaN()), float32(math.NaN())},
	/*#  91 */ {r{1, x, x, x, x, x, x, x, x, x}, float32(math.NaN()), 0.5},
	/*#  92 */ {r{0, x, x, x, x, x, x, x, x, x}, &[1]float64{math.NaN()}, &[1]float64{math.NaN()}},
	/*#  93 */ {r{1, x, x, x, x, x, x, x, x, x}, &[1]float64{math.NaN()}, &[1]float64{0.5}},
	/*#  94 */ {r{0, x, x, x, x, x, x, x, x, x}, &[1]float64{math.NaN()}, self{}},
	/*#  95 */ {r{0, x, x, x, x, x, x, x, x, x}, []float64{math.NaN()}, []float64{math.NaN()}},
	/*#  96 */ {r{0, x, x, x, x, x, x, x, x, x}, []float64{math.NaN()}, self{}},
	/*#  97 */ {r{2, x, x, x, x, x, x, 1, x, x}, map[float64]float64{math.NaN(): 1}, map[float64]float64{1: 2}},
	/*#  98 */ {r{0, x, x, x, x, x, x, x, x, x}, map[float64]float64{math.NaN(): 1}, self{}},

	/*#  99 */ {r{0, x, x, x, x, x, x, x, x, x}, []int(nil), []int(nil)},
	/*# 100 */ {r{1, x, x, x, x, x, x, x, x, 0}, []int(nil), []int{}},
	/*# 101 */ {r{1, x, x, x, x, x, x, x, x, x}, []int(nil), [0]int{}},
	/*# 102 */ {r{1, x, x, x, x, x, x, x, x, x}, []int(nil), []int{1}},
	/*# 103 */ {r{0, x, x, x, x, x, x, x, x, x}, []int(nil), self{}},
	/*# 104 */ {r{0, x, x, x, x, x, x, x, x, x}, []int{}, []int{}},
	/*# 105 */ {r{1, x, x, x, x, x, x, x, x, x}, []int{}, [0]int{}},
	/*# 106 */ {r{1, x, x, x, x, x, x, x, x, x}, []int{}, []int{1}},
	/*# 107 */ {r{0, x, x, x, x, x, x, x, x, x}, []int{}, self{}},
	/*# 108 */ {r{1, x, x, x, x, x, x, x, x, x}, []int{1}, [0]int{}},
	/*# 109 */ {r{0, x, x, x, x, x, x, x, x, x}, []int{1}, []int{1}},
	/*# 110 */ {r{0, x, x, x, x, x, x, x, x, x}, []int{1}, self{}},

	/*# 111 */ {r{0, x, x, x, x, x, x, x, x, x}, map[int]int(nil), map[int]int(nil)},
	/*# 112 */ {r{1, x, x, x, x, x, x, x, 0, x}, map[int]int(nil), map[int]int{}},
	/*# 113 */ {r{1, x, x, x, x, x, x, x, x, x}, map[int]int(nil), map[int]int{1: 1}},
	/*# 114 */ {r{0, x, x, x, x, x, x, x, x, x}, map[int]int(nil), self{}},
	/*# 115 */ {r{0, x, x, x, x, x, x, x, x, x}, map[int]int{}, map[int]int{}},
	/*# 116 */ {r{1, x, x, x, x, x, x, x, x, x}, map[int]int{}, map[int]int{1: 1}},
	/*# 117 */ {r{0, x, x, x, x, x, x, x, x, x}, map[int]int{}, self{}},
	/*# 118 */ {r{0, x, x, x, x, x, x, x, x, x}, map[int]int{1: 1}, map[int]int{1: 1}},
	/*# 119 */ {r{0, x, x, x, x, x, x, x, x, x}, map[int]int{1: 1}, self{}},

	/*# 120 */ {r{0, x, x, x, x, x, x, x, x, x}, &[3]interface{}{1, 2, 3}, &[3]interface{}{1, 2, 3}},
	/*# 121 */ {r{0, x, x, x, x, x, x, x, x, x}, &[3]interface{}{true, 2, ""}, &[3]interface{}{true, 2, ""}},
	/*# 122 */ {r{1, x, x, x, x, x, x, x, x, x}, &[3]interface{}{true, 2, ""}, &[3]interface{}{true, 2, "s"}},
	/*# 123 */ {r{2, x, x, x, x, x, x, 1, x, x}, &[3]interface{}{true, 2, ""}, &[3]interface{}{1, 2, 3}},
	/*# 124 */ {r{3, x, x, x, x, x, x, 1, x, x}, &[3]interface{}{true, 1, ""}, &[3]interface{}{1, 2, 3}},

	/*# 125 */ {r{0, x, x, x, x, x, x, x, x, x}, &tLoop1, &tLoop1},
	/*# 126 */ {r{0, x, x, x, x, x, x, x, x, x}, &tLoop1, &tLoop2},
	/*# 127 */ {r{0, x, x, x, x, x, x, x, x, x}, &iLoop1, &iLoop1},
	/*# 128 */ {r{0, x, x, x, x, x, x, x, x, x}, &iLoop1, &iLoop2},

	/*# 129 */ {r{1, x, x, x, x, x, x, x, x, x}, 1, 1.0},
	/*# 130 */ {r{1, x, x, x, x, x, x, x, x, x}, int32(1), int64(1)},
	/*# 131 */ {r{1, x, x, x, x, x, x, x, x, x}, 0.5, "foo"},
	/*# 132 */ {r{1, x, x, x, x, x, x, x, x, x}, []int{1, 2, 3}, [3]int{1, 2, 3}},
	/*# 133 */ {r{1, x, x, x, x, x, x, x, x, x}, map[uint]string{1: "one", 2: "two"}, map[int]string{2: "two", 1: "one"}},
}

func TestEqual(t *testing.T) {
	for i, c := range equalConfigs {
		for j, test := range equalTests {
			if test.y == (self{}) {
				test.y = test.x
			}
			eq := test.eq[i]
			if eq < 0 {
				eq = test.eq[0]
				if eq < 0 {
					continue
				}
			}
			if r := c.Equal(test.x, test.y); len(r) != eq {
				t.Errorf("[%d][%d]: want %d, got %d from (%v, %v): %v", i, j, eq, len(r), test.x, test.y, r)
			} else if eq == 0 && r != nil {
				t.Errorf("[%d][%d]: want nil, got [] from (%v, %v): %v", i, j, test.x, test.y, r)
			}
			if r := c.Equal(test.y, test.x); len(r) != eq {
				t.Errorf("[%d][%d]: want %d, got %d from (%v, %v): %v", i, j, eq, len(r), test.y, test.x, r)
			} else if eq == 0 && r != nil {
				t.Errorf("[%d][%d]: want nil, got [] from (%v, %v): %v", i, j, test.y, test.x, r)
			}
		}
	}
}
