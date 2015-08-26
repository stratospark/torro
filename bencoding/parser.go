package bencoding

import (
	"fmt"
	"github.com/oleiade/lane"
	"strconv"
)

type ContainerType int

const (
	ContainerBString ContainerType = iota
	ContainerInteger
	ContainerList
	ContainerDict
)

type Container struct {
	Type    ContainerType
	BString string
	Integer int
	List    *[]Container
	Dict    map[string]Container
}

func (c *Container) String() string {
	switch c.Type {
	case ContainerBString:
		return c.BString
	case ContainerInteger:
		return fmt.Sprint(c.Integer)
	case ContainerList:
		substrs := make([]string, 0)
		for _, subContainer := range *c.List {
			substrs = append(substrs, subContainer.String())
		}
		return fmt.Sprint(substrs)
	default:
		substrs := make([]string, 0)
		substrs = append(substrs, "{")
		for key, val := range c.Dict {
			substrs = append(substrs, fmt.Sprint(key, ":", val.String(), ", "))
		}
		substrs = append(substrs, "}")
		return fmt.Sprint(substrs)
	}
}

func (c *Container) SetKey(key string, val Container) {
	c.Dict[key] = val
}

func (c *Container) Append(val Container) {
	*c.List = append(*c.List, val)
}

func (c *Container) Collapse() interface{} {
	switch c.Type {
	case ContainerBString:
		return c.BString
	case ContainerInteger:
		return c.Integer
	case ContainerList:
		listItems := make([]interface{}, 0)
		for _, subContainer := range *c.List {
			listItems = append(listItems, subContainer.Collapse())
		}
		return listItems
	case ContainerDict:
		dict := make(map[string]interface{})
		for key, val := range c.Dict {
			dict[key] = val.Collapse()
		}
		return dict
	default:
		panic(fmt.Sprint("UNKNOWN CONTAINER TYPE ", c.Type))
	}
}

/*
Parser keeps track of parsing state, corresponding tokens,
output data structure, etc.
*/
type Parser struct {
	Tokens []Token
	State  ParseFn
	Output interface{}
	Stack  *lane.Stack

	Pos     int
	NextKey string
}

func (parser *Parser) CurrentType() TokenType {
	return parser.Tokens[parser.Pos].Type
}

func (parser *Parser) CurrentValue() string {
	return parser.Tokens[parser.Pos].Value
}

type ParseFn func(*Parser) ParseFn

/*
Parse takes a list of Tokens from the lexer and creates the final data structure.
*/
func Parse(tokens []Token) *Parser {
	parser := beginParsing(tokens, parseBegin)
	for parser.Pos < len(parser.Tokens) {
		if parser.State != nil {
			parser.State = parser.State(parser)
		} else {
			break
		}
	}
	return parser
}

/*
beginParsing initializes the Parser.
*/
func beginParsing(tokens []Token, state ParseFn) *Parser {
	p := &Parser{
		Tokens: tokens,
		State:  state,
		//		Output: make([]interface{}, 0),
		Stack: lane.NewStack(),
		Pos:   0,
	}

	//	p.CurrentContainer = p.Output
	return p
}

/*
parseBegin is the main state function to begin with and that
all other states eventually transition to.
*/
func parseBegin(parser *Parser) ParseFn {
	token := parser.Tokens[parser.Pos]
	switch token.Type {
	case TOKEN_STRING_LENGTH:
		return parseBString
	case TOKEN_INTEGER_START:
		return parseInteger
	case TOKEN_LIST_START:
		return parseList
	case TOKEN_DICT_START:
		return parseDict
	case TOKEN_LIST_END, TOKEN_DICT_END:
		// Pop stack so new items can be added to the parent container
		parser.Pos++
		if parser.Stack.Size() > 1 {
			parser.Stack.Pop()
		}
		return parseBegin
	case TOKEN_EOF:
		// TODO: Check if containers have been closed

		// Collapse root Container data structure into interface{}
		container := parser.Stack.Head().(*Container)
		parser.Output = container.Collapse()
		return nil
	default:
		// Some tokens should only be handled by other state functions
		fmt.Println("Current Token: ", parser.Pos, ", Total Tokens: ", len(parser.Tokens))
		panic(fmt.Sprint("UNEXPECTED TOKEN TYPE: ", token.Type))
	}
}

func addToContainer(parser *Parser, c *Container, key string) {
	if parser.Stack.Head() != nil {
		head := parser.Stack.Head().(*Container)
		switch head.Type {
		case ContainerBString:
			panic("CANT ADD TO STRING")
		case ContainerInteger:
			panic("CANT ADD TO INTEGER")
		case ContainerList:
			head.Append(*c)
			if c.Type == ContainerList || c.Type == ContainerDict {
				parser.Stack.Push(c)
			}
		case ContainerDict:
			if parser.NextKey != "" {
				head.SetKey(parser.NextKey, *c)
				parser.NextKey = ""
				if c.Type == ContainerList || c.Type == ContainerDict {
					parser.Stack.Push(c)
				}
			} else {
				if key != "" {
					parser.NextKey = key
				} else {
					panic("DICT KEY NOT SET")
				}
			}
		}
	} else {
		container := c
		parser.Stack.Push(container)
	}
}

func parseBString(parser *Parser) ParseFn {
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
	parser.Pos++
	addToContainer(parser, &Container{Type: ContainerBString, BString: strValue}, strValue)

	return parseBegin
}

func parseInteger(parser *Parser) ParseFn {
	//parser.CurrentValue() == TOKEN_INTEGER_START
	parser.Pos++

	num, err := strconv.ParseInt(parser.CurrentValue(), 10, 64)
	if err != nil {
		panic("NOT A VALID INTEGER")
	}
	parser.Pos++
	addToContainer(parser, &Container{Type: ContainerInteger, Integer: int(num)}, "")

	if parser.CurrentType() != TOKEN_INTEGER_END {
		panic("MISSING INTEGER END")
	}
	parser.Pos++

	return parseBegin
}

func parseList(parser *Parser) ParseFn {
	// "l"
	parser.Pos++

	list := make([]Container, 0)
	addToContainer(parser, &Container{Type: ContainerList, List: &list}, "")

	return parseBegin
}

func parseDict(parser *Parser) ParseFn {
	// "d"
	parser.Pos++

	dict := make(map[string]Container)
	addToContainer(parser, &Container{Type: ContainerDict, Dict: dict}, "")

	return parseBegin
}
