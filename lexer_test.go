package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A test helper function that just aggregates tokens into a slice for easier
// assertions.
func getTokens(sql string) ([]token, error) {
	tokens := []token{}
	err := lex(sql, func(t token) {
		tokens = append(tokens, t)
	})
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func requireTok(t *testing.T, actual token, typ tokenType, value string, line int, col int) {
	require.Equal(t, actual.tokType, typ, "token type")
	require.Equal(t, string(actual.value), value, "token value")
	require.Equal(t, actual.location.line, line, "token line")
	require.Equal(t, actual.location.col, col, "token col")
}

func TestLexerOneWord(t *testing.T) {
	tokens, err := getTokens("select")
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	requireTok(t, tokens[0], tokenTypeWord, "select", 1, 1)
}

func TestLexerTwoWords(t *testing.T) {
	tokens, err := getTokens("select foo")
	require.NoError(t, err)
	require.Len(t, tokens, 2)
	requireTok(t, tokens[0], tokenTypeWord, "select", 1, 1)
	requireTok(t, tokens[1], tokenTypeWord, "foo", 1, 8)
}

func TestLexerWordsMultiLine(t *testing.T) {
	tokens, err := getTokens(`
select
	foo
from bar`)
	require.NoError(t, err)
	require.Len(t, tokens, 4)
	requireTok(t, tokens[0], tokenTypeWord, "select", 2, 1)
	requireTok(t, tokens[1], tokenTypeWord, "foo", 3, 2)
	requireTok(t, tokens[2], tokenTypeWord, "from", 4, 1)
	requireTok(t, tokens[3], tokenTypeWord, "bar", 4, 6)
}

func TestLexerRealSelect(t *testing.T) {
	tokens, err := getTokens(`select foo, bar+1 from things;`)
	require.NoError(t, err)
	require.Len(t, tokens, 9)
	requireTok(t, tokens[0], tokenTypeWord, "select", 1, 1)
	requireTok(t, tokens[1], tokenTypeWord, "foo", 1, 8)
	requireTok(t, tokens[2], tokenTypeComma, "", 1, 11)
	requireTok(t, tokens[3], tokenTypeWord, "bar", 1, 13)
	requireTok(t, tokens[4], tokenTypePlus, "", 1, 16)
	requireTok(t, tokens[5], tokenTypeWord, "1", 1, 17)
	requireTok(t, tokens[6], tokenTypeWord, "from", 1, 19)
	requireTok(t, tokens[7], tokenTypeWord, "things", 1, 24)
	requireTok(t, tokens[8], tokenTypeSemicolon, "", 1, 30)
}

func TestLexerString(t *testing.T) {
	tokens, err := getTokens("'foo  bar'")
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	requireTok(t, tokens[0], tokenTypeString, "foo  bar", 1, 1)
}

func TestLexerStringEscapedWithBackslash(t *testing.T) {
	tokens, err := getTokens("'foo\\'s'")
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	requireTok(t, tokens[0], tokenTypeString, "foo\\'s", 1, 1)
}

func TestLexerStringEscapedWithDoubleSingleQuote(t *testing.T) {
	tokens, err := getTokens("'foo''s'")
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	requireTok(t, tokens[0], tokenTypeString, "foo''s", 1, 1)
}

func TestLexerStringInvalid(t *testing.T) {
	_, err := getTokens("'foo")
	require.Error(t, err)
}
