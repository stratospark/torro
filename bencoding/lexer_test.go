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
	tColon        = Token{TOKEN_COLON, []byte{':'}}
	tIntegerStart = Token{TOKEN_INTEGER_START, []byte{'i'}}
	tIntegerEnd   = Token{TOKEN_INTEGER_END, []byte{'e'}}
	tListStart    = Token{TOKEN_LIST_START, []byte{'l'}}
	tListEnd      = Token{TOKEN_LIST_END, []byte{'e'}}
	tDictStart    = Token{TOKEN_DICT_START, []byte{'d'}}
	tDictEnd      = Token{TOKEN_DICT_END, []byte{'e'}}
	tEOF          = Token{TOKEN_EOF, []byte{}}
)

func checkLexTests(tests []LexTest) {
	for _, test := range tests {
		Convey(fmt.Sprintf("%s", test.Input), func() {
			lex := BeginLexing(".torrent", test.Input, test.StartState)
			results := collect(lex)
			lex.Shutdown()
			So(results, ShouldResemble, test.Result)
		})
	}
}

func TestLexer(t *testing.T) {
	Convey("Creating a basic Lexer", t, func() {
		lex := &Lexer{}
		So(*lex, ShouldNotBeNil)
		So(lex.String(), ShouldContainSubstring, "Name")
		So(lex.String(), ShouldContainSubstring, "Input")
		So(lex.String(), ShouldContainSubstring, "Start")
		So(lex.String(), ShouldContainSubstring, "Pos")
		So(lex.String(), ShouldContainSubstring, "Width")
	})
}

func TestStringLexing(t *testing.T) {
	validTests := []LexTest{
		LexTest{"4:spam", LexBegin, []Token{
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "spam"),
			tEOF,
		}},
		LexTest{"0:", LexBegin, []Token{
			NewToken(TOKEN_STRING_LENGTH, "0"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, ""),
			tEOF,
		}},
		LexTest{"1::", LexBegin, []Token{
			NewToken(TOKEN_STRING_LENGTH, "1"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, ":"),
			tEOF,
		}},
	}

	invalidTests := []LexTest{
		LexTest{"-1:a", LexBegin, []Token{
			NewToken(TOKEN_ERROR, LexErrInvalidCharacter),
		}},
		LexTest{"1.4:aa", LexBegin, []Token{
			NewToken(TOKEN_ERROR, LexErrInvalidStringLength),
		}},
		LexTest{"5:asdf", LexBegin, []Token{
			NewToken(TOKEN_STRING_LENGTH, "5"),
			tColon,
			NewToken(TOKEN_ERROR, LexErrUnexpectedEOF),
		}},
		LexTest{"5asdfg", LexBegin, []Token{
			NewToken(TOKEN_ERROR, LexErrUnexpectedEOF),
		}},
		LexTest{"5", LexBegin, []Token{
			NewToken(TOKEN_ERROR, LexErrUnexpectedEOF),
		}},
	}

	Convey("Given valid inputs", t, func() {
		checkLexTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkLexTests(invalidTests)
	})

}

func TestIntegerLexing(t *testing.T) {
	validTests := []LexTest{
		LexTest{"i3e", LexBegin, []Token{
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "3"),
			tIntegerEnd,
			tEOF,
		}},
		LexTest{"i10e", LexBegin, []Token{
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "10"),
			tIntegerEnd,
			tEOF,
		}},
		LexTest{"i-1e", LexBegin, []Token{
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "-1"),
			tIntegerEnd,
			tEOF,
		}},
		LexTest{"i0e", LexBegin, []Token{
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "0"),
			tIntegerEnd,
			tEOF,
		}},
	}

	invalidTests := []LexTest{
		LexTest{"iae", LexBegin, []Token{
			tIntegerStart,
			NewToken(TOKEN_ERROR, LexErrInvalidCharacter),
		}},
		LexTest{"i10", LexBegin, []Token{
			tIntegerStart,
			NewToken(TOKEN_ERROR, LexErrUnexpectedEOF),
		}},
		LexTest{"ie", LexBegin, []Token{
			tIntegerStart,
			NewToken(TOKEN_ERROR, LexErrInvalidCharacter),
		}},
		LexTest{"i1.1e", LexBegin, []Token{
			tIntegerStart,
			NewToken(TOKEN_ERROR, LexErrInvalidCharacter),
		}},
	}

	Convey("Given valid inputs", t, func() {
		checkLexTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkLexTests(invalidTests)
	})

}

func TestListLexing(t *testing.T) {
	validTests := []LexTest{
		LexTest{"l4:spam4:eggse", LexBegin, []Token{
			tListStart,
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "spam"),
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "eggs"),
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
			NewToken(TOKEN_INTEGER_VALUE, "10"),
			tIntegerEnd,
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "-1"),
			tIntegerEnd,
			tListEnd,
			tEOF,
		}},
		LexTest{"l4:thisi10el4:thati-1eee", LexBegin, []Token{
			tListStart,
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "this"),
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "10"),
			tIntegerEnd,
			tListStart,
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "that"),
			tIntegerStart,
			NewToken(TOKEN_INTEGER_VALUE, "-1"),
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
			NewToken(TOKEN_ERROR, LexErrUnclosedDelimeter),
		}},
		LexTest{"lle", LexBegin, []Token{
			tListStart,
			tListStart,
			tListEnd,
			NewToken(TOKEN_ERROR, LexErrUnclosedDelimeter),
		}},
		LexTest{"lleee", LexBegin, []Token{
			tListStart,
			tListStart,
			tListEnd,
			tListEnd,
			NewToken(TOKEN_ERROR, LexErrInvalidCharacter),
		}},
	}

	Convey("Given valid inputs", t, func() {
		checkLexTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkLexTests(invalidTests)
	})

}

func TestDictLexing(t *testing.T) {
	validTests := []LexTest{
		LexTest{"d3:cow3:moo4:spam4:eggse", LexBegin, []Token{
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
		}},
		LexTest{"d4:spaml1:a1:bee", LexBegin, []Token{
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
		}},
		LexTest{"d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee", LexBegin, []Token{
			tDictStart,
			NewToken(TOKEN_STRING_LENGTH, "9"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "publisher"),
			NewToken(TOKEN_STRING_LENGTH, "3"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "bob"),
			NewToken(TOKEN_STRING_LENGTH, "17"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "publisher-webpage"),
			NewToken(TOKEN_STRING_LENGTH, "15"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "www.example.com"),
			NewToken(TOKEN_STRING_LENGTH, "18"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "publisher.location"),
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "home"),
			tDictEnd,
			tEOF,
		}},
		LexTest{"d3:key3:val4:infod6:pieces4:blah4:dictdee6:locale2:ene", LexBegin, []Token{
			tDictStart,
			NewToken(TOKEN_STRING_LENGTH, "3"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "key"),
			NewToken(TOKEN_STRING_LENGTH, "3"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "val"),
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "info"),
			tDictStart,
			NewToken(TOKEN_STRING_LENGTH, "6"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "pieces"),
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "blah"),
			NewToken(TOKEN_STRING_LENGTH, "4"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "dict"),
			tDictStart,
			tDictEnd,
			tDictEnd,
			NewToken(TOKEN_STRING_LENGTH, "6"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "locale"),
			NewToken(TOKEN_STRING_LENGTH, "2"),
			tColon,
			NewToken(TOKEN_STRING_VALUE, "en"),
			tDictEnd,
			tEOF,
		}},
		LexTest{"de", LexBegin, []Token{
			tDictStart,
			tDictEnd,
			tEOF,
		}},
	}

	invalidTests := []LexTest{
		LexTest{"d", LexBegin, []Token{
			tDictStart,
			NewToken(TOKEN_ERROR, LexErrUnclosedDelimeter),
		}},
		LexTest{"dee", LexBegin, []Token{
			tDictStart,
			tDictEnd,
			NewToken(TOKEN_ERROR, LexErrInvalidCharacter),
		}},
	}

	Convey("Given valid inputs", t, func() {
		checkLexTests(validTests)
	})

	Convey("Given invalid inputs", t, func() {
		checkLexTests(invalidTests)
	})

}

func TestGetBencodedInfo(t *testing.T) {
	Convey("Returns the original bencoded info dictionary from tokens", t, func() {
		Convey("Using artifical data", func() {
			input := "d4:infod3:keyd1:x1:yeed1:a1:bee"
			lex := BeginLexing(".torrent", input, LexBegin)
			tokens := collect(lex)
			lex.Shutdown()
			info := GetBencodedInfo(tokens)
			So(info, ShouldResemble, []byte("d3:keyd1:x1:ye"))
		})

		Convey("Using data extracted from torrent file", func() {
			input := "d8:announce39:http://torrent.ubuntu.com:6969/announce13:announce-listll39:http://torrent.ubuntu.com:6969/announceel44:http://ipv6.torrent.ubuntu.com:6969/announceee7:comment29:Ubuntu CD releases.ubuntu.com13:creation datei1406245935e4:infod6:lengthi1028653056e4:name32:ubuntu-14.04.1-desktop-amd64.iso12:piece lengthi524288eee"
			lex := BeginLexing(".torrent", input, LexBegin)
			tokens := collect(lex)
			lex.Shutdown()
			info := GetBencodedInfo(tokens)
			So(info, ShouldResemble, []byte("d6:lengthi1028653056e4:name32:ubuntu-14.04.1-desktop-amd64.iso12:piece lengthi524288e"))
		})
	})
}
