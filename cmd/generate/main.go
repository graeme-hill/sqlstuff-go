package main

import "github.com/graeme-hill/sqlstuff-go/lib"

func main() {
	err := lib.Generate("./test/migrations", "./test/queries", "./test/generated/queries.go", "test")
	if err != nil {
		panic(err)
	}
}