package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNext(t *testing.T) {
	buf := newTokenBuffer()

	buf.Write(token{tokType: tokenTypeWord, value: []rune("hello")})

	tok, done, err := buf.Next()
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, tokenTypeWord, tok.tokType)
	require.Equal(t, "hello", string(tok.value))
}

func TestNextDone(t *testing.T) {
	buf := newTokenBuffer()

	buf.Write(token{tokType: tokenTypeWord, value: []rune("hello")})
	buf.Done()

	tok, done, err := buf.Next()
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, tokenTypeWord, tok.tokType)
	require.Equal(t, "hello", string(tok.value))

	_, done, err = buf.Next()
	require.NoError(t, err)
	require.True(t, done)
}

func TestNextDoneMulti(t *testing.T) {
	buf := newTokenBuffer()

	buf.Write(token{tokType: tokenTypeWord, value: []rune("hello")})
	buf.Done()

	tok, done, err := buf.Next()
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, tokenTypeWord, tok.tokType)
	require.Equal(t, "hello", string(tok.value))

	_, done, err = buf.Next()
	require.NoError(t, err)
	require.True(t, done)

	_, done, err = buf.Next()
	require.NoError(t, err)
	require.True(t, done)

	_, done, err = buf.Next()
	require.NoError(t, err)
	require.True(t, done)
}

func TestNextTimeout(t *testing.T) {
	oldTimeout := TokenReadTimeout
	TokenReadTimeout = 1 * time.Microsecond
	defer func() {
		TokenReadTimeout = oldTimeout
	}()

	buf := newTokenBuffer()
	_, done, err := buf.Next()
	require.Error(t, err)
	require.False(t, done)
}

func TestPeek(t *testing.T) {
	buf := newTokenBuffer()

	buf.Write(token{tokType: tokenTypeWord, value: []rune("hello")})
	buf.Done()

	tok, done, err := buf.Peek()
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, tokenTypeWord, tok.tokType)
	require.Equal(t, "hello", string(tok.value))

	tok, done, err = buf.Next()
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, tokenTypeWord, tok.tokType)
	require.Equal(t, "hello", string(tok.value))

	_, done, err = buf.Next()
	require.NoError(t, err)
	require.True(t, done)
}
