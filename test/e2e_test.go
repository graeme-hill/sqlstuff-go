package test

import (
	"context"
	"testing"

	"github.com/graeme-hill/sqlstuff-go/lib"
	"github.com/graeme-hill/sqlstuff-go/test/store"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	connStr := "user=postgres password=password"
	client, err := store.NewDBClient(connStr)
	require.NoError(t, err)

	ctx := context.Background()
	err = lib.RunMigrations(ctx, "./migrations", connStr)
	require.NoError(t, err)

	users, err := client.GetUsers()
	require.NoError(t, err)

	require.Len(t, users, 2)

	require.Equal(t, "Graeme", users[0].FirstName)
	require.Equal(t, "Hill", users[0].LastName)

	require.Equal(t, "Graeme", users[1].FirstName)
	require.Equal(t, "Hill", users[1].LastName)
}
