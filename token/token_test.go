package token

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlow(t *testing.T) {
	storage := GetTokenStorageInstance()
	token := storage.GenerateToken()

	_, err := storage.GetUsernameByToken(token)
	require.Error(t, err)
	_, err = storage.GetRoleByToken(token)
	require.Error(t, err)
	err = storage.DeleteToken("key-username", "role")
	require.Error(t, err)

	storage.AddToken("key-username", token, "role")

	username, err := storage.GetUsernameByToken(token)
	require.NoError(t, err)
	require.Equal(t, "key-username", username)

	role, err := storage.GetRoleByToken(token)
	require.NoError(t, err)
	require.Equal(t, "role", role)

	err = storage.DeleteToken("key-username", token)
	require.NoError(t, err)

	_, err = storage.GetUsernameByToken(token)
	require.Error(t, err)
	_, err = storage.GetRoleByToken(token)
	require.Error(t, err)
}
