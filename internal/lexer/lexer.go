package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// TokenType defines the type of lexical tokens
type TokenType string

// Token types
const (
	TokenLeftBrace    TokenType = "{"
	TokenRightBrace   TokenType = "}"
	TokenLeftBracket  TokenType = "["
	TokenRightBracket TokenType = "]"
	TokenColon        TokenType = ":"
	TokenComma        TokenType = ","
	TokenString       TokenType = "STRING"
	TokenNumber       TokenType = "NUMBER"
	TokenTrue         TokenType = "TRUE"
	TokenFalse        TokenType = "FALSE"
	TokenNull         TokenType = "NULL"
	TokenEOF          TokenType = "EOF"
)

// Token represents a lexical token with type and literal value
type Token struct {
	Type    TokenType
	Literal string
	Line    int // Line number in input
	Column  int // Column number in input
}

// Lexer represents a lexical scanner
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           rune // current char under examination
	line         int  // current line number
	column       int  // current column number
}

// NewLexer initializes a new Lexer with the given input
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar reads the next character and updates positions
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		r, size := utf8.DecodeLastRuneInString(l.input[l.readPosition:])
		l.ch = r
		l.readPosition += size
		l.position = l.readPosition
		l.column++
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
	}
	l.position = l.readPosition
	l.readPosition++
}

// peekChar peeks ahead to the next character without advancing the lexer
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeLastRuneInString(l.input[l.readPosition:])
	return r
}

// skipWhiteSpace skips over any whitespace characters
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}
}

// Tokenize converts the input string into a slice of Tokens
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for l.ch != 0 {
		l.skipWhitespace() // Skip any whitespace characters

		var tok Token
		tok.Line = l.line
		tok.Column = l.column

		switch l.ch {
		case '{':
			tok = Token{Type: TokenLeftBrace, Literal: "{", Line: l.line, Column: l.column}
		case '}':
			tok = Token{Type: TokenRightBrace, Literal: "}", Line: l.line, Column: l.column}
		case '[':
			tok = Token{Type: TokenLeftBracket, Literal: "[", Line: l.line, Column: l.column}
		case ']':
			tok = Token{Type: TokenRightBracket, Literal: "]", Line: l.line, Column: l.column}
		case ':':
			tok = Token{Type: TokenColon, Literal: ":", Line: l.line, Column: l.column}
		case ',':
			tok = Token{Type: TokenComma, Literal: ",", Line: l.line, Column: l.column} // Create token for comma
		case '"':
			str, err := l.readString()
			if err != nil {
				return nil, err
			}
			tok = Token{Type: TokenString, Literal: str, Line: l.line, Column: l.column}
			tokens = append(tokens, tok)
			continue
		case 't':
			if l.peekKeyWord("true") {
				tok = Token{Type: TokenTrue, Literal: "true", Line: l.line, Column: l.column}
				l.advanceBy(len("true"))
			} else {
				return nil, fmt.Errorf("Lexer error at line %d, column %d: invalid token starting with 't'", l.line, l.column)
			}
		case 'f':
			if l.peekKeyWord("false") {
				tok = Token{Type: TokenFalse, Literal: "false", Line: l.line, Column: l.column}
				l.advanceBy(len("false"))
			} else {
				return nil, fmt.Errorf("Lexer error at line %d, column %d: invalid token starting with 'f'", l.line, l.column)
			}
		case 'n':
			if l.peekKeyWord("null") {
				tok = Token{Type: TokenNull, Literal: "null", Line: l.line, Column: l.column}
				l.advanceBy(len("null"))
			} else {
				return nil, fmt.Errorf("Lexer error at line %d, column %d: invalid token starting with 'n'", l.line, l.column)
			}
		default:
			if l.isStartOfNumber(l.ch) {
				num, err := l.readNumber()
				if err != nil {
					return nil, fmt.Errorf("Lexer error at line %d, column %d: %v", l.line, l.column, err)
				}
				tok = Token{Type: TokenNumber, Literal: num, Line: l.line, Column: l.column}
			} else {
				return nil, fmt.Errorf("Lexer error at line %d, column %d: unexpected character: %c", l.ch, l.line, l.column)
			}
		}

		tokens = append(tokens, tok) // Append the created token to the tokens slice
		l.readChar()                 // Move to the next character for the next iteration
	}

	// Append EOF token
	tokens = append(tokens, Token{Type: TokenEOF, Literal: "", Line: l.line, Column: l.column})

	return tokens, nil
}

// peekKeyword checks if the upcoming characters match the expected keyword
func (l *Lexer) peekKeyWord(expected string) bool {
	end := l.readPosition + len(expected)
	if end > len(l.input) {
		return false
	}

	return l.input[l.readPosition:end] == expected
}

// advanceBy advances the lexer by n characters
func (l *Lexer) advanceBy(n int) {
	for i := 0; i < n; i++ {
		l.readChar()
	}
}

// isStartOfNumber checks if the rune can start a number
func (l *Lexer) isStartOfNumber(r rune) bool {
	return r == '-' || unicode.IsDigit(r)
}

// readNumber reads a number token from the input, including exponents
func (l *Lexer) readNumber() (string, error) {
	startPos := l.position
	startLine := l.line
	startColumn := l.column

	if err := l.consumeMinus(); err != nil {
		return "", err
	}

	// Handle optional minus sign
	if l.ch == '-' {
		numBuilder.WriteRune(l.ch)
		l.readChar()
	}

	if !unicode.IsDigit(l.ch) {
		return "", fmt.Errorf("invalid number format: expected digit after '-'")
	}

	if l.ch == '.' {
		numBuilder.WriteRune(l.ch)
		l.readChar()

		if !unicode.IsDigit(l.ch) {
			return "", fmt.Errorf("invalid number format: expected digit after '.'")
		}

		for unicode.IsDigit(l.ch) {
			numBuilder.WriteRune(l.ch)
			l.readChar()
		}
	}

	if l.ch == 'e' || l.ch == 'E' {
		numBuilder.WriteRune(l.ch)
		l.readChar()

		if l.ch == '+' || l.ch == '-' {
			numBuilder.WriteRune(l.ch)
			l.readChar()
		}

		if !unicode.IsDigit(l.ch) {
			return "", fmt.Errorf("invalid number format: expected digit after exponent indicater")
		}

		for unicode.IsDigit(l.ch) {
			numBuilder.WriteRune(l.ch)
			l.readChar()
		}
	}

	return numBuilder.String(), nil
}

// consumeMinus handles the optional minus sign
func (l *Lexer) consumeMinus() error {
	if l.ch == '-' {
		l.readChar()
	}
	return nil
}

func isHighSurrogate(r rune) bool {
	return r >= 0xD800 && r <= 0xDBFF
}

func isLowSurrogate(r rune) bool {
	return r >= 0xDC00 && r <= 0xDFFF
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readString() (string, error) {
	var strBuilder strings.Builder

	// Read the opening quote
	l.readChar()

	for {
		switch l.ch {
		case '"': // Closing quote found, return the string
			l.readChar() // Move past the closing quote
			return strBuilder.String(), nil
		case '\\': // Handle escape sequences
			l.readChar() // Move past the backslash
			switch l.ch {
			case '"':
				strBuilder.WriteByte('"')
			case '\\':
				strBuilder.WriteByte('\\')
			case '/':
				strBuilder.WriteByte('/')
			case 'b':
				strBuilder.WriteByte('\b')
			case 'f':
				strBuilder.WriteByte('\f')
			case 'n':
				strBuilder.WriteByte('\n')
			case 'r':
				strBuilder.WriteByte('\r')
			case 't':
				strBuilder.WriteByte('\t')
			case 'u':
				// Handle Unicode escape sequences (e.g., \uXXXX)
				runeValue, err := l.readUnicode()
				if err != nil {
					return "", err
				}

				// Check if the rune is a high surrogate
				if isHighSurrogate(runeValue) {
					// Expecting a low surrogate next
					if l.ch != '\\' {
						return "", fmt.Errorf("expected '\\' after high surrogate, got '%c'", l.ch)
					}
					l.readChar() // Skip the backslash
					if l.ch != 'u' {
						return "", fmt.Errorf("expected 'u' after '\\', got '%c'", l.ch)
					}
					l.readChar() // Skip the 'u'

					lowSurrogate, err := l.readUnicode()
					if err != nil {
						return "", err
					}

					if !isLowSurrogate(lowSurrogate) {
						return "", fmt.Errorf("invalid low surrogate: \\u%04X", lowSurrogate)
					}

					// Combine the surrogate pair into a single rune
					combinedRune := utf16.DecodeRune(runeValue, lowSurrogate)
					if combinedRune == utf8.RuneError {
						return "", fmt.Errorf("invalid surrogate pair: \\u%04X\\u%04X", runeValue, lowSurrogate)
					}

					strBuilder.WriteRune(combinedRune)
				} else {
					// Regular Unicode character
					strBuilder.WriteRune(runeValue)
				}
			default:
				return "", fmt.Errorf("unexpected character: '%c' in string escape", l.ch)
			}
		case 0: // End of input, but no closing quote found
			return "", fmt.Errorf("unterminated string")
		default:
			strBuilder.WriteByte(l.ch)
		}

		l.readChar() // Read the next character
	}
}

func (l *Lexer) readUnicode() (rune, error) {
	var hex string
	for i := 0; i < 4; i++ {
		if !isHexDigit(l.ch) {
			return 0, fmt.Errorf("invalid Unicode escape sequence: \\u%s", hex)
		}
		hex += string(l.ch)
		l.readChar()
	}

	var unicodeValue uint32
	_, err := fmt.Sscanf(hex, "%04x", &unicodeValue)
	if err != nil {
		return 0, fmt.Errorf("invalid Unicode escape sequence: \\u%s", hex)
	}

	return rune(unicodeValue), nil
}

func isHexDigit(ch byte) bool {
	return ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}
