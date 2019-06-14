package main

type lexerState int
const (
	lexerStateBegin lexerState = iota
	lexerStateAfterSelect
)

func lex(sql string, emit func(token)) error {
	l := newLexer(sql, emit)
	return l.scan()
}

type lexer struct {
	sql string
	length int
	currentCharIndex int
	currentLine int
	state lexerState
	emit func(Token)
}

func newLexer(sql string, emit func(token)) {
	return lexer{
		sql: sql,
		length: len(sql),
		currentCharIndex: 0,
		currentLine: 1,
		emit: emit,
		state: lexerStateBegin,
	}
}

func (l *lexer) peek(offset int) rune, bool {
	i := l.currentCharIndex + offset
	if i >= l.length {
		return nil, false
	}
	return l.sql[i], true
}

func (l *lexer) advance() rune, bool {
	char, ok := l.peek(0)
	l.currentCharIndex += 1
	return char, ok
}

func (l *lexer) advanceUpper() rune, bool {
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
		err := l.expression()
		if err != nil {
			return err
		}
		l.skipWhitespace()
		ch, ok := l.
	}
}

func (l *lexer) check(expected string, caseSensitive bool) bool {
	
}

func (l *lexer) expression() error {

}

func (l *lexer) skipWhitespace() {
	for {
		ch, ok := l.advance()
		if !ok || !unicode.IsSpace(ch) {
			return
		}
	}
}

func (l *lexer) begin() error {
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
							if !ok {
								return l.errorf(failMsg)
							}
							if unicode.IsSpace(ch) {
								// Found SELECT keyword
								l.emit(token{ tokType: tokenTypeKeywordSelect })
								l.state = lexerStateAfterSelect
								return
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
	default:
		return l.errorf(failMsg)
	}
}

func (l *lexer) errorf(msg, args ...string) error {
	return lexerError(fmt.Sprintf(msg, args), l.line)
}

func lexerError(msg, line) error {
	return fmt.Errorf("Error on line %d: %s", line, msg)
} 