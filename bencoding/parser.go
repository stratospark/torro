package bencoding

import (
	"strconv"
	"fmt"
)

type Parser struct {
	Tokens []Token
	State ParseFn
	Output []interface{}

	Pos int
}

func (parser *Parser) CurrentType() TokenType {
	return parser.Tokens[parser.Pos].Type
}

func (parser *Parser) CurrentValue() string {
	return parser.Tokens[parser.Pos].Value
}

type ParseFn func(*Parser) ParseFn

func Parse(tokens []Token) *Parser {
	parser := beginParsing(tokens, parseBegin)
	for parser.Pos < len(parser.Tokens) {
		fmt.Println(len(parser.Tokens), parser.Pos)
		fmt.Println("Loop")
		if parser.State != nil {
			parser.State = parser.State(parser)
		} else {
			break
		}
	}
	return parser
}

func beginParsing(tokens []Token, state ParseFn) *Parser {
	p := &Parser{
		Tokens: tokens,
		State: state,
		Output: make([]interface{}, 0),
		Pos: 0,
	}
	return p
}

func parseBegin(parser *Parser) ParseFn {
	fmt.Println("ParseBegin")
	token := parser.Tokens[parser.Pos]
	switch token.Type {
	case TOKEN_STRING_LENGTH:
		return parseString
	case TOKEN_INTEGER_START:
		return parseInteger
	case TOKEN_COLON:
		// shouldn't get here directly
		return nil
	case TOKEN_STRING_VALUE:
		// shouldn't get here directly
		return nil
	default:
		return nil
	}
}

func parseString(parser *Parser) ParseFn {
	fmt.Println("ParseString")
	// Get Length
	strLength, err := strconv.ParseInt(parser.CurrentValue(), 10, 64)
	if err != nil {
		panic("NOT A VALID STRING LENGTH")
	}
	parser.Pos++

	// Get Colon
	colon := parser.CurrentValue()
	if colon != ":" {
		panic("MISSING REQUIRED COLON")
	}
	parser.Pos++

	// Get Value
	strValue := parser.CurrentValue()
	if len(strValue) != int(strLength) {
		panic("STRING LENGTH DOESNT MATCH")
	}
	fmt.Println(strValue)
	parser.Output = append(parser.Output, strValue)
	parser.Pos++

	return parseBegin
}

func parseInteger(parser *Parser) ParseFn {
	fmt.Println("ParseInteger")

	//parser.CurrentValue() == TOKEN_INTEGER_START
	parser.Pos++

	num, err := strconv.ParseInt(parser.CurrentValue(), 10, 64)
	if err != nil {
		panic("NOT A VALID INTEGER")
	}
	parser.Output = append(parser.Output, int(num))
	parser.Pos++

	if parser.CurrentType() != TOKEN_INTEGER_END {
		panic("MISSING INTEGER END")
	}
	parser.Pos++

	return parseBegin
}
