package main

import "github.com/graeme-hill/sqlstuff-go/lib"

func main() {
	err := lib.RunMigrations("./test/migrations", "dbname=graeme")
	if err != nil {
		panic(err)
	}
}