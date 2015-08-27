package bencoding

import (
	"fmt"
	"github.com/oleiade/lane"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	Name   string
	Input  []byte
	Tokens chan Token
	State  LexFn

	Start        int
	Pos          int
	Width        int
	StringLength int
	NestedStack  *lane.Stack
}

var (
	LexErrInvalidStringLength string = "Invalid String Length"
	LexErrInvalidCharacter    string = "Invalid Character"
	LexErrUnclosedDelimeter          = "Unclosed Delimeter"
	LexErrUnexpectedEOF              = "Unexpected EOF"
)

func (lex *Lexer) String() string {
	return fmt.Sprintf("Name: %s, Input: %s, Start: %d, Pos: %d, Width: %d",
		lex.Name,
		lex.Input,
		lex.Start,
		lex.Pos,
		lex.Width,
	)
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

var Collect = collect

type LexFn func(*Lexer) LexFn

/*
Backup to the beginning of the last read token.
*/
func (lex *Lexer) Backup() {
	//	lex.Pos -= lex.Width
	lex.Pos--
}

/*
Returns a slice of the current input from the current lexer
start position to the current position.
*/
func (lex *Lexer) CurrentInput() []byte {
	return lex.Input[lex.Start:lex.Pos]
}

/*
Decrement the position
*/
func (lex *Lexer) Dec() {
	lex.Pos--
}

/*
Puts a token on the token channel. The value of this
token  is read from the input based on the current lexer position.
*/
func (lex *Lexer) Emit(tokenType TokenType) {
	token := Token{Type: tokenType, Value: lex.Input[lex.Start:lex.Pos]}
	lex.Tokens <- token
	lex.Start = lex.Pos
}

/*
Returns a token with error information.
*/
func (lex *Lexer) Errorf(format string, args ...interface{}) LexFn {
	lex.Tokens <- Token{
		Type:  TOKEN_ERROR,
		Value: []byte(fmt.Sprintf(format, args...)),
	}
	return nil
}

/*
Ignores the current token by setting the lexer's start position
to the current reading position.
*/
func (lex *Lexer) Ignore() {
	lex.Start = lex.Pos
}

/*
Return a slice of the input from the current lexer position
to the end of the input string.
*/
func (lex *Lexer) InputToEnd() []byte {
	return lex.Input[lex.Pos:]
}

/*
Returns the true/false if the lexer is at the end of the input stream.
*/
func (lex *Lexer) IsEOF() bool {
	return lex.Pos >= len(lex.Input)
}

/*
Reads the next rune (character) from the input stream
and advances the lexer position.
*/
//func (lex *Lexer) Next() rune {
//	if lex.Pos >= utf8.RuneCountInString(lex.Input) {
//		lex.Width = 0
//		fmt.Println("               EOF                   ")
//		return EOF
//	}
//
//	result, width := utf8.DecodeRuneInString(lex.Input[lex.Pos:])
//
//	fmt.Println("NEXT: ", result, " WIDTH: ", width)
//	lex.Width = width
//	lex.Pos += lex.Width
//	return result
//}
func (lex *Lexer) Next() byte {
	next := lex.Input[lex.Pos : lex.Pos+1]
	lex.Pos++
	return next[0]
}

/*
Return the next token from the channel
*/
func (lex *Lexer) NextToken() Token {
	for {
		select {
		case token := <-lex.Tokens:
			return token
		default:
			lex.State = lex.State(lex)

		}
	}
	panic("Lexer.NextToken reached an invalid state!")
}

/*
Returns the next rune in the stream, then puts the lexer
position back. Basically reads the next rune without consuming it.
*/
func (lex *Lexer) Peek() byte {
	r := lex.Next()
	lex.Backup()
	return r
}

/*
Starts the lexical analysis and feeding tokens into the token channel
*/
func (lex *Lexer) Run() {
	for state := LexBegin; state != nil; {
		state = state(lex)
	}
	lex.Shutdown()
}

/*
Shuts down the token stream
*/
func (lex *Lexer) Shutdown() {
	close(lex.Tokens)
}

func BeginLexing(name, input string, state LexFn) *Lexer {
	l := &Lexer{
		Name:        name,
		Input:       []byte(input),
		State:       state,
		Tokens:      make(chan Token, 3),
		NestedStack: lane.NewStack(),
	}

	return l
}

/*
STATES
*/

func LexBegin(lex *Lexer) LexFn {
	// TODO: Make this EOF detection cleaner
	var next byte
	if lex.Pos >= len(lex.Input) {
		next = ' '
	} else {
		next = lex.Peek()
	}
	r, _ := utf8.DecodeRune([]byte{next})

	switch {
	case next == 'i':
		return LexIntegerStart
	case unicode.IsDigit(r):
		return LexStringStart
	case next == 'l':
		return LexListStart
	case next == 'd':
		return LexDictStart
	default:
		if lex.IsEOF() {
			if lex.NestedStack.Size() > 0 {
				lex.Errorf(LexErrUnclosedDelimeter)
			}
			lex.Emit(TOKEN_EOF)
		}
		if lex.NestedStack.Size() > 0 {
			if closeState := lex.NestedStack.Pop(); closeState != nil {
				return closeState.(func(*Lexer) LexFn)
			}
		}

		lex.Errorf(LexErrInvalidCharacter)

		return lex.Errorf("done")
	}

	panic("Shouldn't get here")
}

func LexStringStart(lex *Lexer) LexFn {
	for {
		lex.Next()
		if lex.IsEOF() {
			return lex.Errorf(LexErrUnexpectedEOF)
		}

		if strings.HasPrefix(string(lex.InputToEnd()), COLON) {
			n, err := strconv.ParseInt(string(lex.CurrentInput()), 10, 64)
			if err != nil || n < 0 {
				return lex.Errorf(LexErrInvalidStringLength)
			}
			lex.StringLength = int(n)
			lex.Emit(TOKEN_STRING_LENGTH)
			return LexStringValue
		}
	}
}

func LexStringValue(lex *Lexer) LexFn {
	lex.Next()
	lex.Emit(TOKEN_COLON)

	startPos := lex.Pos

	//	for i := 0; i < lex.StringLength; i++ {
	for lex.Pos < startPos+lex.StringLength {
		if lex.IsEOF() {
			return lex.Errorf(LexErrUnexpectedEOF)
		}
		lex.Next()
		//		next := lex.Next()
		//		if next == EOF {
		//			return lex.Errorf(LexErrUnexpectedEOF)
		//		}

		//		fmt.Println(next)
		//		lex.Pos++
	}

	lex.Emit(TOKEN_STRING_VALUE)

	return LexBegin
}

func LexIntegerStart(lex *Lexer) LexFn {
	lex.Pos += len(INTEGER_START)
	lex.Emit(TOKEN_INTEGER_START)
	return LexIntegerValue
}

func LexIntegerValue(lex *Lexer) LexFn {
	for {
		next := lex.Peek()
		r, _ := utf8.DecodeRune([]byte{next})
		if unicode.IsDigit(r) || next == '-' {
			lex.Pos++
		} else {
			return lex.Errorf(LexErrInvalidCharacter)
		}

		if strings.HasPrefix(string(lex.InputToEnd()), INTEGER_END) {
			lex.Emit(TOKEN_INTEGER_VALUE)
			return LexIntegerEnd
		}

		if lex.IsEOF() {
			return lex.Errorf(LexErrUnexpectedEOF)
		}
	}
}

func LexIntegerEnd(lex *Lexer) LexFn {
	lex.Pos += len(INTEGER_END)
	lex.Emit(TOKEN_INTEGER_END)
	return LexBegin
}

func LexDictStart(lex *Lexer) LexFn {
	lex.Pos += len(DICT_START)
	lex.Emit(TOKEN_DICT_START)
	return LexDictValue
}

func LexDictValue(lex *Lexer) LexFn {
	lex.NestedStack.Push(LexDictEnd)
	return LexBegin
}

func LexDictEnd(lex *Lexer) LexFn {
	lex.Pos += len(DICT_END)
	lex.Emit(TOKEN_DICT_END)
	return LexBegin
}

func LexListStart(lex *Lexer) LexFn {
	lex.Pos += len(LIST_START)
	lex.Emit(TOKEN_LIST_START)
	return LexListValue
}

func LexListValue(lex *Lexer) LexFn {
	lex.NestedStack.Push(LexListEnd)
	return LexBegin
}

func LexListEnd(lex *Lexer) LexFn {
	lex.Pos += len(LIST_END)
	lex.Emit(TOKEN_LIST_END)
	return LexBegin
}
