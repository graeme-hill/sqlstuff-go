package main

import (
	"fmt"
	"unicode"
)

type lexerState int

const (
	lexerStateNormal lexerState = iota
	lexerStateQuotedIdentifier
)

func lex(sql string, emit func(token)) error {
	l := newLexer(sql, emit)
	return l.scan()
}

type lexer struct {
	sql              string
	length           int
	currentCharIndex int
	currentLine      int
	state            lexerState
	emit             func(Token)
}

func newLexer(sql string, emit func(token)) {
	return lexer{
		sql:              sql,
		length:           len(sql),
		currentCharIndex: 0,
		currentLine:      1,
		emit:             emit,
		state:            lexerStateBegin,
	}
}

func (l *lexer) peek(offset int) (rune, bool) {
	i := l.currentCharIndex + offset
	if i >= l.length {
		return nil, false
	}
	return l.sql[i], true
}

func (l *lexer) advance() (rune, bool) {
	char, ok := l.peek(0)
	l.currentCharIndex++
	return char, ok
}

func (l *lexer) advanceUpper() (rune, bool) {
	char, ok := l.advance()
	if !ok {
		return char, false
	}
	return unicode.ToUpper(char), true
}

func (l *lexer) scan() error {
	for l.currentCharIndex < len(l.sql) {
		err := l.next()
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *lexer) next() error {
	switch lexerState {
	case lexerStateBegin:
		l.begin()
	case lexerStateAfterSelect:
		l.fields()
	default:
		return l.errorf("sql lexer encountered an unknown state")
	}
}

func (l *lexer) fields() error {
	for {
		// get next expression in select list
		err := l.expression()
		if err != nil {
			return err
		}
		l.skipWhitespace()

		// emit alias tokens if there is an alias
		if l.check("AS", true) {
			l.emit(token{tokType: tokenTypeKeywordAs})
			err = l.columnName()
			if err != nil {
				return err
			}
			l.skipWhitespace()
		}

		// keep looping iff there is a comma
		next, ok := l.peek(1)
		if !ok || next != ',' {
			break
		}
		l.currentCharIndex++
	}

	return nil
}

// Examples: name or user.name
func (l *lexer) columnName() error {
	// Get the required first part
	first, err := l.symbol()
	if err != nil {
		return err
	}

	// If followed by a dot then get the optional second part
	ch, ok := l.advance()
	if ok && ch == '.' {
		second, err := l.symbol
		if err != nil {
			return err
		}
	}
}

func (l *lexer) symbol() string, error {
	start := l.currentCharIndex
	for {
		ch, ok := 
	}
}

func (l *lexer) check(expected string, capitalized bool) bool {
	for i := 0; i < len(expected); i++ {
		expectedChar := expected[i]
		actualChar, ok := l.peek(i)
		if !ok {
			return false
		}
		if capitalized {
			actualChar = unicode.ToUpper(actualChar)
		}
		if expectedChar != actualChar {
			return false
		}
	}

	// it matched so advance and return true
	l.currentCharIndex += len(expected)
	return true
}

func (l *lexer) expression() error {

}

func (l *lexer) skipWhitespace() {
	for {
		ch, ok := l.advance()
		if !ok {
			// there are no more chars so stop advancing
			return
		}

		if ch == '\n' {
			// started a new line so increment line index
			l.currentLine++
		}

		if !unicode.IsSpace(ch) {
			// not whitespace so stop advancing
			return
		}
	}
}

func (l *lexer) normal() error {
	const failMsg = "expected a new statement"
	ch, ok := l.advanceUpper()
	if !ok {
		return l.errorf(failMsg)
	}

	switch ch {
	// Look for SELECT
	case 'S':
		ch, ok = l.advanceUpper()
		if !ok {
			return l.errorf(failMsg)
		}
		switch ch {
		case 'E':
			ch, ok = l.advanceUpper()
			if !ok {
				return l.errorf(failMsg)
			}
			switch ch {
			case 'L':
				ch, ok = l.advanceUpper()
				if !ok {
					return l.errorf(failMsg)
				}
				switch ch {
				case 'E':
					ch, ok = l.advanceUpper()
					if !ok {
						return l.errorf(failMsg)
					}
					switch ch {
					case 'C':
						ch, ok = l.advanceUpper()
						if !ok {
							return l.errorf(failMsg)
						}
						switch ch {
						case 'T':
							ch, ok = l.advanceUpper()
							if !ok || unicode.IsSpace(ch) {
								l.emit(token{tokType: tokenTypeKeywordSelect})
							}
						default:
							return l.errorf(failMsg)
						}
					default:
						return l.errorf(failMsg)
					}
				default:
					return l.errorf(failMsg)
				}
			default:
				return l.errorf(failMsg)
			}
		default:
			return l.errorf(failMsg)
		}
	case '.':
		l.emit(token{tokType: tokenTypeDot})
	case ',':
		l.emit(token{tokType: tokenTypeComma})
	case '(':
		l.emit(token{tokType: tokenTypeLParen})
	case ')':
		l.emit(token{tokType: tokenTypeRParen})
	case ' ':
	case '\t':
	case '\r':
	case '\n':
		l.currentLine++
	case '\'':
		l.stringLiteral()
	case '"':
		l.quotedIentifier()
	default:
		return l.errorf(failMsg)
	}
}

func (l *lexer) stringLiteral() error {

}

func (l *lexer) errorf(msg string, args ...string) error {
	return lexerError(fmt.Sprintf(msg, args), l.line)
}

func lexerError(msg, line) error {
	return fmt.Errorf("Error on line %d: %s", line, msg)
}
