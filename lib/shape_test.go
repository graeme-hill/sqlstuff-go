package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetShape(t *testing.T) {
	migrations, err := ReadMigrationsDir("../test/basic/migrations")
	require.NoError(t, err)
	model, err := ModelFromMigrations(migrations)
	require.NoError(t, err)
	prog, err := Parse(`
		SELECT id, email FROM users;
		INSERT INTO users (id, email) VALUES (1, 'foo@bar.com');
		SELECT id, email FROM users LIMIT 1;
		SELECT id, email FROM users LIMIT 2;
		SELECT id, email FROM users WHERE id=$id;
		SELECT id, email FROM users WHERE email=$email;`)
	require.NoError(t, err)
	require.Len(t, prog.Statements, 6)

	// SELECT id, email FROM users
	query, ok := prog.Statements[0].(Select)
	require.True(t, ok)

	shape, err := getShape(query, model)
	require.NoError(t, err)
	require.Len(t, shape.Columns, 2)
	require.Equal(t, QueryResultTypeManyRows, shape.Type)

	require.Equal(t, "id", shape.Columns[0].Name)
	require.Equal(t, "email", shape.Columns[1].Name)

	require.Equal(t, DataTypeInteger, shape.Columns[0].Type)
	require.Equal(t, DataTypeVarChar, shape.Columns[1].Type)
	require.Equal(t, 200, shape.Columns[1].Param1)

	// INSERT INTO users (id, email) VALUES (1, 'foo@bar.com')
	cmd, ok := prog.Statements[1].(Insert)
	require.True(t, ok)

	shape, err = getShape(cmd, model)
	require.NoError(t, err)
	require.Equal(t, QueryResultTypeCommand, shape.Type)
	require.Empty(t, shape.Columns)

	// SELECT id, email FROM users LIMIT 1
	query, ok = prog.Statements[2].(Select)
	require.True(t, ok)
	shape, err = getShape(query, model)
	require.NoError(t, err)
	require.Equal(t, QueryResultTypeOneRow, shape.Type)

	// SELECT id, email FROM users LIMIT 2
	query, ok = prog.Statements[3].(Select)
	require.True(t, ok)
	shape, err = getShape(query, model)
	require.NoError(t, err)
	require.Equal(t, QueryResultTypeManyRows, shape.Type)

	// SELECT id, email FROM users WHERE id=$id
	query, ok = prog.Statements[4].(Select)
	require.True(t, ok)
	shape, err = getShape(query, model)
	require.NoError(t, err)
	require.Equal(t, QueryResultTypeOneRow, shape.Type)

	// SELECT id, email FROM users WHERE email=$email
	query, ok = prog.Statements[5].(Select)
	require.True(t, ok)
	shape, err = getShape(query, model)
	require.NoError(t, err)
	require.Equal(t, QueryResultTypeManyRows, shape.Type)
}
