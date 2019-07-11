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

	batches, err := ReadQueriesFromDir("../test/basic/queries", model)
	require.NoError(t, err)
	require.Len(t, batches, 1)

	batch := batches[0]
	require.Len(t, batch.AST, 1)
	require.Len(t, batch.Shapes, 1)
	require.Equal(t, "get_users", batch.Name)

	shape := batch.Shapes[0]
	require.Equal(t, QueryResultTypeManyRows, shape.Type)
	columns := shape.Columns
	require.Len(t, columns, 5)

	require.Equal(t, "id", columns[0].Name)
	require.Equal(t, "email", columns[1].Name)
	require.Equal(t, "first_name", columns[2].Name)
	require.Equal(t, "last_name", columns[3].Name)
	require.Equal(t, "group_name", columns[4].Name)

	require.Equal(t, DataTypeInteger, columns[0].Type)
	require.Equal(t, DataTypeVarChar, columns[1].Type)
	require.Equal(t, 200, columns[1].Param1)
	require.Equal(t, DataTypeVarChar, columns[2].Type)
	require.Equal(t, 200, columns[2].Param1)
	require.Equal(t, DataTypeVarChar, columns[3].Type)
	require.Equal(t, 200, columns[3].Param1)
	require.Equal(t, DataTypeVarChar, columns[4].Type)
	require.Equal(t, 200, columns[4].Param1)
}

func TestGetShapeBugTracker(t *testing.T) {
	migrations, err := ReadMigrationsDir("../test/bugtracker/migrations")
	require.NoError(t, err)
	model, err := ModelFromMigrations(migrations)
	require.NoError(t, err)

	batches, err := ReadQueriesFromDir("../test/bugtracker/queries", model)
	require.NoError(t, err)
	require.Len(t, batches, 7)
}
