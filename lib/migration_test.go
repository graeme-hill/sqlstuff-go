package lib

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	migrations, err := ReadMigrationsDir("../test/basic/migrations")
	require.NoError(t, err)
	require.Len(t, migrations, 2)
	require.True(t, strings.HasPrefix(migrations[0].UpSQL, "CREATE TABLE"))

	model, err := ModelFromMigrations(migrations)
	require.NoError(t, err)

	require.Len(t, model.Tables, 3)

	// users
	require.Equal(t, "users", model.Tables["users"].Name)
	require.Len(t, model.Tables["users"].Columns, 4)

	require.Equal(t, "id", model.Tables["users"].Columns[0].Name)
	require.Equal(t, "email", model.Tables["users"].Columns[1].Name)
	require.Equal(t, "first_name", model.Tables["users"].Columns[2].Name)
	require.Equal(t, "last_name", model.Tables["users"].Columns[3].Name)

	require.Equal(t, DataTypeInteger, model.Tables["users"].Columns[0].Type)
	require.Equal(t, DataTypeVarChar, model.Tables["users"].Columns[1].Type)
	require.Equal(t, 200, model.Tables["users"].Columns[1].Param1)
	require.Equal(t, DataTypeVarChar, model.Tables["users"].Columns[2].Type)
	require.Equal(t, 200, model.Tables["users"].Columns[2].Param1)
	require.Equal(t, DataTypeVarChar, model.Tables["users"].Columns[3].Type)
	require.Equal(t, 200, model.Tables["users"].Columns[3].Param1)

	require.False(t, model.Tables["users"].Columns[0].Nullable)
	require.False(t, model.Tables["users"].Columns[1].Nullable)
	require.True(t, model.Tables["users"].Columns[2].Nullable)
	require.True(t, model.Tables["users"].Columns[3].Nullable)

	// groups
	require.Equal(t, "groups", model.Tables["groups"].Name)
	require.Len(t, model.Tables["groups"].Columns, 2)

	require.Equal(t, "id", model.Tables["groups"].Columns[0].Name)
	require.Equal(t, "name", model.Tables["groups"].Columns[1].Name)

	require.Equal(t, DataTypeInteger, model.Tables["groups"].Columns[0].Type)
	require.Equal(t, DataTypeVarChar, model.Tables["groups"].Columns[1].Type)
	require.Equal(t, 200, model.Tables["groups"].Columns[1].Param1)

	require.False(t, model.Tables["groups"].Columns[0].Nullable)
	require.False(t, model.Tables["groups"].Columns[1].Nullable)

	// user groups
	require.Equal(t, "user_groups", model.Tables["user_groups"].Name)
	require.Len(t, model.Tables["user_groups"].Columns, 2)

	require.Equal(t, "user_id", model.Tables["user_groups"].Columns[0].Name)
	require.Equal(t, "group_id", model.Tables["user_groups"].Columns[1].Name)

	require.Equal(t, DataTypeInteger, model.Tables["user_groups"].Columns[0].Type)
	require.Equal(t, DataTypeInteger, model.Tables["user_groups"].Columns[1].Type)

	require.False(t, model.Tables["user_groups"].Columns[0].Nullable)
	require.False(t, model.Tables["user_groups"].Columns[1].Nullable)
}
