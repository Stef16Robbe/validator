// Package validator implements value validations
//
// Copyright 2014 Roberto Teixeira <robteix@robteix.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator_test

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
	"gopkg.in/validator.v2"
)

func Test(t *testing.T) {
	TestingT(t)
}

type MySuite struct{}

var _ = Suite(&MySuite{})

type Simple struct {
	A int `validate:"min=10"`
}

type I interface {
	Foo() string
}

type Impl struct {
	F string `validate:"len=3"`
}

func (i *Impl) Foo() string {
	return i.F
}

type Impl2 struct {
	F string `validate:"len=3"`
}

func (i Impl2) Foo() string {
	return i.F
}

type NestedStruct struct {
	A string `validate:"nonzero" json:"a"`
}

type TestStruct struct {
	A   int    `validate:"nonzero" json:"a"`
	B   string `validate:"len=8,min=6,max=4"`
	Sub struct {
		A int `validate:"nonzero" json:"sub_a"`
		B string
		C float64 `validate:"nonzero,min=1" json:"c_is_a_float"`
		D *string `validate:"nonzero"`
	}
	D *Simple `validate:"nonzero"`
	E I       `validate:"nonzero"`
}

type TestCompositedStruct struct {
	NestedStruct `json:""`
	OtherNested  NestedStruct   `json:"otherNested"`
	Items        []NestedStruct `json:"nestedItems"`
}

func (ms *MySuite) TestValidate(c *C) {
	t := TestStruct{
		A: 0,
		B: "12345",
	}
	t.Sub.A = 1
	t.Sub.B = ""
	t.Sub.C = 0.0
	t.D = &Simple{10}
	t.E = &Impl{"hello"}

	err := validator.Validate(t)
	c.Assert(err, NotNil)

	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], HasError, validator.ErrZeroValue)
	c.Assert(errs["B"], HasError, validator.ErrLen)
	c.Assert(errs["B"], HasError, validator.ErrMin)
	c.Assert(errs["B"], HasError, validator.ErrMax)
	c.Assert(errs["Sub.A"], HasLen, 0)
	c.Assert(errs["Sub.B"], HasLen, 0)
	c.Assert(errs["Sub.C"], HasLen, 2)
	c.Assert(errs["Sub.D"], HasError, validator.ErrZeroValue)
	c.Assert(errs["E.F"], HasError, validator.ErrLen)
}

func (ms *MySuite) TestValidSlice(c *C) {
	s := make([]int, 0, 10)
	err := validator.Valid(s, "nonzero")
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, validator.ErrZeroValue)

	for i := 0; i < 10; i++ {
		s = append(s, i)
	}

	err = validator.Valid(s, "min=11,max=5,len=9,nonzero")
	c.Assert(err, NotNil)
	errs, ok = err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, validator.ErrMin)
	c.Assert(errs, HasError, validator.ErrMax)
	c.Assert(errs, HasError, validator.ErrLen)
	c.Assert(errs, Not(HasError), validator.ErrZeroValue)
}

func (ms *MySuite) TestValidMap(c *C) {
	m := make(map[string]string)
	err := validator.Valid(m, "nonzero")
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, validator.ErrZeroValue)

	err = validator.Valid(m, "min=1")
	c.Assert(err, NotNil)
	errs, ok = err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, validator.ErrMin)

	m = map[string]string{"A": "a", "B": "a"}
	err = validator.Valid(m, "max=1")
	c.Assert(err, NotNil)
	errs, ok = err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, validator.ErrMax)

	err = validator.Valid(m, "min=2, max=5")
	c.Assert(err, IsNil)

	m = map[string]string{
		"1": "a",
		"2": "b",
		"3": "c",
		"4": "d",
		"5": "e",
	}
	err = validator.Valid(m, "len=4,min=6,max=1,nonzero")
	c.Assert(err, NotNil)
	errs, ok = err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, validator.ErrLen)
	c.Assert(errs, HasError, validator.ErrMin)
	c.Assert(errs, HasError, validator.ErrMax)
	c.Assert(errs, Not(HasError), validator.ErrZeroValue)

}

func (ms *MySuite) TestValidFloat(c *C) {
	err := validator.Valid(12.34, "nonzero")
	c.Assert(err, IsNil)

	err = validator.Valid(0.0, "nonzero")
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, validator.ErrZeroValue)
}

func (ms *MySuite) TestValidInt(c *C) {
	i := 123
	err := validator.Valid(i, "nonzero")
	c.Assert(err, IsNil)

	err = validator.Valid(i, "min=1")
	c.Assert(err, IsNil)

	err = validator.Valid(i, "min=124, max=122")
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, validator.ErrMin)
	c.Assert(errs, HasError, validator.ErrMax)

	err = validator.Valid(i, "max=10")
	c.Assert(err, NotNil)
	errs, ok = err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, validator.ErrMax)
}

func (ms *MySuite) TestValidString(c *C) {
	s := "test1234"
	err := validator.Valid(s, "len=8")
	c.Assert(err, IsNil)

	err = validator.Valid(s, "len=0")
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, validator.ErrLen)

	err = validator.Valid(s, "regexp=^[tes]{4}.*")
	c.Assert(err, IsNil)

	err = validator.Valid(s, "regexp=^.*[0-9]{5}$")
	c.Assert(err, NotNil)

	err = validator.Valid("", "nonzero,len=3,max=1")
	c.Assert(err, NotNil)
	errs, ok = err.(validator.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 2)
	c.Assert(errs, HasError, validator.ErrZeroValue)
	c.Assert(errs, HasError, validator.ErrLen)
	c.Assert(errs, Not(HasError), validator.ErrMax)
}

func (ms *MySuite) TestValidateStructVar(c *C) {
	// just verifies that a the given val is a struct
	validator.SetValidationFunc("struct", func(val interface{}, _ string) error {
		v := reflect.ValueOf(val)
		if v.Kind() == reflect.Struct {
			return nil
		}
		return validator.ErrUnsupported
	})

	type test struct {
		A int
	}
	err := validator.Valid(test{}, "struct")
	c.Assert(err, IsNil)

	type test2 struct {
		B int
	}
	type test1 struct {
		A test2 `validate:"struct"`
	}

	err = validator.Validate(test1{})
	c.Assert(err, IsNil)

	type test4 struct {
		B int `validate:"foo"`
	}
	type test3 struct {
		A test4
	}
	err = validator.Validate(test3{})
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A.B"], HasError, validator.ErrUnknownTag)
}

func (ms *MySuite) TestValidatePointerVar(c *C) {
	// just verifies that a the given val is a struct
	validator.SetValidationFunc("struct", func(val interface{}, _ string) error {
		v := reflect.ValueOf(val)
		if v.Kind() == reflect.Struct {
			return nil
		}
		return validator.ErrUnsupported
	})
	validator.SetValidationFunc("nil", func(val interface{}, _ string) error {
		v := reflect.ValueOf(val)
		if v.IsNil() {
			return nil
		}
		return validator.ErrUnsupported
	})

	type test struct {
		A int
	}
	err := validator.Valid(&test{}, "struct")
	c.Assert(err, IsNil)

	type test2 struct {
		B int
	}
	type test1 struct {
		A *test2 `validate:"struct"`
	}

	err = validator.Validate(&test1{&test2{}})
	c.Assert(err, IsNil)

	type test4 struct {
		B int `validate:"foo"`
	}
	type test3 struct {
		A test4
	}
	err = validator.Validate(&test3{})
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A.B"], HasError, validator.ErrUnknownTag)

	err = validator.Valid((*test)(nil), "nil")
	c.Assert(err, IsNil)

	type test5 struct {
		A *test2 `validate:"nil"`
	}
	err = validator.Validate(&test5{})
	c.Assert(err, IsNil)

	type test6 struct {
		A *test2 `validate:"nonzero"`
	}
	err = validator.Validate(&test6{})
	errs, ok = err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], HasError, validator.ErrZeroValue)

	err = validator.Validate(&test6{&test2{}})
	c.Assert(err, IsNil)

	type test7 struct {
		A *string `validate:"min=6"`
		B *int    `validate:"len=7"`
		C *int    `validate:"min=12"`
		D *int    `validate:"nonzero"`
		E *int    `validate:"nonzero"`
		F *int    `validate:"nonnil"`
		G *int    `validate:"nonnil"`
	}
	s := "aaa"
	b := 8
	e := 0
	err = validator.Validate(&test7{&s, &b, nil, nil, &e, &e, nil})
	errs, ok = err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], HasError, validator.ErrMin)
	c.Assert(errs["B"], HasError, validator.ErrLen)
	c.Assert(errs["C"], IsNil)
	c.Assert(errs["D"], HasError, validator.ErrZeroValue)
	c.Assert(errs["E"], HasError, validator.ErrZeroValue)
	c.Assert(errs["F"], IsNil)
	c.Assert(errs["G"], HasError, validator.ErrZeroValue)
}

func (ms *MySuite) TestValidateOmittedStructVar(c *C) {
	type test2 struct {
		B int `validate:"min=1"`
	}
	type test1 struct {
		A test2 `validate:"-"`
	}

	t := test1{}
	err := validator.Validate(t)
	c.Assert(err, IsNil)

	errs := validator.Valid(test2{}, "-")
	c.Assert(errs, IsNil)
}

func (ms *MySuite) TestUnknownTag(c *C) {
	type test struct {
		A int `validate:"foo"`
	}
	t := test{}
	err := validator.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["A"], HasError, validator.ErrUnknownTag)
}

func (ms *MySuite) TestValidateStructWithSlice(c *C) {
	type test2 struct {
		Num    int    `validate:"max=2"`
		String string `validate:"nonzero"`
	}

	type test struct {
		Slices []test2 `validate:"len=1"`
	}

	t := test{
		Slices: []test2{{
			Num:    6,
			String: "foo",
		}},
	}
	err := validator.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["Slices[0].Num"], HasError, validator.ErrMax)
	c.Assert(errs["Slices[0].String"], IsNil) // sanity check
}

func (ms *MySuite) TestValidateStructWithNestedSlice(c *C) {
	type test2 struct {
		Num int `validate:"max=2"`
	}

	type test struct {
		Slices [][]test2
	}

	t := test{
		Slices: [][]test2{{{Num: 6}}},
	}
	err := validator.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["Slices[0][0].Num"], HasError, validator.ErrMax)
}

func (ms *MySuite) TestValidateStructWithMap(c *C) {
	type test2 struct {
		Num int `validate:"max=2"`
	}

	type test struct {
		Map          map[string]test2
		StructKeyMap map[test2]test2
	}

	t := test{
		Map: map[string]test2{
			"hello": {Num: 6},
		},
		StructKeyMap: map[test2]test2{
			{Num: 3}: {Num: 1},
		},
	}
	err := validator.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)

	c.Assert(errs["Map[hello](value).Num"], HasError, validator.ErrMax)
	c.Assert(errs["StructKeyMap[{Num:3}](key).Num"], HasError, validator.ErrMax)
}

func (ms *MySuite) TestUnsupported(c *C) {
	type test struct {
		A int     `validate:"regexp=a.*b"`
		B float64 `validate:"regexp=.*"`
	}
	t := test{}
	err := validator.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 2)
	c.Assert(errs["A"], HasError, validator.ErrUnsupported)
	c.Assert(errs["B"], HasError, validator.ErrUnsupported)
}

func (ms *MySuite) TestBadParameter(c *C) {
	type test struct {
		A string `validate:"min="`
		B string `validate:"len=="`
		C string `validate:"max=foo"`
	}
	t := test{}
	err := validator.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 3)
	c.Assert(errs["A"], HasError, validator.ErrBadParameter)
	c.Assert(errs["B"], HasError, validator.ErrBadParameter)
	c.Assert(errs["C"], HasError, validator.ErrBadParameter)
}

func (ms *MySuite) TestCopy(c *C) {
	v := validator.NewValidator()
	// WithTag calls copy, so we just copy the validator with the same tag
	v2 := v.WithTag("validate")
	// now we add a custom func only to the second one, it shouldn't get added
	// to the first
	v2.SetValidationFunc("custom", func(_ interface{}, _ string) error { return nil })
	type test struct {
		A string `validate:"custom"`
	}
	err := v2.Validate(test{})
	c.Assert(err, IsNil)

	err = v.Validate(test{})
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["A"], HasError, validator.ErrUnknownTag)
}

func (ms *MySuite) TestTagEscape(c *C) {
	type test struct {
		A string `validate:"min=0,regexp=^a{3\\,10}"`
	}
	t := test{"aaaa"}
	err := validator.Validate(t)
	c.Assert(err, IsNil)

	t2 := test{"aa"}
	err = validator.Validate(t2)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], HasError, validator.ErrRegexp)
}

func (ms *MySuite) TestEmbeddedFields(c *C) {
	type baseTest struct {
		A string `validate:"min=1"`
	}
	type test struct {
		baseTest
		B string `validate:"min=1"`
	}

	err := validator.Validate(test{})
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 2)
	c.Assert(errs["baseTest.A"], HasError, validator.ErrMin)
	c.Assert(errs["B"], HasError, validator.ErrMin)

	type test2 struct {
		baseTest `validate:"-"`
	}
	err = validator.Validate(test2{})
	c.Assert(err, IsNil)
}

func (ms *MySuite) TestEmbeddedPointerFields(c *C) {
	type baseTest struct {
		A string `validate:"min=1"`
	}
	type test struct {
		*baseTest
		B string `validate:"min=1"`
	}

	err := validator.Validate(test{baseTest: &baseTest{}})
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 2)
	c.Assert(errs["baseTest.A"], HasError, validator.ErrMin)
	c.Assert(errs["B"], HasError, validator.ErrMin)
}

func (ms *MySuite) TestEmbeddedNilPointerFields(c *C) {
	type baseTest struct {
		A string `validate:"min=1"`
	}
	type test struct {
		*baseTest
	}

	err := validator.Validate(test{})
	c.Assert(err, IsNil)
}

func (ms *MySuite) TestPrivateFields(c *C) {
	type test struct {
		b string `validate:"min=1"`
	}

	err := validator.Validate(test{
		b: "",
	})
	c.Assert(err, IsNil)
}

func (ms *MySuite) TestEmbeddedUnexported(c *C) {
	type baseTest struct {
		A string `validate:"min=1"`
	}
	type test struct {
		baseTest `validate:"nonnil"`
	}

	err := validator.Validate(test{})
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 2)
	c.Assert(errs["baseTest"], HasError, validator.ErrCannotValidate)
	c.Assert(errs["baseTest.A"], HasError, validator.ErrMin)
}

func (ms *MySuite) TestValidateStructWithByteSliceSlice(c *C) {
	type test struct {
		Slices [][]byte `validate:"len=1"`
	}

	t := test{
		Slices: [][]byte{[]byte(``)},
	}
	err := validator.Validate(t)
	c.Assert(err, IsNil)
}

func (ms *MySuite) TestEmbeddedInterface(c *C) {
	type test struct {
		I
	}

	err := validator.Validate(test{Impl2{"hello"}})
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["I.F"], HasError, validator.ErrLen)

	err = validator.Validate(test{&Impl{"hello"}})
	c.Assert(err, NotNil)
	errs, ok = err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["I.F"], HasError, validator.ErrLen)

	err = validator.Validate(test{})
	c.Assert(err, IsNil)

	type test2 struct {
		I `validate:"nonnil"`
	}
	err = validator.Validate(test2{})
	c.Assert(err, NotNil)
	errs, ok = err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["I"], HasError, validator.ErrZeroValue)
}

func (ms *MySuite) TestErrors(c *C) {
	err := validator.ErrorMap{
		"foo": validator.ErrorArray{
			fmt.Errorf("bar"),
		},
		"baz": validator.ErrorArray{
			fmt.Errorf("qux"),
		},
	}
	sep := ", "
	expected := "foo: bar, baz: qux"

	expectedParts := strings.Split(expected, sep)
	sort.Strings(expectedParts)

	errString := err.Error()
	errStringParts := strings.Split(errString, sep)
	sort.Strings(errStringParts)

	c.Assert(expectedParts, DeepEquals, errStringParts)
}

func (ms *MySuite) TestJSONPrint(c *C) {
	t := TestStruct{
		A: 0,
	}
	err := validator.WithPrintJSON(true).Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], IsNil)
	c.Assert(errs["a"], HasError, validator.ErrZeroValue)
}

func (ms *MySuite) TestPrintNestedJson(c *C) {
	t := TestCompositedStruct{
		Items: []NestedStruct{{}},
	}
	err := validator.WithPrintJSON(true).Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["a"], HasError, validator.ErrZeroValue)
	c.Assert(errs["otherNested.a"], HasError, validator.ErrZeroValue)
	c.Assert(errs["nestedItems[0].a"], HasError, validator.ErrZeroValue)
}

func (ms *MySuite) TestJSONPrintOff(c *C) {
	t := TestStruct{
		A: 0,
	}
	err := validator.WithPrintJSON(false).Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], HasError, validator.ErrZeroValue)
	c.Assert(errs["a"], IsNil)
}

func (ms *MySuite) TestJSONPrintNoTag(c *C) {
	t := TestStruct{
		B: "te",
	}
	err := validator.WithPrintJSON(true).Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["B"], HasError, validator.ErrLen)
}

func (ms *MySuite) TestValidateSlice(c *C) {
	type test2 struct {
		Num    int    `validate:"max=2"`
		String string `validate:"nonzero"`
	}

	err := validator.Validate([]test2{
		{
			Num:    6,
			String: "foo",
		},
		{
			Num:    1,
			String: "foo",
		},
	})
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["[0].Num"], HasError, validator.ErrMax)
	c.Assert(errs["[0].String"], IsNil) // sanity check
	c.Assert(errs["[1].Num"], IsNil)    // sanity check
	c.Assert(errs["[1].String"], IsNil) // sanity check
}

func (ms *MySuite) TestValidateMap(c *C) {
	type test2 struct {
		Num    int    `validate:"max=2"`
		String string `validate:"nonzero"`
	}

	err := validator.Validate(map[string]test2{
		"first": {
			Num:    6,
			String: "foo",
		},
		"second": {
			Num:    1,
			String: "foo",
		},
	})
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["[first](value).Num"], HasError, validator.ErrMax)
	c.Assert(errs["[first](value).String"], IsNil)  // sanity check
	c.Assert(errs["[second](value).Num"], IsNil)    // sanity check
	c.Assert(errs["[second](value).String"], IsNil) // sanity check

	err = validator.Validate(map[test2]string{
		{
			Num:    6,
			String: "foo",
		}: "first",
		{
			Num:    1,
			String: "foo",
		}: "second",
	})
	c.Assert(err, NotNil)
	errs, ok = err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["[{Num:6 String:foo}](key).Num"], HasError, validator.ErrMax)
	c.Assert(errs["[{Num:6 String:foo}](key).String"], IsNil) // sanity check
	c.Assert(errs["[{Num:1 String:foo}](key).Num"], IsNil)    // sanity check
	c.Assert(errs["[{Num:1 String:foo}](key).String"], IsNil) // sanity check
}

func (ms *MySuite) TestNonNilFunction(c *C) {
	type test struct {
		A func() `validate:"nonnil"`
	}

	err := validator.Validate(test{})
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["A"], HasError, validator.ErrZeroValue)

	err = validator.Validate(test{
		A: func() {},
	})
	c.Assert(err, IsNil)
}

func (ms *MySuite) TestTypeAliases(c *C) {
	type A string
	type B int64
	type test struct {
		A1 A  `validate:"regexp=^[0-9]+$"`
		A2 *A `validate:"regexp=^[0-9]+$"`
		B1 B  `validate:"min=10"`
		B2 B  `validate:"max=10"`
	}
	a123 := A("123")
	err := validator.Validate(test{
		A1: a123,
		A2: &a123,
		B1: B(11),
		B2: B(9),
	})
	c.Assert(err, IsNil)

	abc := A("abc")
	err = validator.Validate(test{
		A1: abc,
		A2: &abc,
		B2: B(11),
	})
	c.Assert(err, NotNil)
	errs, ok := err.(validator.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 4)
	c.Assert(errs["A1"], HasError, validator.ErrRegexp)
	c.Assert(errs["A2"], HasError, validator.ErrRegexp)
	c.Assert(errs["B1"], HasError, validator.ErrMin)
	c.Assert(errs["B2"], HasError, validator.ErrMax)
}

type hasErrorChecker struct {
	*CheckerInfo
}

func (c *hasErrorChecker) Check(params []interface{}, names []string) (bool, string) {
	var (
		ok    bool
		slice []error
		value error
	)
	slice, ok = params[0].(validator.ErrorArray)
	if !ok {
		return false, "First parameter is not an Errorarray"
	}
	value, ok = params[1].(error)
	if !ok {
		return false, "Second parameter is not an error"
	}

	for _, v := range slice {
		if v == value {
			return true, ""
		}
	}
	return false, ""
}

func (c *hasErrorChecker) Info() *CheckerInfo {
	return c.CheckerInfo
}

var HasError = &hasErrorChecker{&CheckerInfo{Name: "HasError", Params: []string{"HasError", "expected to contain"}}}
