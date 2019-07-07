package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetShapeBasic(t *testing.T) {
	migrations, err := ReadMigrationsDir("../test/basic/migrations")
	require.NoError(t, err)
	model, err := ModelFromMigrations(migrations)
	require.NoError(t, err)
	statements, err := Parse("SELECT id, email FROM users")
	require.NoError(t, err)
	query, ok := statements[0].(Select)
	require.True(t, ok)

	columns, err := getShape(query, model)
	require.NoError(t, err)
	require.Len(t, columns, 2)

	require.Equal(t, "id", columns[0].Name)
	require.Equal(t, "email", columns[1].Name)

	require.Equal(t, DataTypeInteger, columns[0].Type)
	require.Equal(t, DataTypeVarChar, columns[1].Type)
	require.Equal(t, 200, columns[1].Param1)
}
