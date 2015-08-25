package bencoding

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type ParseTest struct {
	Input  []Token
	Result interface{}
}

/*
Take input vals of arbitrary types and return an array of them
*/
func makeResultList(vals ...interface{}) []interface{} {
	result := []interface{}{}
	for _, val := range vals {
		result = append(result, val)
	}
	return result
}

/*
Execute given tests
*/
func checkTests(tests []ParseTest) {
	for _, test := range tests {
		Convey(fmt.Sprintf("%s", test.Input), func() {
			result := Parse(test.Input)
			So(result.Output, ShouldResemble, test.Result)
		})
	}
}

func TestStringParsing(t *testing.T) {
	validTests := []ParseTest{
		ParseTest{[]Token{
			Token{TOKEN_STRING_LENGTH, "4"},
			tColon,
			Token{TOKEN_STRING_VALUE, "spam"},
			tEOF,
		}, "spam"},
		ParseTest{[]Token{
			Token{TOKEN_STRING_LENGTH, "0"},
			tColon,
			Token{TOKEN_STRING_VALUE, ""},
			tEOF,
		}, ""},
	}

	invalidTests := []ParseTest{}

	Convey("Given valid inputs", t, func() {
		checkTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkTests(invalidTests)
	})

}

func TestIntegerParsing(t *testing.T) {
	validTests := []ParseTest{
		ParseTest{[]Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "3"},
			tIntegerEnd,
			tEOF,
		}, 3},
		ParseTest{[]Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "10"},
			tIntegerEnd,
			tEOF,
		}, 10},
		ParseTest{[]Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "-1"},
			tIntegerEnd,
			tEOF,
		}, -1},
		ParseTest{[]Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "0"},
			tIntegerEnd,
			tEOF,
		}, 0},
	}

	invalidTests := []ParseTest{}

	Convey("Given valid inputs", t, func() {
		checkTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkTests(invalidTests)
	})

}

func TestListParsing(t *testing.T) {
	validTests := []ParseTest{
		ParseTest{[]Token{
			tListStart,
			Token{TOKEN_STRING_LENGTH, "4"},
			tColon,
			Token{TOKEN_STRING_VALUE, "spam"},
			Token{TOKEN_STRING_LENGTH, "4"},
			tColon,
			Token{TOKEN_STRING_VALUE, "eggs"},
			tListEnd,
			tEOF,
		}, makeResultList("spam", "eggs")},
		ParseTest{[]Token{
			tListStart,
			Token{TOKEN_STRING_LENGTH, "4"},
			tColon,
			Token{TOKEN_STRING_VALUE, "spam"},
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "10"},
			tIntegerEnd,
			tListEnd,
			tEOF,
		}, makeResultList("spam", 10)},
		ParseTest{[]Token{
			tListStart,
			Token{TOKEN_STRING_LENGTH, "3"},
			tColon,
			Token{TOKEN_STRING_VALUE, "hey"},
			tListStart,
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "1"},
			tIntegerEnd,
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "2"},
			tIntegerEnd,
			tListEnd,
			Token{TOKEN_STRING_LENGTH, "5"},
			tColon,
			Token{TOKEN_STRING_VALUE, "there"},
			tListEnd,
			tEOF,
		}, makeResultList("hey", makeResultList(1, 2), "there")},
		ParseTest{[]Token{
			tListStart,
			tListStart,
			tListStart,
			tListEnd,
			tListEnd,
			tListEnd,
			tEOF,
		}, makeResultList(makeResultList(makeResultList()))},
	}

	invalidTests := []ParseTest{}

	Convey("Given valid inputs", t, func() {
		checkTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkTests(invalidTests)
	})

}

func TestDictParsing(t *testing.T) {
	validTests := []ParseTest{
		ParseTest{
			[]Token{
				tDictStart,
				Token{TOKEN_STRING_LENGTH, "3"},
				tColon,
				Token{TOKEN_STRING_VALUE, "cow"},
				Token{TOKEN_STRING_LENGTH, "3"},
				tColon,
				Token{TOKEN_STRING_VALUE, "moo"},
				Token{TOKEN_STRING_LENGTH, "4"},
				tColon,
				Token{TOKEN_STRING_VALUE, "spam"},
				Token{TOKEN_STRING_LENGTH, "4"},
				tColon,
				Token{TOKEN_STRING_VALUE, "eggs"},
				tDictEnd,
				tEOF,
			}, map[string]interface{}{"cow": "moo", "spam": "eggs"},
		},
		ParseTest{
			[]Token{
				tDictStart,
				Token{TOKEN_STRING_LENGTH, "4"},
				tColon,
				Token{TOKEN_STRING_VALUE, "spam"},
				tListStart,
				Token{TOKEN_STRING_LENGTH, "1"},
				tColon,
				Token{TOKEN_STRING_VALUE, "a"},
				Token{TOKEN_STRING_LENGTH, "1"},
				tColon,
				Token{TOKEN_STRING_VALUE, "b"},
				tListEnd,
				tDictEnd,
				tEOF,
			}, map[string]interface{}{"spam": makeResultList("a", "b")},
		},
		ParseTest{
			[]Token{
				tDictStart,
				Token{TOKEN_STRING_LENGTH, "4"},
				tColon,
				Token{TOKEN_STRING_VALUE, "dict"},
				tDictStart,
				Token{TOKEN_STRING_LENGTH, "1"},
				tColon,
				Token{TOKEN_STRING_VALUE, "a"},
				tListStart,
				tIntegerStart,
				Token{TOKEN_INTEGER_VALUE, "10"},
				tIntegerEnd,
				Token{TOKEN_STRING_LENGTH, "1"},
				tColon,
				Token{TOKEN_STRING_VALUE, "b"},
				tListEnd,
				tDictEnd,
				Token{TOKEN_STRING_LENGTH, "3"},
				tColon,
				Token{TOKEN_STRING_VALUE, "int"},
				tIntegerStart,
				Token{TOKEN_INTEGER_VALUE, "99"},
				tIntegerEnd,
				tDictEnd,
				tEOF,
			}, map[string]interface{}{"dict": map[string]interface{}{"a": makeResultList(10, "b")}, "int": 99},
		},
	}

	invalidTests := []ParseTest{}

	Convey("Given valid inputs", t, func() {
		checkTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkTests(invalidTests)
	})
}
