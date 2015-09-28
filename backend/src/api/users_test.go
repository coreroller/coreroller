package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	defaultTeamID = "d89342dc-9214-441d-a4af-bdd837a3b239"
)

func TestGetUser(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	_, err := a.GetUser("non-existent")
	assert.Error(t, err)

	user, err := a.GetUser("admin")
	assert.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, defaultTeamID, user.TeamID)
	assert.Equal(t, "8b31292d4778582c0e5fa96aee5513f1", user.Secret)
}

func TestUpdateUserPassword(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	err := a.UpdateUserPassword("non-existent", "new-password")
	assert.Error(t, err)

	err = a.UpdateUserPassword("admin", "new-password")
	assert.NoError(t, err)

	user, err := a.GetUser("admin")
	assert.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, defaultTeamID, user.TeamID)
	assert.NotEqual(t, "8b31292d4778582c0e5fa96aee5513f1", user.Secret)
	assert.Equal(t, "4ae4f3f6b4be19d33d41f66aa9c79bea", user.Secret)
}
