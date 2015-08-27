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
		switch val.(type) {
		case string:
			result = append(result, []byte(val.(string)))
		default:
			result = append(result, val)
		}
	}
	return result
}

/*
Execute given tests
*/
func checkParseTests(tests []ParseTest) {
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
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "spam"),
			tEOF,
		}, []byte("spam")},
		ParseTest{[]Token{
			NewToken(TOKEN_STRING_LENGTH, "0"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, ""),
			tEOF,
		}, []byte("")},
	}

	invalidTests := []ParseTest{}

	Convey("Given valid inputs", t, func() {
		checkParseTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkParseTests(invalidTests)
	})

}

func TestIntegerParsing(t *testing.T) {
	validTests := []ParseTest{
		ParseTest{[]Token{
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "3"),
			tIntegerEnd,
			tEOF,
		}, 3},
		ParseTest{[]Token{
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "10"),
			tIntegerEnd,
			tEOF,
		}, 10},
		ParseTest{[]Token{
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "-1"),
			tIntegerEnd,
			tEOF,
		}, -1},
		ParseTest{[]Token{
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "0"),
			tIntegerEnd,
			tEOF,
		}, 0},
	}

	invalidTests := []ParseTest{}

	Convey("Given valid inputs", t, func() {
		checkParseTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkParseTests(invalidTests)
	})

}

func TestListParsing(t *testing.T) {
	validTests := []ParseTest{
		ParseTest{[]Token{
			tListStart,
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "spam"),
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "eggs"),
			tListEnd,
			tEOF,
		}, makeResultList("spam", "eggs")},
		ParseTest{[]Token{
			tListStart,
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "spam"),
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "10"),
			tIntegerEnd,
			tListEnd,
			tEOF,
		}, makeResultList("spam", 10)},
		ParseTest{[]Token{
			tListStart,
			NewToken(TOKEN_STRING_LENGTH, "3"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "hey"),
			tListStart,
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "1"),
			tIntegerEnd,
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "2"),
			tIntegerEnd,
			tListEnd,
			NewToken(TOKEN_STRING_LENGTH, "5"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "there"),
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
		checkParseTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkParseTests(invalidTests)
	})

}

func TestDictParsing(t *testing.T) {
	validTests := []ParseTest{
		ParseTest{
			[]Token{
				tDictStart,
				NewToken(TOKEN_STRING_LENGTH, "3"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "cow"),
				NewToken(TOKEN_STRING_LENGTH, "3"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "moo"),
				NewToken(TOKEN_STRING_LENGTH, "4"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "spam"),
				NewToken(TOKEN_STRING_LENGTH, "4"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "eggs"),
				tDictEnd,
				tEOF,
			}, map[string]interface{}{"cow": []byte("moo"), "spam": []byte("eggs")},
		},
		ParseTest{
			[]Token{
				tDictStart,
				NewToken(TOKEN_STRING_LENGTH, "4"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "spam"),
				tListStart,
				NewToken(TOKEN_STRING_LENGTH, "1"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "a"),
				NewToken(TOKEN_STRING_LENGTH, "1"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "b"),
				tListEnd,
				tDictEnd,
				tEOF,
			}, map[string]interface{}{"spam": makeResultList("a", "b")},
		},
		ParseTest{
			[]Token{
				tDictStart,
				NewToken(TOKEN_STRING_LENGTH, "4"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "dict"),
				tDictStart,
				NewToken(TOKEN_STRING_LENGTH, "1"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "a"),
				tListStart,
				tIntegerStart,
				NewToken(TOKEN_INTEGER_VALUE, "10"),
				tIntegerEnd,
				NewToken(TOKEN_STRING_LENGTH, "1"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "b"),
				tListEnd,
				tDictEnd,
				NewToken(TOKEN_STRING_LENGTH, "3"),
				tColon,
				NewToken(TOKEN_STRING_VALUE, "int"),
				tIntegerStart,
				NewToken(TOKEN_INTEGER_VALUE, "99"),
				tIntegerEnd,
				tDictEnd,
				tEOF,
			}, map[string]interface{}{"dict": map[string]interface{}{"a": makeResultList(10, "b")}, "int": 99},
		},
	}

	invalidTests := []ParseTest{}

	Convey("Given valid inputs", t, func() {
		checkParseTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkParseTests(invalidTests)
	})
}
