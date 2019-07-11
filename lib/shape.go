package lib

import (
	"errors"
	"fmt"
	"strings"
)

// Enum for the possible return values of a query. Examples:
//   QueryResultTypeManyRows: SELECT foo FROM bar
//   QueryResultTypeOneRow: SELECT foo FROM bar LIMIT 1
//   QueryResultTypeCommand: INSERT INTO foo (bar) VALUES ('hello')
type queryResultType int

const (
	QueryResultTypeManyRows queryResultType = iota
	QueryResultTypeOneRow
	QueryResultTypeCommand
)

type Shape struct {
	Columns []ColumnDefinition
	Type    queryResultType
}

func getShape(stmt Statement, model Model) (Shape, error) {
	switch typed := stmt.(type) {
	case Select:
		return getSelectShape(typed, model)
	case Insert:
		return getInsertShape(typed, model)
	default:
		return Shape{}, errors.New("getShape not implemented for this type of statement yet")
	}
}

// A normal insert that looks like INSERT INTO foo (x) VALUES (1) will not have
// any result columns, but it you use RETURNING then it will behave like a
// SELECT.
func getInsertShape(query Insert, model Model) (Shape, error) {
	// Since RETURNS is not implemented yet always assume no fields are selected
	return Shape{
		Columns: []ColumnDefinition{},
		Type:    QueryResultTypeCommand,
	}, nil
}

// Returns the data types and names of the columns that will come out of the
// given query/model pair.
func getSelectShape(query Select, model Model) (Shape, error) {
	available, err := getAvailableColumns(query, model)
	if err != nil {
		return Shape{}, err
	}

	resultColumns := []ColumnDefinition{}
	for _, field := range query.Fields {
		def, err := fieldAsColumnDefinition(field, model, available)
		if err != nil {
			return Shape{}, err
		}
		resultColumns = append(resultColumns, def)
	}

	resultType := getSelectResultType(query)

	return Shape{
		Columns: resultColumns,
		Type:    resultType,
	}, nil
}

func conditionCoversConstraint(cond Condition, constraint Constraint) bool {
	// TO DO
	return false
}

func getSelectResultType(s Select, model Model) queryResultType {
	if s.Limit.HasLimit && s.Limit.Count == 1 {
		// If the top level query explcitly includes "LIMIT 1"
		return QueryResultTypeOneRow
	}
	
	isMany := true

	for 

	for _, constraint := range model.Tables[]

	return QueryResultTypeManyRows
}

func fieldAsColumnDefinition(
	field Field,
	model Model,
	available map[string][]ColumnDefinition,
) (ColumnDefinition, error) {
	def, err := exprAsColumnDefinition(field.Expr, model, available)
	if err != nil {
		return ColumnDefinition{}, err
	}

	if len(field.Alias) > 0 {
		def.Name = field.Alias
	}

	return def, nil
}

func exprAsColumnDefinition(
	expr Expression,
	model Model,
	available map[string][]ColumnDefinition,
) (ColumnDefinition, error) {
	switch typed := expr.(type) {
	case ColumnExpression:
		return findColumn(typed.TableName, typed.ColumnName, available)
	case FunctionExpression:
		return getFuncReturnType(typed)
	default:
		return ColumnDefinition{}, errors.New("Expression type not implemented yet")
	}
}

func getFuncReturnType(fnExpr FunctionExpression) (ColumnDefinition, error) {
	switch strings.ToUpper(fnExpr.FuncName) {
	case "COUNT":
		return ColumnDefinition{
			Type: DataTypeBigInt,
		}, nil
	default:
		return ColumnDefinition{}, fmt.Errorf("Function '%s' not supported", fnExpr.FuncName)
	}
}

func findColumn(
	table string,
	column string,
	available map[string][]ColumnDefinition,
) (ColumnDefinition, error) {
	if len(table) > 0 {
		return findAliasedColumn(table, column, available)
	}
	return findUnaliasedColumn(column, available)
}

func findAliasedColumn(
	table string,
	column string,
	available map[string][]ColumnDefinition,
) (ColumnDefinition, error) {
	defs, ok := available[table]
	if !ok {
		return ColumnDefinition{}, fmt.Errorf("Invalid table/alias '%s'", table)
	}

	for _, def := range defs {
		if def.Name == column {
			return def, nil
		}
	}

	return ColumnDefinition{}, fmt.Errorf("Column '%s' not found on '%s'", column, table)
}

func findUnaliasedColumn(
	column string,
	available map[string][]ColumnDefinition,
) (ColumnDefinition, error) {
	found := false
	result := ColumnDefinition{}

	for _, defs := range available {
		for _, def := range defs {
			if def.Name == column {
				if found {
					return ColumnDefinition{}, fmt.Errorf("Ambiguous column '%s'", column)
				}
				result = def
				found = true
			}
		}
	}

	if found {
		return result, nil
	}
	return ColumnDefinition{}, fmt.Errorf("Column '%s' not found", column)
}

func getAvailableColumns(
	query Select,
	model Model,
) (map[string][]ColumnDefinition, error) {
	// Map to store all columns available for select list
	available := map[string][]ColumnDefinition{}

	// FROM clause
	err := addTargetTable(available, model, query.From)
	if err != nil {
		return nil, err
	}

	// JOINs
	for _, join := range query.Joins {
		err = addTargetTable(available, model, join.Target)
		if err != nil {
			return nil, err
		}
	}

	return available, nil
}

func addTargetTable(
	available map[string][]ColumnDefinition,
	model Model,
	target TargetTable,
) error {
	if len(target.TableName) > 0 {
		// Just a table name (possibly aliased)
		key := target.TableName
		if len(target.Alias) > 0 {
			key = target.Alias
		}
		tbl, ok := model.Tables[target.TableName]
		if !ok {
			return fmt.Errorf("Unknown table '%s'", target.TableName)
		}

		_, exists := available[key]
		if exists {
			return fmt.Errorf("Duplicate alias '%s'", key)
		}

		available[key] = tbl.Columns
	} else {
		// With a subselect (must be aliased)
		if len(target.Alias) <= 0 {
			return errors.New("Subselect requires alias")
		}
		_, exists := available[target.Alias]
		if exists {
			return fmt.Errorf("Duplicate alias '%s'", target.Alias)
		}
		shape, err := getShape(*target.Subselect, model)
		if err != nil {
			return err
		}
		available[target.Alias] = shape.Columns
	}

	return nil
}
