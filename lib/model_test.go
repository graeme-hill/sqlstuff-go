package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelBuilder(t *testing.T) {
	b := NewModelBuilder()

	err := b.handleStmt(CreateTable{
		Name: "users",
		Columns: []ColumnDefinition{
			ColumnDefinition{Name: "id"},
			ColumnDefinition{Name: "name"},
		},
	})

	require.NoError(t, err)
	require.Len(t, b.model.Tables, 1)
	require.Len(t, b.model.Tables["users"].Columns, 2)
	require.Equal(t, "id", b.model.Tables["users"].Columns[0].Name)
	require.Equal(t, "name", b.model.Tables["users"].Columns[1].Name)

	err = b.handleStmt(AddColumn{
		TableName: "users",
		Column:    ColumnDefinition{Name: "email"},
	})

	require.NoError(t, err)
	require.Len(t, b.model.Tables, 1)
	require.Len(t, b.model.Tables["users"].Columns, 3)
	require.Equal(t, "id", b.model.Tables["users"].Columns[0].Name)
	require.Equal(t, "name", b.model.Tables["users"].Columns[1].Name)
	require.Equal(t, "email", b.model.Tables["users"].Columns[2].Name)

	err = b.handleStmt(DropColumn{
		TableName:  "users",
		ColumnName: "name",
	})

	require.NoError(t, err)
	require.Len(t, b.model.Tables, 1)
	require.Len(t, b.model.Tables["users"].Columns, 2)
	require.Equal(t, "id", b.model.Tables["users"].Columns[0].Name)
	require.Equal(t, "email", b.model.Tables["users"].Columns[1].Name)
}

func TestModelBuilderDropColumnFail(t *testing.T) {
	b := NewModelBuilder()

	err := b.handleStmt(CreateTable{
		Name: "users",
		Columns: []ColumnDefinition{
			ColumnDefinition{Name: "id"},
			ColumnDefinition{Name: "name"},
		},
	})
	require.NoError(t, err)

	err = b.handleStmt(DropColumn{
		TableName:  "users",
		ColumnName: "doesnotexist",
	})
	require.Error(t, err)

	err = b.handleStmt(DropColumn{
		TableName:  "doesnotexist",
		ColumnName: "name",
	})
	require.Error(t, err)
}

func TestModelBuilderCreateTableFail(t *testing.T) {
	b := NewModelBuilder()

	err := b.handleStmt(CreateTable{
		Name: "users",
		Columns: []ColumnDefinition{
			ColumnDefinition{Name: "id"},
			ColumnDefinition{Name: "name"},
		},
	})
	require.NoError(t, err)

	err = b.handleStmt(CreateTable{
		Name: "users",
		Columns: []ColumnDefinition{
			ColumnDefinition{Name: "id2"},
			ColumnDefinition{Name: "name2"},
		},
	})
	require.Error(t, err)
}

func TestModelBuilderAddColumnFail(t *testing.T) {
	b := NewModelBuilder()

	err := b.handleStmt(CreateTable{
		Name: "users",
		Columns: []ColumnDefinition{
			ColumnDefinition{Name: "id"},
			ColumnDefinition{Name: "name"},
		},
	})
	require.NoError(t, err)

	err = b.handleStmt(AddColumn{
		TableName: "users",
		Column:    ColumnDefinition{Name: "name"},
	})
	require.Error(t, err)
}
