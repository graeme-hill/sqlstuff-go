package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasicSelect(t *testing.T) {
	prog, err := Parse("SELECT foo FROM bar")
	require.NoError(t, err)
	require.Len(t, prog.Statements, 1)

	selectStmt, ok := prog.Statements[0].(Select)
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

func TestParameterized(t *testing.T) {
	prog, err := Parse("SELECT foo FROM bar WHERE id=$my_id")
	require.NoError(t, err)
	require.Len(t, prog.Statements, 1)

	require.Len(t, prog.Parameters, 1)
	require.Equal(t, "my_id", prog.Parameters[0].Name)

	selectStmt, ok := prog.Statements[0].(Select)
	require.True(t, ok)
	require.Len(t, selectStmt.Fields, 1)

	where, ok := selectStmt.Where.(BinaryCondition)
	require.True(t, ok)
	id, ok := where.Left.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "id", id.ColumnName)
	param, ok := where.Right.(ParameterExpression)
	require.True(t, ok)
	require.Equal(t, "my_id", param.Name)
}

func TestSubSelect(t *testing.T) {
	prog, err := Parse("SELECT things.foo as stuff FROM (select bar from blah) things")
	require.NoError(t, err)
	require.Len(t, prog.Statements, 1)

	selectStmt, ok := prog.Statements[0].(Select)
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
	prog, err := Parse(`
		SELECT foo FROM bar;
		SELECT hello FROM world;`)
	require.NoError(t, err)
	require.Len(t, prog.Statements, 2)

	// First query
	selectStmt, ok := prog.Statements[0].(Select)
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
	selectStmt, ok = prog.Statements[1].(Select)
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
	prog, err := Parse("CREATE TABLE people(id int not null primary key, name varchar(200))")
	require.NoError(t, err)
	require.Len(t, prog.Statements, 1)

	create, ok := prog.Statements[0].(CreateTable)
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

func TestAddColumn(t *testing.T) {
	prog, err := Parse("ALTER TABLE people ADD COLUMN name VARCHAR(200) NOT NULL")
	require.NoError(t, err)
	require.Len(t, prog.Statements, 1)

	addColumn, ok := prog.Statements[0].(AddColumn)
	require.True(t, ok)
	require.Equal(t, "people", addColumn.TableName)
	require.Equal(t, "name", addColumn.Column.Name)
	require.Equal(t, DataTypeVarChar, addColumn.Column.Type)
	require.False(t, addColumn.Column.Nullable)
	require.Equal(t, 200, addColumn.Column.Param1)
}

func TestSelectFeatures(t *testing.T) {
	prog, err := Parse(`
		SELECT 
			u.name,
			u.email,
			g.name AS group
		FROM users u
		LEFT JOIN user_groups ug ON ug.user_id = u.id
		LEFT JOIN groups g ON g.id = ug.group_id
		WHERE u.id = 77`)
	require.NoError(t, err)
	require.Len(t, prog.Statements, 1)

	selectStmt, ok := prog.Statements[0].(Select)
	require.True(t, ok)
	require.Len(t, selectStmt.Fields, 3)

	field := selectStmt.Fields[0]
	col, ok := field.Expr.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "name", col.ColumnName)
	require.Equal(t, "u", col.TableName)
	require.Equal(t, "", field.Alias)

	field = selectStmt.Fields[1]
	col, ok = field.Expr.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "email", col.ColumnName)
	require.Equal(t, "u", col.TableName)
	require.Equal(t, "", field.Alias)

	field = selectStmt.Fields[2]
	col, ok = field.Expr.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "name", col.ColumnName)
	require.Equal(t, "g", col.TableName)
	require.Equal(t, "group", field.Alias)

	target := selectStmt.From
	require.Equal(t, "u", target.Alias)
	require.Nil(t, target.Subselect)
	require.Equal(t, "users", target.TableName)

	require.Len(t, selectStmt.Joins, 2)

	join := selectStmt.Joins[0]
	require.Equal(t, JoinTypeLeftOuter, join.Type)
	require.Equal(t, "user_groups", join.Target.TableName)
	require.Equal(t, "ug", join.Target.Alias)
	require.Nil(t, join.Target.Subselect)

	on, ok := join.On.(BinaryCondition)
	require.True(t, ok)
	require.Equal(t, BinaryCondOpEqual, on.Op)
	left, ok := on.Left.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "ug", left.TableName)
	require.Equal(t, "user_id", left.ColumnName)
	right, ok := on.Right.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "u", right.TableName)
	require.Equal(t, "id", right.ColumnName)

	join = selectStmt.Joins[1]
	require.Equal(t, JoinTypeLeftOuter, join.Type)
	require.Equal(t, "groups", join.Target.TableName)
	require.Equal(t, "g", join.Target.Alias)
	require.Nil(t, join.Target.Subselect)

	on, ok = join.On.(BinaryCondition)
	require.True(t, ok)
	require.Equal(t, BinaryCondOpEqual, on.Op)
	left, ok = on.Left.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "g", left.TableName)
	require.Equal(t, "id", left.ColumnName)
	right, ok = on.Right.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "ug", right.TableName)
	require.Equal(t, "group_id", right.ColumnName)

	where, ok := selectStmt.Where.(BinaryCondition)
	require.True(t, ok)

	leftWhere, ok := where.Left.(ColumnExpression)
	require.True(t, ok)
	require.Equal(t, "id", leftWhere.ColumnName)
	require.Equal(t, "u", leftWhere.TableName)

	rightWhere, ok := where.Right.(NumberLiteral)
	require.True(t, ok)
	require.Equal(t, "77", rightWhere.Value)
}

func TestInsert(t *testing.T) {
	prog, err := Parse("INSERT INTO users (name, email) VALUES ('Graeme', 'graeme@foobar.com');")
	require.NoError(t, err)
	require.Len(t, prog.Statements, 1)

	insertStmt, ok := prog.Statements[0].(Insert)
	require.True(t, ok)
	require.Equal(t, "users", insertStmt.Target.TableName)

	require.Len(t, insertStmt.Columns, 2)
	require.Len(t, insertStmt.Values, 2)

	require.Equal(t, "name", insertStmt.Columns[0].ColumnName)
	require.Equal(t, "users", insertStmt.Columns[0].TableName)
	require.Equal(t, "email", insertStmt.Columns[1].ColumnName)
	require.Equal(t, "users", insertStmt.Columns[1].TableName)

	expr1, ok := insertStmt.Values[0].(StringLiteral)
	require.True(t, ok)
	require.Equal(t, "Graeme", expr1.Value)

	expr2, ok := insertStmt.Values[1].(StringLiteral)
	require.True(t, ok)
	require.Equal(t, "graeme@foobar.com", expr2.Value)
}

func TestLimit(t *testing.T) {
	prog, err := Parse("SELECT foo FROM bar LIMIT 1")
	require.NoError(t, err)
	require.Len(t, prog.Statements, 1)

	selectStmt, ok := prog.Statements[0].(Select)
	require.True(t, ok)
	require.Len(t, selectStmt.Fields, 1)
	require.True(t, selectStmt.Limit.HasLimit)
	require.Equal(t, selectStmt.Limit.Count, 1)
}
