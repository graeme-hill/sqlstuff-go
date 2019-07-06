package main

import "github.com/graeme-hill/sqlstuff-go/lib"

func main() {
	err := lib.Generate("./test/migrations", "./test/queries", "./test/store/queries.go", "store")
	if err != nil {
		panic(err)
	}
}