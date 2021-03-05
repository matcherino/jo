package jo

import (
	"encoding/json"
	"testing"

	"github.com/matcherino/jo/dummy"
	"github.com/stretchr/testify/assert"
)

var jsonStrings = map[string]string{
	"array": `["foo", {"fruit": "apple"}]`,

	"s1": `{
		"a": 1,
		"b": "moo",
		"c": true,
		"d": 1.2,
		"access_token": "123"
	}`,

	"s2": `{
		"a": {
			"b": "moo",
			"c": 1,
			"d": false
		},
		"b": 0,
		"c": [1,2,3],
		"d": [[0, 1], {"a": 1}, [{"b": 2}, {"c": 3}]],
		"users": [{"name": "foo"}, {"name": "bash"}]
	}`,

	"complex": `{
	  "a": {
	    "b": {
	      "c": {
			"an-array": [1, 2, 3],
			"foo-bar": "foobar",
	        "h": [
	          [1, 2, 3],
	          ["a", "b", "c"],
	          [1.2, 4.5, 7.8],
	          [
	            ["h", "i", "j"],
	            ["k", "l", "m"]
	          ]
	        ],
	        "e": "moo",
	        "d": 1,
	        "g": {
	          "y": [1.3, 1.5, 2.8],
	          "x": [0, 1, 2],
	          "z": [
	            {"a": "hello", "b": "world"},
	            {"a": 100.12, "b": 200.24},
	            {"a": 1, "c": "go rocks", "b": 2}
	          ]
	        },
	        "f": ["cow", "dog", "bird"]
	      }
	    }
	  }
	}`,
}

func getTestJSON(t *testing.T, name string) *Object {
	n, err := NewFromBytes([]byte(jsonStrings[name]))
	assert.NoError(t, err, "JSON Decode failed: %s", name)
	return n
}

func TestSplitPath(t *testing.T) {
	var testPaths = []struct {
		path  string
		parts []interface{}
	}{
		{"a", []interface{}{"a"}},
		{"a.b", []interface{}{"a", "b"}},
		{"[0]", []interface{}{0}},
		{"[0][1][2]", []interface{}{0, 1, 2}},
		{"a.b.c[0][1].d[0]", []interface{}{"a", "b", "c", 0, 1, "d", 0}},
		{"[0][1].a", []interface{}{0, 1, "a"}},
		{"[0].a[1].b[2][3].c.a", []interface{}{0, "a", 1, "b", 2, 3, "c", "a"}},
	}

	for _, item := range testPaths {
		parts, err := splitPath(item.path)
		assert.Nil(t, err)
		assert.Equal(t, parts, item.parts)
	}
}

func TestSplitPathErrors(t *testing.T) {
	var errorPaths = []string{
		"",
		"a..b.",
		"..",
		"a[[2]",
		"[]",
		"a[0.",
		"a[0].[1]",
	}

	for _, item := range errorPaths {
		_, err := splitPath(item)
		assert.Error(t, err)
	}
}

func TestGetSimple(t *testing.T) {
	json := getTestJSON(t, "s1")
	testPaths := []struct {
		path string
		val  interface{}
	}{
		{"a", float64(1)},
		{"b", "moo"},
		{"c", true},
		{"d", float64(1.2)},
		{"access_token", "123"},
	}

	for _, i := range testPaths {
		v, err := json.Get(i.path)
		assert.Nil(t, err)
		assert.Equal(t, v, i.val)
	}
}

func TestArrayBody(t *testing.T) {
	json := getTestJSON(t, "array")
	testPaths := []struct {
		path string
		val  interface{}
	}{
		{"[0]", "foo"},
		{"[1].fruit", "apple"},
	}

	for _, i := range testPaths {
		v, err := json.Get(i.path)
		assert.Nil(t, err)
		assert.Equal(t, v, i.val, i.path)
	}
}

func TestGetComplex(t *testing.T) {
	json := getTestJSON(t, "complex")
	testPaths := []struct {
		path string
		val  interface{}
	}{
		{"a.b.c.d", float64(1)},
		{"a.b.c.e", "moo"},
		{"a.b.c.f", []interface{}{"cow", "dog", "bird"}},
		{"a.b.c.an-array[0]", float64(1)},
		{"a.b.c.foo-bar", "foobar"},
		{"a.b.c.g.x[0]", float64(0)},
		{"a.b.c.g.y[1]", float64(1.5)},
		{"a.b.c.g.z[0].a", "hello"},
		{"a.b.c.g.z[1].b", float64(200.24)},
		{"a.b.c.g.z[2].c", "go rocks"},
		{"a.b.c.h[0][0]", float64(1)},
		{"a.b.c.h[0][1]", float64(2)},
		{"a.b.c.h[0][2]", float64(3)},
		{"a.b.c.h[3][0][0]", "h"},
		{"a.b.c.h[3][1][2]", "m"},
	}

	for _, i := range testPaths {
		v, err := json.Get(i.path)
		assert.Nil(t, err)
		assert.Equal(t, v, i.val, i.path)
	}
}

func TestGetErrors(t *testing.T) {
	json := getTestJSON(t, "s2")
	testPaths := []struct {
		path      string
		errString string
	}{
		{"a.b.e", "moo is not an object"},
		{"a.f.m.a", "Key does not exist"},
		{"a[0]", "not an array"},
		{"c[10]", "out of bounds"},
		{"d[0][5]", "out of bounds"},
		{"d[1].b", "does not exist"},
		{"d[2][0].b.e", "not an object"},
		{"d[2][0].c", "does not exist"},
	}

	for _, item := range testPaths {
		_, err := json.Get(item.path)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), item.errString, item.path)
		}
	}
}

func TestSetNew(t *testing.T) {
	json := New()
	tests := []struct {
		path string
		val  interface{}
	}{
		{"a.b.c", 1},
		{"a.b.d", "moo"},
		{"b", []interface{}{1, 2, 3}},
		{"b[0]", 4},
		{"c", map[string]interface{}{
			"A": 1, "B": 1.2, "C": true,
		}},
		{"c.A", false},
		{"c.A", "X"},
		{"c.B", 4.5},
		{"b[0]", []interface{}{1.2, 1.3, 1.4}},
		{"b[0][0]", []interface{}{"a", "b", "c"}},
		{"b[0][0][1]", "FUU"},
	}

	for _, i := range tests {
		json.Set(i.path, i.val)
		v, err := json.Get(i.path)
		assert.NoError(t, err)
		assert.Equal(t, i.val, v, "%s != %s", i.path, i.val)
	}

	jsonString := json.Stringify()
	assert.Equal(t, jsonString,
		`{"a":{"b":{"c":1,"d":"moo"}},"b":[[["a","FUU","c"],`+
			`1.3,1.4],2,3],"c":{"A":"X","B":4.5,"C":true}}`)
}

func TestSetExisting(t *testing.T) {
	json := getTestJSON(t, "s2")
	tests := []struct {
		path string
		val  interface{}
	}{
		{"a.b", map[string]interface{}{
			"x": float64(0.5), "y": float64(10),
		}},
		{"c[0]", "xxx"},
		{"b", []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)}},
		{"d[1].a", "zzz"},
	}

	for _, i := range tests {
		json.Set(i.path, i.val)
		v, err := json.Get(i.path)
		assert.NoError(t, err)
		assert.Equal(t, i.val, v, "%s != %s", i.path, i.val)
	}

	jsonString := json.Stringify()
	assert.Equal(t,
		`{"a":{"b":{"x":0.5,"y":10},"c":1,"d":false},`+
			`"b":[1,2,3,4,5],"c":["xxx",2,3],"d":[[0,1],`+
			`{"a":"zzz"},[{"b":2},{"c":3}]],`+
			`"users":[{"name":"foo"},{"name":"bash"}]}`, jsonString)
}

func TestScalar(t *testing.T) {
	s := []byte(`"hello"`)
	n := []byte(`1`)
	var o Object

	err := json.Unmarshal(s, &o)
	assert.NoError(t, err)
	assert.Equal(t, "hello", o.AsString("."))

	err = json.Unmarshal(n, &o)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), o.AsFloat("."))
}

func TestMarshal(t *testing.T) {
	m := New()
	m.Set("foo.bar", "hello")
	m.Set("foo.nums", []int{1, 2})

	b, err := json.Marshal(m)
	assert.NoError(t, err)
	assert.Equal(t, `{"foo":{"bar":"hello","nums":[1,2]}}`, string(b))
}

func TestGetObjects(t *testing.T) {
	o := getTestJSON(t, "s2")
	objs := o.AsObjects("users")
	assert.Equal(t, "foo", objs[0].AsString("name"))
}

func TestDelete(t *testing.T) {
	o := getTestJSON(t, "complex")
	s := o.AsString("a.b.c.e")
	assert.Equal(t, "moo", s)

	o.Delete("a.b.c.e")
	_, err := o.Get("a.b.c.e")
	assert.EqualError(t, ErrKeyDoesNotExist, err.Error())

	// Arrays are problematic as the recursive delete needs to know its grandparent
	// to replace the slice.
	//
	// fmt.Println("PRE", o.MustArray("a.b.c.an-array"))
	// err := o.Delete("a.b.c.an-array[0]")
	// assert.NoError(t, err)
	// arr := o.MustArray("a.b.c.an-array")
	// fmt.Println("POST", o.MustArray("a.b.c.an-array"))
	// assert.Equal(t, 2, len(arr))
}

func TestConversion(t *testing.T) {
	d := dummy.Object{
		"foo": "bar",
		"nested": dummy.Object{
			"bah": "baz",
		},
		"fruits": []string{"apple"},
		"users": dummy.Array{
			dummy.Object{"name": "mario"},
		},
	}
	o, err := NewFromAny(d)
	assert.NoError(t, err)
	assert.Equal(t, "bar", o.AsString("foo"))
	assert.Equal(t, "baz", o.AsString("nested.bah"))
	assert.Equal(t, "apple", o.AsString("fruits[0]"))
	assert.Equal(t, "mario", o.AsString("users[0].name"))
}

func TestMap(t *testing.T) {
	d := map[string]interface{}{
		"foo": "bar",
		"nested": dummy.Object{
			"bah": "baz",
		},
		"fruits": []string{"apple"},
	}
	o := NewFromMap(d)
	for k, v := range o.AsMap("nested") {
		assert.Equal(t, "bah", k)
		assert.Equal(t, "baz", v)
	}
}
