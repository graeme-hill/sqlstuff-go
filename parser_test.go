package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasicSelect(t *testing.T) {
	statements, err := Parse("SELECT foo FROM bar")
	require.NoError(t, err)
	require.Len(t, statements, 1)

	selectStmt, ok := statements[0].(Select)
	require.True(t, ok)
	require.Len(t, selectStmt.Fields, 1)

	field := selectStmt.Fields[0]
	col, ok := field.Expr.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "foo", col.ColumnName)
	require.Equal(t, "", col.TableName)
	require.Equal(t, "", field.Alias)

	target := selectStmt.From
	require.Equal(t, "", target.Alias)
	require.Nil(t, target.Subselect)
	require.Equal(t, "bar", target.TableName)
}

func TestSubSelect(t *testing.T) {
	statements, err := Parse("SELECT things.foo as stuff FROM (select bar from blah) AS things")
	require.NoError(t, err)
	require.Len(t, statements, 1)

	selectStmt, ok := statements[0].(Select)
	require.True(t, ok)
	require.Len(t, selectStmt.Fields, 1)

	field := selectStmt.Fields[0]
	col, ok := field.Expr.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "foo", col.ColumnName)
	require.Equal(t, "things", col.TableName)
	require.Equal(t, "stuff", field.Alias)

	target := selectStmt.From
	require.Equal(t, "things", target.Alias)
	require.NotNil(t, target.Subselect)
	require.Equal(t, "", target.TableName)
}

func TestMultiSelect(t *testing.T) {
	statements, err := Parse(`
		SELECT foo FROM bar;
		SELECT hello FROM world;
`)
	require.NoError(t, err)
	require.Len(t, statements, 2)

	// First query
	selectStmt, ok := statements[0].(Select)
	require.True(t, ok)
	require.Len(t, selectStmt.Fields, 1)

	field := selectStmt.Fields[0]
	col, ok := field.Expr.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "foo", col.ColumnName)
	require.Equal(t, "", col.TableName)
	require.Equal(t, "", field.Alias)

	target := selectStmt.From
	require.Equal(t, "", target.Alias)
	require.Nil(t, target.Subselect)
	require.Equal(t, "bar", target.TableName)

	// Second query
	selectStmt, ok = statements[1].(Select)
	require.True(t, ok)
	require.Len(t, selectStmt.Fields, 1)

	field = selectStmt.Fields[0]
	col, ok = field.Expr.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "hello", col.ColumnName)
	require.Equal(t, "", col.TableName)
	require.Equal(t, "", field.Alias)

	target = selectStmt.From
	require.Equal(t, "", target.Alias)
	require.Nil(t, target.Subselect)
	require.Equal(t, "world", target.TableName)
}

func TestCreateTable(t *testing.T) {
	statements, err := Parse("CREATE TABLE people(id int not null primary key, name varchar(200))")
	require.NoError(t, err)
	require.Len(t, statements, 1)

	create, ok := statements[0].(CreateTable)
	require.True(t, ok)
	require.Equal(t, "people", create.Name)
	require.Len(t, create.Columns, 2)
	require.Equal(t, "id", create.Columns[0].Name)
	require.Equal(t, DataTypeInteger, create.Columns[0].Type)
	require.Equal(t, 0, create.Columns[0].Param1)
	require.Equal(t, 0, create.Columns[0].Param2)
	require.False(t, create.Columns[0].Nullable)
	require.Equal(t, "name", create.Columns[1].Name)
	require.Equal(t, DataTypeVarChar, create.Columns[1].Type)
	require.Equal(t, 200, create.Columns[1].Param1)
	require.Equal(t, 0, create.Columns[1].Param2)
	require.True(t, create.Columns[1].Nullable)
}
