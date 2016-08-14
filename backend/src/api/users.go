package api

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	// Realm used for basic authentication.
	Realm = "coreroller.org"
)

var (
	// ErrUpdatingPassword indicates that something went wrong while updating
	// the user's password.
	ErrUpdatingPassword = errors.New("coreroller: error updating password")
)

// User represents a CoreRoller user.
type User struct {
	ID        string    `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Secret    string    `db:"secret" json:"secret"`
	CreatedTs time.Time `db:"created_ts" json:"-"`
	TeamID    string    `db:"team_id" json:"team_id"`
}

// GetUser returns the user identified by the username provided.
func (api *API) GetUser(username string) (*User, error) {
	var user User

	err := api.dbR.SelectDoc("*").
		From("users").
		Where("username = $1", username).
		QueryStruct(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUserPassword updates the password of the provided user.
func (api *API) UpdateUserPassword(username, newPassword string) error {
	secret, err := api.GenerateUserSecret(username, newPassword)
	if err != nil {
		return err
	}

	result, err := api.dbR.
		Update("users").
		Set("secret", secret).
		Where("username = $1", username).
		Exec()

	if err != nil || result.RowsAffected == 0 {
		return ErrUpdatingPassword
	}

	return nil
}

// GenerateUserSecret generates a md5 hash from the username and password
// provided (username:realm:password).
func (api *API) GenerateUserSecret(username, password string) (string, error) {
	h := md5.New()
	if _, err := io.WriteString(h, username+":"+Realm+":"+password); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
