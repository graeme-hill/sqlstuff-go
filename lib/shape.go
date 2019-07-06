package lib

import (
	"errors"
	"fmt"
	"strings"
)

func getShape(stmt Statement, model Model) ([]ColumnDefinition, error) {
	switch typed := stmt.(type) {
	case Select:
		return getSelectShape(typed, model)
	default:
		return nil, errors.New("getShape not implemented for this type of statement yet")
	}
}

// Returns the data types and names of the columns that will come out of the
// given query/model pair.
func getSelectShape(query Select, model Model) ([]ColumnDefinition, error) {
	available, err := getAvailableColumns(query, model)
	if err != nil {
		return nil, err
	}

	resultColumns := []ColumnDefinition{}
	for _, field := range query.Fields {
		def, err := fieldAsColumnDefinition(field, model, available)
		if err != nil {
			return nil, err
		}
		resultColumns = append(resultColumns, def)
	}

	return resultColumns, nil
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
	err := addSelectTarget(available, model, query.From)
	if err != nil {
		return nil, err
	}

	// JOINs
	for _, join := range query.Joins {
		err = addSelectTarget(available, model, join.Target)
		if err != nil {
			return nil, err
		}
	}

	return available, nil
}

func addSelectTarget(
	available map[string][]ColumnDefinition,
	model Model,
	target SelectTarget,
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
		subColumns, err := getShape(*target.Subselect, model)
		if err != nil {
			return err
		}
		available[target.Alias] = subColumns
	}

	return nil
}
