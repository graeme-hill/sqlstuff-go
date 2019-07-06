package lib

type tokenType int

const (
	tokenTypeWord tokenType = iota
	tokenTypeNumber
	tokenTypeLParen
	tokenTypeRParen
	tokenTypeString
	tokenTypeDot
	tokenTypeComma
	tokenTypeSemicolon
	tokenTypePlus
	tokenTypeMinus
	tokenTypeSlash
	tokenTypeAsterisk
	tokenTypeLess
	tokenTypeLessOrEqual
	tokenTypeGreater
	tokenTypeGreaterOrEqual
	tokenTypeEqual
	tokenTypeNotEqual
)

type token struct {
	tokType  tokenType
	value    []rune
	location charLocation
}
