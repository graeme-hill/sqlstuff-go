package main

import (
	"github.com/graeme-hill/sqlstuff-go/lib"
	"context"
)

func main() {
	ctx := context.Background()
	err := lib.RunMigrations(ctx, "./test/migrations", "dbname=graeme")
	if err != nil {
		panic(err)
	}
}