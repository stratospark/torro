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

func collect(lex *Lexer) (tokens []Token) {
	for {
		token := lex.NextToken()
		tokens = append(tokens, token)
		if token.Type == TOKEN_EOF || token.Type == TOKEN_ERROR {
			break
		}
	}
	return
}

var (
	tColon        = Token{TOKEN_COLON, ":"}
	tIntegerStart = Token{TOKEN_INTEGER_START, "i"}
	tIntegerEnd   = Token{TOKEN_INTEGER_END, "e"}
	tListStart    = Token{TOKEN_LIST_START, "l"}
	tListEnd      = Token{TOKEN_LIST_END, "e"}
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

	invalidTests := []LexTest{}

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
		LexTest{"i04e", LexBegin, []Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "04"},
			tIntegerEnd,
			tEOF,
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
	}

	invalidTests := []LexTest{}

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
