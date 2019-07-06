package lib

import (
	"errors"
	"time"
)

const TOKEN_BUF_SIZE = 100

var TokenReadTimeout = 1 * time.Second

type peekResult struct {
	tok  token
	done bool
	err  error
}

type tokenBuffer struct {
	tokChan      chan token
	doneChan     chan struct{}
	peeked       *peekResult
	doneReceived bool
}

func newTokenBuffer() *tokenBuffer {
	return &tokenBuffer{
		tokChan:      make(chan token, TOKEN_BUF_SIZE),
		doneChan:     make(chan struct{}, 1),
		peeked:       nil,
		doneReceived: false,
	}
}

func (tb *tokenBuffer) Next() (tok token, done bool, err error) {
	if tb.peeked != nil {
		res := tb.peeked
		tb.peeked = nil
		return res.tok, res.done, res.err
	}

	timeout := TokenReadTimeout
	if tb.doneReceived {
		timeout = 0
	}

	select {
	case tok := <-tb.tokChan:
		return tok, false, nil
	case <-tb.doneChan:
		tb.doneReceived = true
		return tb.Next()
	case <-time.After(timeout):
		if tb.doneReceived {
			return token{}, true, nil
		}
		return token{}, false, errors.New("timed out waiting for next token")
	}
}

func (tb *tokenBuffer) Peek() (token, bool, error) {
	if tb.peeked != nil {
		return tb.peeked.tok, tb.peeked.done, tb.peeked.err
	}
	tok, done, err := tb.Next()
	tb.peeked = &peekResult{tok: tok, done: done, err: err}
	return tok, done, err
}

func (tb *tokenBuffer) Write(tok token) {
	tb.tokChan <- tok
}

func (tb *tokenBuffer) Done() {
	tb.doneChan <- struct{}{}
}
