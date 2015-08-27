package bencoding

import "fmt"

type Token struct {
	Type  TokenType
	Value []byte
}

func NewToken(tokenType TokenType, val string) Token {
	return Token{
		Type:  tokenType,
		Value: []byte(val),
	}
}

type TokenType int

const (
	TOKEN_ERROR TokenType = iota
	TOKEN_EOF

	TOKEN_COLON

	TOKEN_STRING_LENGTH
	TOKEN_STRING_VALUE

	TOKEN_INTEGER_START
	TOKEN_INTEGER_VALUE
	TOKEN_INTEGER_END

	TOKEN_LIST_START
	TOKEN_LIST_VALUE
	TOKEN_LIST_END

	TOKEN_DICT_START
	TOKEN_DICT_VALUE
	TOKEN_DICT_END
)

var TokenNames = map[TokenType]string{
	TOKEN_ERROR: "ERROR",
	TOKEN_EOF:   "EOF",

	TOKEN_COLON: "COLON",

	TOKEN_STRING_LENGTH: "STRING_LENGTH",
	TOKEN_STRING_VALUE:  "STRING_VALUE",

	TOKEN_INTEGER_START: "INTEGER START",
	TOKEN_INTEGER_VALUE: "INTEGER_VALUE",
	TOKEN_INTEGER_END:   "INTEGER_END",

	TOKEN_LIST_START: "LIST_START",
	TOKEN_LIST_VALUE: "LIST_VALUE",
	TOKEN_LIST_END:   "LIST_END",

	TOKEN_DICT_START: "DICT_START",
	TOKEN_DICT_VALUE: "DICT_VALUE",
	TOKEN_DICT_END:   "DICT_END",
}

func (t Token) String() string {
	maxLen := 40
	value := string(t.Value)
	if len(t.Value) > maxLen {
		value = value[:maxLen] + "..." + value[len(value)-10:]
	}
	output := fmt.Sprintf("[%s: %q]", TokenNames[t.Type], value)
	return output
}

//const EOF byte = -1

const (
	COLON         string = ":"
	INTEGER_START string = "i"
	INTEGER_END   string = "e"
	LIST_START    string = "l"
	LIST_END      string = "e"
	DICT_START    string = "d"
	DICT_END      string = "e"
)
