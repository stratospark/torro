package bencoding

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type LexTest struct {
	Input      string
	StartState LexFn
	Result     []Token
}

var (
	tColon        = Token{TOKEN_COLON, ":"}
	tIntegerStart = Token{TOKEN_INTEGER_START, "i"}
	tIntegerEnd   = Token{TOKEN_INTEGER_END, "e"}
	tListStart    = Token{TOKEN_LIST_START, "l"}
	tListEnd      = Token{TOKEN_LIST_END, "e"}
	tDictStart    = Token{TOKEN_DICT_START, "d"}
	tDictEnd      = Token{TOKEN_DICT_END, "e"}
	tEOF          = Token{TOKEN_EOF, ""}
)

func TestStringLexing(t *testing.T) {
	validTests := []LexTest{
		LexTest{"4:spam", LexBegin, []Token{
			Token{TOKEN_STRING_LENGTH, "4"},
			tColon,
			Token{TOKEN_STRING_VALUE, "spam"},
			tEOF,
		}},
		LexTest{"0:", LexBegin, []Token{
			Token{TOKEN_STRING_LENGTH, "0"},
			tColon,
			Token{TOKEN_STRING_VALUE, ""},
			tEOF,
		}},
	}

	invalidTests := []LexTest{
		LexTest{"-1:a", LexBegin, []Token{
			Token{TOKEN_ERROR, LexErrInvalidCharacter},
		}},
		LexTest{"1.4:aa", LexBegin, []Token{
			Token{TOKEN_ERROR, LexErrInvalidStringLength},
		}},
		LexTest{"5:asdf", LexBegin, []Token{
			Token{TOKEN_STRING_LENGTH, "5"},
			tColon,
			Token{TOKEN_ERROR, LexErrUnexpectedEOF},
		}},
	}

	checkTests := func(tests []LexTest) {
		for _, test := range tests {
			Convey(fmt.Sprintf("%s", test.Input), func() {
				lex := BeginLexing(".torrent", test.Input, test.StartState)
				results := collect(lex)
				So(results, ShouldResemble, test.Result)
			})
		}
	}

	Convey("Given valid inputs", t, func() {
		checkTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkTests(invalidTests)
	})

}

func TestIntegerLexing(t *testing.T) {
	validTests := []LexTest{
		LexTest{"i3e", LexBegin, []Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "3"},
			tIntegerEnd,
			tEOF,
		}},
		LexTest{"i10e", LexBegin, []Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "10"},
			tIntegerEnd,
			tEOF,
		}},
		LexTest{"i-1e", LexBegin, []Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "-1"},
			tIntegerEnd,
			tEOF,
		}},
		LexTest{"i0e", LexBegin, []Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "0"},
			tIntegerEnd,
			tEOF,
		}},
	}

	invalidTests := []LexTest{
		LexTest{"iae", LexBegin, []Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "a"},
			tIntegerEnd,
			tEOF,
		}},
		LexTest{"i10", LexBegin, []Token{
			tIntegerStart,
			Token{TOKEN_ERROR, LexErrUnexpectedEOF},
		}},
	}

	checkTests := func(tests []LexTest) {
		for _, test := range tests {
			Convey(fmt.Sprintf("%s", test.Input), func() {
				lex := BeginLexing(".torrent", test.Input, test.StartState)
				results := collect(lex)
				So(results, ShouldResemble, test.Result)
			})
		}
	}

	Convey("Given valid inputs", t, func() {
		checkTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkTests(invalidTests)
	})

}

func TestListLexing(t *testing.T) {
	validTests := []LexTest{
		LexTest{"l4:spam4:eggse", LexBegin, []Token{
			tListStart,
			Token{TOKEN_STRING_LENGTH, "4"},
			tColon,
			Token{TOKEN_STRING_VALUE, "spam"},
			Token{TOKEN_STRING_LENGTH, "4"},
			tColon,
			Token{TOKEN_STRING_VALUE, "eggs"},
			tListEnd,
			tEOF,
		}},
		LexTest{"le", LexBegin, []Token{
			tListStart,
			tListEnd,
			tEOF,
		}},
		LexTest{"li10ei-1ee", LexBegin, []Token{
			tListStart,
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "10"},
			tIntegerEnd,
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "-1"},
			tIntegerEnd,
			tListEnd,
			tEOF,
		}},
		LexTest{"l4:thisi10el4:thati-1eee", LexBegin, []Token{
			tListStart,
			Token{TOKEN_STRING_LENGTH, "4"},
			tColon,
			Token{TOKEN_STRING_VALUE, "this"},
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "10"},
			tIntegerEnd,
			tListStart,
			Token{TOKEN_STRING_LENGTH, "4"},
			tColon,
			Token{TOKEN_STRING_VALUE, "that"},
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "-1"},
			tIntegerEnd,
			tListEnd,
			tListEnd,
			tEOF,
		}},
		LexTest{"llleee", LexBegin, []Token{
			tListStart,
			tListStart,
			tListStart,
			tListEnd,
			tListEnd,
			tListEnd,
			tEOF,
		}},
	}

	invalidTests := []LexTest{
		LexTest{"l", LexBegin, []Token{
			tListStart,
			Token{TOKEN_ERROR, LexErrUnclosedDelimeter},
		}},
		LexTest{"lle", LexBegin, []Token{
			tListStart,
			tListStart,
			tListEnd,
			Token{TOKEN_ERROR, LexErrUnclosedDelimeter},
		}},
	}

	checkTests := func(tests []LexTest) {
		for _, test := range tests {
			//			fmt.Println(test.Input)
			Convey(fmt.Sprintf("%s", test.Input), func() {
				lex := BeginLexing(".torrent", test.Input, test.StartState)
				results := collect(lex)
				So(results, ShouldResemble, test.Result)
			})
		}
	}

	Convey("Given valid inputs", t, func() {
		checkTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkTests(invalidTests)
	})

}

func TestDictLexing(t *testing.T) {
	validTests := []LexTest{
		LexTest{"d3:cow3:moo4:spam4:eggse", LexBegin, []Token{
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
		}},
		LexTest{"d4:spaml1:a1:bee", LexBegin, []Token{
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
		}},
		LexTest{"d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee", LexBegin, []Token{
			tDictStart,
			Token{TOKEN_STRING_LENGTH, "9"},
			tColon,
			Token{TOKEN_STRING_VALUE, "publisher"},
			Token{TOKEN_STRING_LENGTH, "3"},
			tColon,
			Token{TOKEN_STRING_VALUE, "bob"},
			Token{TOKEN_STRING_LENGTH, "17"},
			tColon,
			Token{TOKEN_STRING_VALUE, "publisher-webpage"},
			Token{TOKEN_STRING_LENGTH, "15"},
			tColon,
			Token{TOKEN_STRING_VALUE, "www.example.com"},
			Token{TOKEN_STRING_LENGTH, "18"},
			tColon,
			Token{TOKEN_STRING_VALUE, "publisher.location"},
			Token{TOKEN_STRING_LENGTH, "4"},
			tColon,
			Token{TOKEN_STRING_VALUE, "home"},
			tDictEnd,
			tEOF,
		}},
		LexTest{"de", LexBegin, []Token{
			tDictStart,
			tDictEnd,
			tEOF,
		}},
	}

	invalidTests := []LexTest{}

	checkTests := func(tests []LexTest) {
		for _, test := range tests {
			//			fmt.Println(test.Input)
			Convey(fmt.Sprintf("%s", test.Input), func() {
				lex := BeginLexing(".torrent", test.Input, test.StartState)
				results := collect(lex)
				So(results, ShouldResemble, test.Result)
			})
		}
	}

	Convey("Given valid inputs", t, func() {
		checkTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkTests(invalidTests)
	})

}
