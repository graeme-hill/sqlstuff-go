package main

import "github.com/graeme-hill/sqlstuff-go/lib"

func main() {
	err := lib.Generate("./test/basic/migrations", "./test/basic/queries", "./test/basic/store/queries.go", "store")
	if err != nil {
		panic(err)
	}

	err = lib.Generate("./test/bugtracker/migrations", "./test/bugtracker/queries", "./test/bugtracker/store/queries.go", "store")
	if err != nil {
		panic(err)
	}
}
