package main

type tokenType int

const (
	tokenTypeKeywordSelect tokenType = iota
	tokenTypeLiteral
	tokenTypeLParen
	tokenTypeRParen
	tokenTypeLDoubleQuote
	tokenTypeRDoubleQuote
	tokenTypeLSingleQuote
	tokenTypeRSingleQuote
	tokenTypeDot
	tokenTypeIdentifier
	tokenTypeComma
	tokenTypeSemicolon
)

type token struct {
	tokType tokenType
	value   string
}
