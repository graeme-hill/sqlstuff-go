package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/graeme-hill/sqlstuff-go/test/store"
)

func TestAll(t *testing.T) {
	client, err := store.NewDBClient("dbname=graeme")
	require.NoError(t, err)

	users, err := client.GetUsers()
	require.NoError(t, err)

	require.Len(t, users, 2)

	require.Equal(t, "Graeme", users[0].FirstName)
	require.Equal(t, "Hill", users[0].LastName)

	require.Equal(t, "Graeme", users[1].FirstName)
	require.Equal(t, "Hill", users[2].LastName)
}