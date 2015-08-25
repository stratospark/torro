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
	ContainerMap
)

type Container struct {
	Type    ContainerType
	BString string
	Integer int
	List    *[]Container
	Map     map[string]interface{}
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
		return "?"
	}
}

func (c *Container) SetBString(val string) {
	c.BString = val
}

func (c *Container) SetInteger(val int) {
	c.Integer = val
}

func (c *Container) SetKey(key string, val interface{}) {
	c.Map[key] = val
}

func (c *Container) Append(val Container) {
	*c.List = append(*c.List, val)
}

func (c *Container) Collapse() interface{} {
	fmt.Println("COLLAPSE", c)
	switch c.Type {
	case ContainerBString:
		fmt.Println("Returning Bstring: ", c.BString)
		return c.BString
	case ContainerInteger:
		fmt.Println("Returning Integer: ", c.Integer)
		return c.Integer
	case ContainerList:
		listItems := make([]interface{}, 0)
		for _, subContainer := range *c.List {
			listItems = append(listItems, subContainer.Collapse())
		}
		fmt.Println("Returning List: ", listItems)
		return listItems
	default:
		return "asdf"
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

	Pos int
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
	fmt.Println("ParseBegin ", parser.Stack.Head())
	token := parser.Tokens[parser.Pos]
	switch token.Type {
	case TOKEN_STRING_LENGTH:
		return parseBString
	case TOKEN_INTEGER_START:
		return parseInteger
	case TOKEN_LIST_START:
		return parseList
	case TOKEN_LIST_END:
		parser.Pos++
		fmt.Println("POPPING STACK")
		fmt.Println("OLD HEAD ", parser.Stack.Head())
		if parser.Stack.Size() > 1 {
			parser.Stack.Pop()
		}
		fmt.Println("New HEAD ", parser.Stack.Head())
		return parseBegin
	case TOKEN_COLON:
		// shouldn't get here directly
		return nil
	case TOKEN_STRING_VALUE:
		// shouldn't get here directly
		return nil
	default:
		fmt.Println("STACK SIZE ", parser.Stack.Size())
		container := parser.Stack.Head().(*Container)
		fmt.Println("DEFAULT CONTAINER", container)
		parser.Output = container.Collapse()
		fmt.Println("DEFAULT", parser.Output)
		return nil
	}
}

func parseBString(parser *Parser) ParseFn {
	fmt.Println("ParseBString")
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

	if parser.Stack.Head() != nil {
		head := parser.Stack.Head().(*Container)
		switch head.Type {
		case ContainerBString:
			panic("CANT ADD TO STRING")
		case ContainerInteger:
			panic("CANT ADD TO INTEGER")
		case ContainerList:
			head.Append(Container{Type: ContainerBString, BString: strValue})
		case ContainerMap:
			break
		}
	} else {
		container := &Container{Type: ContainerBString, BString: strValue}
		parser.Stack.Push(container)
	}

	fmt.Println("Output: ", parser.Output)

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

	if parser.Stack.Head() != nil {
		head := parser.Stack.Head().(*Container)
		switch head.Type {
		case ContainerBString:
			panic("CANT ADD TO STRING")
		case ContainerInteger:
			panic("CANT ADD TO INTEGER")
		case ContainerList:
			head.Append(Container{Type: ContainerInteger, Integer: int(num)})
			fmt.Println("INT APPEND, ")

			fmt.Println("STACK SIZE ", parser.Stack.Size())
		case ContainerMap:
			break
		}
	} else {
		container := Container{Type: ContainerInteger, Integer: int(num)}
		parser.Stack.Push(&container)
	}
	fmt.Println(num)
	parser.Pos++

	if parser.CurrentType() != TOKEN_INTEGER_END {
		panic("MISSING INTEGER END")
	}
	parser.Pos++

	return parseBegin
}

func parseList(parser *Parser) ParseFn {
	fmt.Println("ParseList")

	// "l"
	parser.Pos++

	list := make([]Container, 0)
	if parser.Stack.Head() != nil {
		head := parser.Stack.Head().(*Container)
		switch head.Type {
		case ContainerBString:
			panic("CANT ADD TO STRING")
		case ContainerInteger:
			panic("CANT ADD TO INTEGER")
		case ContainerList:
			container := Container{Type: ContainerList, List: &list}
			head.Append(container)
			parser.Stack.Push(&container)
			fmt.Println("STACK SIZE ", parser.Stack.Size())
			//			fmt.Println("OLD HEAD, ", head, " NEW HEAD ", parser.Stack.Head())
		case ContainerMap:
			break
		}
	} else {
		container := Container{Type: ContainerList, List: &list}
		parser.Stack.Push(&container)
	}

	return parseBegin
}
