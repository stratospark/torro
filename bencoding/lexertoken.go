package bencoding

type Token struct {
	Type TokenType
	Value string
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

const EOF rune = 0

const (
	COLON string = ":"
	INTEGER_START string = "i"
	INTEGER_END string = "e"
	LIST_START string = "l"
	LIST_END string = "e"
	DICT_START string = "d"
	DICT_END string = "e"
)






/*
9:publisher
d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee represents { "publisher" => "bob", "publisher-webpage" => "www.example.com", "publisher.location" => "home" }


{
	"publisher: "bob",
	"publisher-webpage": "www.....",
	"publisher.location": ["loc1", "loc2"]
}

[
TOKEN_DICT_START,

TOKEN_STRING_LENGTH,
TOKEN_COLON,
TOKEN_STRING_VALUE,

TOKEN_STRING_LENGTH,
TOKEN_COLON,
TOKEN_STRING_VALUE,
]


 */