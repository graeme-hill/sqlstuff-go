package lib

import (
	"fmt"
)

type charInfo struct {
	ch       rune
	location charLocation
}

func lex(sql string, emit func(token)) error {
	l := newLexer(sql, emit)
	return l.scan()
}

type lexer struct {
	sql              []rune
	length           int
	currentCharIndex int
	currentLocation  charLocation
	tokenStartIndex  int
	tokenLocation    charLocation
	emitCallback     func(token)
}

func newLexer(sql string, emit func(token)) *lexer {
	return &lexer{
		sql:              []rune(sql),
		length:           len(sql),
		currentCharIndex: 0,
		currentLocation:  charLocation{line: 1, col: 1},
		tokenStartIndex:  0,
		tokenLocation:    charLocation{line: 1, col: 1},
		emitCallback:     emit,
	}
}

func (l *lexer) emit(tok token) {
	l.endWord()
	l.emitCallback(tok)
	l.resetToken()
}

func (l *lexer) peek(offset int) (charInfo, bool) {
	i := l.currentCharIndex + offset
	if i >= l.length {
		return charInfo{}, false
	}
	return charInfo{ch: l.sql[i], location: l.currentLocation}, true
}

func (l *lexer) advance() (charInfo, bool) {
	info, ok := l.peek(0)
	l.currentCharIndex++
	if info.ch == '\n' {
		l.currentLocation.line++
		l.currentLocation.col = 1
	} else {
		l.currentLocation.col++
	}
	return info, ok
}

func (l *lexer) scan() error {
	for {
		more, err := l.next()
		if err != nil {
			return err
		}
		if !more {
			break
		}
	}
	return nil
}

func (l *lexer) next() (bool, error) {
	chInfo, ok := l.advance()
	if !ok {
		l.endWord()
		return false, nil
	}
	ch := chInfo.ch
	var err error = nil
	more := true

	switch ch {
	case '=':
		l.emit(token{tokType: tokenTypeEqual, location: chInfo.location})
	case '<':
		{
			ahead, ok := l.peek(1)
			if ok && ahead.ch == '=' {
				_, _ = l.advance()
				l.emit(token{tokType: tokenTypeLessOrEqual, location: chInfo.location})
			} else if ok && ahead.ch == '>' {
				_, _ = l.advance()
				l.emit(token{tokType: tokenTypeNotEqual, location: chInfo.location})
			} else {
				l.emit(token{tokType: tokenTypeLess, location: chInfo.location})
			}
		}
	case '>':
		{
			ahead, ok := l.peek(1)
			if ok && ahead.ch == '=' {
				_, _ = l.advance()
				l.emit(token{tokType: tokenTypeGreaterOrEqual, location: chInfo.location})
			} else {
				l.emit(token{tokType: tokenTypeGreater, location: chInfo.location})
			}
		}
	case '.':
		{
			ahead, ok := l.peek(1)
			if ok && isDigit(ahead.ch) && l.isFirstCharOfToken() {
				more, err = l.scanNumber(ch)
			} else {
				l.emit(token{tokType: tokenTypeDot, location: chInfo.location})
			}
		}
	case ',':
		l.emit(token{tokType: tokenTypeComma, location: chInfo.location})
	case ';':
		l.emit(token{tokType: tokenTypeSemicolon, location: chInfo.location})
	case '(':
		l.emit(token{tokType: tokenTypeLParen, location: chInfo.location})
	case ')':
		l.emit(token{tokType: tokenTypeRParen, location: chInfo.location})
	case '+':
		l.emit(token{tokType: tokenTypePlus, location: chInfo.location})
	case '-':
		l.emit(token{tokType: tokenTypeMinus, location: chInfo.location})
	case '/':
		l.emit(token{tokType: tokenTypeSlash, location: chInfo.location})
	case '*':
		l.emit(token{tokType: tokenTypeAsterisk, location: chInfo.location})
	case ' ':
		fallthrough
	case '\t':
		fallthrough
	case '\r':
		fallthrough
	case '\n':
		l.eatWhitespace()
	case '\'':
		more, err = l.wrapped('\'', tokenTypeString)
	case '"':
		more, err = l.wrapped('"', tokenTypeWord)
	default:
		// keep going with this word unless starting a number, then do some
		// different stuff because that's different
		if isDigit(ch) {
			more, err = l.scanNumber(ch)
		}
	}

	return more, err
}

func (l *lexer) isFirstCharOfToken() bool {
	return l.currentCharIndex-1 == l.tokenStartIndex
}

func (l *lexer) scanNumber(first rune) (bool, error) {
	hasDecimal := first == '.'
	end := l.currentCharIndex

	for {
		next, ok := l.peek(0)
		if !ok {
			_, _ = l.advance()
			end = l.currentCharIndex
			break
		}

		isDecimal := next.ch == '.'
		if isDecimal && hasDecimal {
			return false, fmt.Errorf("Cannot have multiple decimals in one number")
		}
		hasDecimal = hasDecimal || isDecimal

		if !isDecimal && !isDigit(next.ch) {
			end++
			break
		}

		_, _ = l.advance()
		end = l.currentCharIndex
	}

	substr := l.sql[l.tokenStartIndex : end-1]
	l.emitCallback(token{tokType: tokenTypeNumber, value: substr, location: l.tokenLocation})
	l.resetToken()
	return true, nil
}

func isDigit(ch rune) bool {
	return ch == '0' ||
		ch == '1' ||
		ch == '2' ||
		ch == '3' ||
		ch == '4' ||
		ch == '5' ||
		ch == '6' ||
		ch == '7' ||
		ch == '8' ||
		ch == '9'
}

func (l *lexer) eatWhitespace() {
	l.endWord()
}

func (l *lexer) endWord() {
	if !l.isFirstCharOfToken() {
		substr := l.sql[l.tokenStartIndex : l.currentCharIndex-1]
		l.emitCallback(token{tokType: tokenTypeWord, value: substr, location: l.tokenLocation})
	}
	l.resetToken()
}

func (l *lexer) resetToken() {
	l.tokenLocation = l.currentLocation
	l.tokenStartIndex = l.currentCharIndex
}

func (l *lexer) wrapped(terminator rune, tokType tokenType) (bool, error) {
	l.endWord()

	escaping := false
	start := l.currentCharIndex
	startLoc := l.currentLocation
	startLoc.col--

	for {
		escapingThisIteration := false
		current, ok := l.advance()
		if !ok {
			return false, l.errorf("looking for %s", string(terminator))
		}

		if !escaping {
			if current.ch == '\\' {
				escapingThisIteration = true
			} else if current.ch == terminator {
				next, hasNext := l.peek(0)
				if hasNext && next.ch == terminator {
					escapingThisIteration = true
				} else {
					break
				}
			}
		}

		escaping = escapingThisIteration
	}

	substr := l.sql[start : l.currentCharIndex-1]
	l.emitCallback(token{tokType: tokType, value: substr, location: startLoc})
	l.resetToken()
	return true, nil
}

func (l *lexer) errorf(msg string, args ...string) error {
	formatted := fmt.Sprintf(msg, args)
	return fmt.Errorf("Error at line %d:%d: %s", l.currentLocation.line, l.currentLocation.col, formatted)
}
