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

//var tokenNames = map[TokenType]string{
//	TOKEN_ERROR: "ERROR",
//	TOKEN_EOF:   "EOF",
//
//	TOKEN_COLON: "COLON",
//
//	TOKEN_STRING_LENGTH: "STRING_LENGTH",
//	TOKEN_STRING_VALUE:  "STRING_VALUE",
//
//	TOKEN_INTEGER_START: "INTEGER START",
//	TOKEN_INTEGER_VALUE: "INTEGER_VALUE",
//	TOKEN_INTEGER_END:   "INTEGER_END",
//
//	TOKEN_LIST_START: "LIST_START",
//	TOKEN_LIST_VALUE: "LIST_VALUE",
//	TOKEN_LIST_END:   "LIST_END",
//
//	TOKEN_DICT_START: "DICT_START",
//	TOKEN_DICT_VALUE: "DICT_VALUE",
//	TOKEN_DICT_END:   "DICT_END",
//}


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
	tIntegerStart = Token{TOKEN_INTEGER_START, "i"}
	tIntegerEnd   = Token{TOKEN_INTEGER_END, "e"}
	tEOF          = Token{TOKEN_EOF, ""}
)

func TestStringLexing(t *testing.T) {
	validTests := []LexTest{
		LexTest{"4:spam", LexStringStart, []Token{
			Token{TOKEN_STRING_LENGTH, "4"},
			Token{TOKEN_COLON, ":"},
			Token{TOKEN_STRING_VALUE, "spam"},
			tEOF,
		}},
		LexTest{"0:", LexStringStart, []Token {
			Token{TOKEN_STRING_LENGTH, "0"},
			Token{TOKEN_COLON, ":"},
			Token{TOKEN_STRING_VALUE: ""},
			tEOF,
		}},
	}

	invalidTests := []LexTest{
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
		LexTest{"i3e", LexIntegerStart, []Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "3"},
			tIntegerEnd,
			tEOF,
		}},
		LexTest{"i10e", LexIntegerStart, []Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "10"},
			tIntegerEnd,
			tEOF,
		}},
		LexTest{"i-1e", LexIntegerStart, []Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "-1"},
			tIntegerEnd,
			tEOF,
		}},
		LexTest{"i0e", LexIntegerStart, []Token{
			tIntegerStart,
			Token{TOKEN_INTEGER_VALUE, "0"},
			tIntegerEnd,
			tEOF,
		}},
	}

	invalidTests := []LexTest{
		LexTest{"i04e", LexIntegerStart, []Token{
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
