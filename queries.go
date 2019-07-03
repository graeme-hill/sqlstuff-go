package main

import (
	"io/ioutil"
	"path"
)

type Query struct {
	SQL    string
	AST    []Statement
	Shapes [][]ColumnDefinition
}

func ReadQueriesFromDir(dir string, model Model) ([]Query, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	queries := []Query{}

	for _, file := range files {
		filePath := path.Join(dir, file.Name())
		q, err := ReadQueryFromFile(filePath, model)
		if err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}

	return queries, nil
}

func ReadQueryFromFile(filePath string, model Model) (Query, error) {
	// Read text from file
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return Query{}, err
	}
	query := Query{
		SQL:    string(bytes),
		Shapes: [][]ColumnDefinition{},
	}

	// Parse the file to get AST
	statements, err := Parse(query.SQL)
	if err != nil {
		return Query{}, err
	}
	query.AST = statements

	// Extract the shape of each statement within the query
	for _, stmt := range statements {
		defs, err := getShape(stmt, model)
		if err != nil {
			return Query{}, err
		}
		query.Shapes = append(query.Shapes, defs)
	}

	return query, nil
}
