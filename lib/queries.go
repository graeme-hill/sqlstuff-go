package lib

import (
	"io/ioutil"
	"path"
	"strings"
)

type QueryBatch struct {
	Name       string
	SQL        string
	AST        []Statement
	Shapes     []Shape
	Parameters []Parameter
}

func ReadQueriesFromDir(dir string, model Model) ([]QueryBatch, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	batches := []QueryBatch{}

	for _, file := range files {
		filePath := path.Join(dir, file.Name())
		b, err := ReadBatchFromFile(filePath, model)
		if err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}

	return batches, nil
}

func ReadBatchFromFile(filePath string, model Model) (QueryBatch, error) {
	// Read text from file
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return QueryBatch{}, err
	}
	query := QueryBatch{
		Name:   batchNameFromPath(filePath),
		SQL:    string(bytes),
		Shapes: []Shape{},
	}

	// Parse the file to get AST
	prog, err := Parse(query.SQL)
	if err != nil {
		return QueryBatch{}, err
	}
	query.AST = prog.Statements
	query.Parameters = prog.Parameters

	// Extract the shape of each statement within the query
	for _, stmt := range prog.Statements {
		shape, err := getShape(stmt, model)
		if err != nil {
			return QueryBatch{}, err
		}
		query.Shapes = append(query.Shapes, shape)
	}

	return query, nil
}

func batchNameFromPath(filePath string) string {
	_, fileName := path.Split(filePath)
	parts := strings.Split(fileName, ".")
	return parts[0]
}
