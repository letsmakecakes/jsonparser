package lexer

import (
	"fmt"
	"strings"
)

type TokenType string

const (
	TokenLeftBrace  TokenType = "{"
	TokenRightBrace TokenType = "}"
	TokenColon      TokenType = ":"
	TokenComma      TokenType = ","
	TokenString     TokenType = "STRING"
	TokenNumber     TokenType = "NUMBER"
	TokenTrue       TokenType = "TRUE"
	TokenFalse      TokenType = "FALSE"
	TokenNull       TokenType = "NULL"
	TokenEOF        TokenType = "EOF"
)

type Token struct {
	Type    TokenType
	Literal string
}

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for {
		l.skipWhiteSpace()
		var tok Token

		switch l.ch {
		case '{':
			tok = Token{Type: TokenLeftBrace, Literal: "{"}
		case '}':
			tok = Token{Type: TokenRightBrace, Literal: "}"}
		case ':':
			tok = Token{Type: TokenColon, Literal: ":"}
		case ',':
			tok = Token{Type: TokenComma, Literal: ","}
		case '"':
			str, err := l.readString()
			if err != nil {
				return nil, err
			}
			tok = Token{Type: TokenString, Literal: str}
			tokens = append(tokens, tok)
			continue
		case 0:
			tok = Token{Type: TokenEOF, Literal: ""}
			tokens = append(tokens, tok)
			return tokens, nil
		default:
			if isLetter(l.ch) {
				literal := l.readLiteral()
				switch literal {
				case "true":
					tok = Token{Type: TokenTrue, Literal: literal}
				case "false":
					tok = Token{Type: TokenFalse, Literal: literal}
				case "null":
					tok = Token{Type: TokenNull, Literal: literal}
				default:
					return nil, NewUnexpectedCharacterError(l.ch)
				}
			} else if isDigit(l.ch) || l.ch == '-' {
				num := l.readNumber()
				tok = Token{Type: TokenNumber, Literal: num}
			} else {
				return nil, NewUnexpectedCharacterError(l.ch)
			}
		}

		tokens = append(tokens, tok)
		l.readChar()
	}
}

func (l *Lexer) skipWhiteSpace() {
	for isWhiteSpace(l.ch) {
		l.readChar()
	}
}

func isWhiteSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
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
			l.readChar() // Move past the backlash
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
				unicodeChar, err := l.readUnicode()
				if err != nil {
					return "", err
				}
				strBuilder.WriteByte(byte(unicodeChar))
			default:
				return "", NewUnexpectedCharacterError(l.ch)
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
		l.readChar()
		if !isHexDigit(l.ch) {
			return 0, fmt.Errorf("invalid Unicode escpae sequence")
		}
		hex += string(l.ch)
	}

	var unicodeValue rune
	_, err := fmt.Sscanf(hex, "%04x", &unicodeValue)
	if err != nil {
		return 0, fmt.Errorf("invalid Unicode escape sequence")
	}

	return unicodeValue, nil
}

func isHexDigit(ch byte) bool {
	return ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readLiteral() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}

	// Return the literal by slicing the input string from the starting position to the current position
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position

	// Check for a leading minus sign(-).
	if l.ch == '-' {
		l.readChar() // Move past the minus sign.
	}

	// Read through all digits before the decimal point.
	for isDigit(l.ch) {
		l.readChar()
	}

	// If there is a decimal point, read the fractional part.
	if l.ch == '.' {
		l.readChar() // Move past the decimal point.

		// Read through digits after the decimal point
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Return the full number as a string.
	return l.input[position:l.position]
}
