package lib

type tokenReader interface {
	Next() (tok token, done bool, err error)
	Peek() (tok token, done bool, err error)
}
