package test

import (
	"context"
	"testing"

	"github.com/graeme-hill/sqlstuff-go/lib"
	basic "github.com/graeme-hill/sqlstuff-go/test/basic/store"
	bugtracker "github.com/graeme-hill/sqlstuff-go/test/bugtracker/store"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	connStr := "user=postgres password=password"
	client, err := basic.NewDBClient(connStr)
	require.NoError(t, err)

	ctx := context.Background()
	err = lib.RunMigrations(ctx, "./basic/migrations", connStr)
	require.NoError(t, err)

	users, err := client.GetUsers()
	require.NoError(t, err)

	require.Len(t, users, 2)

	require.Equal(t, "Graeme", users[0].FirstName)
	require.Equal(t, "Hill", users[0].LastName)

	require.Equal(t, "Graeme", users[1].FirstName)
	require.Equal(t, "Hill", users[1].LastName)
}

func TestBugTracker(t *testing.T) {
	connStr := "user=postgres password=password"
	client, err := bugtracker.NewDBClient(connStr)
	require.NoError(t, err)

	ctx := context.Background()
	err = lib.RunMigrations(ctx, "./basic/migrations", connStr)
	require.NoError(t, err)

	_, _, err = client.GetIssue(1, 2)
	require.NoError(t, err)
}
