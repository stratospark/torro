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
	Input  string
	Tokens chan Token
	State  LexFn

	Start        int
	Pos          int
	Width        int
	StringLength int
	NestedStack  *lane.Stack
}

func (lex *Lexer) String() string {
	return fmt.Sprintf("Name: %s, Input: %s, Start: %d, Pos: %d, Width: %d",
		lex.Name,
		lex.Input,
		lex.Start,
		lex.Pos,
		lex.Width,
	)
}

type LexFn func(*Lexer) LexFn

/*
Backup to the beginning of the last read token.
*/
func (lex *Lexer) Backup() {
	lex.Pos -= lex.Width
}

/*
Returns a slice of the current input from the current lexer
start position to the current position.
*/
func (lex *Lexer) CurrentInput() string {
	return lex.Input[lex.Start:lex.Pos]
}

/*
Increment the position
*/
func (lex *Lexer) Inc() {
	lex.Pos++
	if lex.Pos >= utf8.RuneCountInString(lex.Input) {
		lex.Emit(TOKEN_EOF)
	}
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
	lex.Tokens <- Token{Type: tokenType, Value: lex.Input[lex.Start:lex.Pos]}
	lex.Start = lex.Pos
}

/*
Returns a token with error information.
*/
func (lex *Lexer) Errorf(format string, args ...interface{}) LexFn {
	lex.Tokens <- Token{
		Type:  TOKEN_ERROR,
		Value: fmt.Sprintf(format, args...),
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
func (lex *Lexer) InputToEnd() string {
	return lex.Input[lex.Pos:]
}

/*
Returns the true/false if the lexer is at the end of the input stream.
*/
func (lex *Lexer) IsEOF() bool {
	return lex.Pos >= len(lex.Input)
}

/*
Returns true/false if the next character is whitespace
*/
func (lex *Lexer) IsWhitespace() bool {
	ch, _ := utf8.DecodeRuneInString(lex.Input[lex.Pos:])
	return unicode.IsSpace(ch)
}

/*
Reads the next rune (character) from the input stream
and advances the lexer position.
*/
func (lex *Lexer) Next() rune {
	if lex.Pos >= utf8.RuneCountInString(lex.Input) {
		lex.Width = 0
		return EOF
	}

	result, width := utf8.DecodeRuneInString(lex.Input[lex.Pos:])

	lex.Width = width
	lex.Pos += lex.Width
	return result
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
func (lex *Lexer) Peek() rune {
	rune := lex.Next()
	lex.Backup()
	return rune
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

/*
Skips whitespace until we get something meaningful.
*/
func (lex *Lexer) SkipWhitespace() {
	for {
		ch := lex.Next()

		if !unicode.IsSpace(ch) {
			lex.Dec()
			break
		}

		if ch == EOF {
			lex.Emit(TOKEN_EOF)
			break
		}
	}
}

func BeginLexing(name, input string, state LexFn) *Lexer {
	l := &Lexer{
		Name:        name,
		Input:       input,
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
	//	lex.SkipWhitespace()
	//	if strings.HasPrefix(lex.InputToEnd(), DICT_START) {
	//		return LexDictStart
	//	} else {
	//		lex.Inc()
	//		return lex.Errorf("done")
	//	}

	//	lex.NestedStack = append(lex.NestedStack, LexListEnd)
	next := lex.Peek()
	switch {
	case next == 'i':
		return LexIntegerStart
	case unicode.IsDigit(next):
		return LexStringStart
	case next == 'l':
		return LexListStart
	default:
		if lex.NestedStack.Size() > 0 {
			if closeState := lex.NestedStack.Pop(); closeState != nil {
				return closeState.(func(*Lexer) LexFn)
			}
		}
		if lex.IsEOF() {
			lex.Emit(TOKEN_EOF)
		}
		return lex.Errorf("done")
	}

	panic("Shouldn't get here")
}

func LexStringStart(lex *Lexer) LexFn {
	for {
		lex.Inc()

		if strings.HasPrefix(lex.InputToEnd(), COLON) {
			n, err := strconv.ParseInt(lex.CurrentInput(), 10, 64)
			if err != nil {
				return lex.Errorf("Invalid string length")
			}
			lex.StringLength = int(n)
			lex.Emit(TOKEN_STRING_LENGTH)
			return LexStringValue
		}

		if lex.IsEOF() {
			return lex.Errorf("Unexpected EOF")
		}
	}
}

func LexStringValue(lex *Lexer) LexFn {
	lex.Pos++
	lex.Emit(TOKEN_COLON)

	for i := 0; i < lex.StringLength; i++ {
		if lex.IsEOF() {
			return lex.Errorf("Unexpected EOF")
		}
		lex.Pos++
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
		lex.Inc()

		if strings.HasPrefix(lex.InputToEnd(), INTEGER_END) {
			lex.Emit(TOKEN_INTEGER_VALUE)
			return LexIntegerEnd
		}

		if lex.IsEOF() {
			return lex.Errorf("Unexpected EOF")
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
	lex.Emit(TOKEN_INTEGER_START)
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
