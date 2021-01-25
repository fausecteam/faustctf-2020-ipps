package user

import (
	"errors"
	"net/mail"

	"github.com/google/uuid"
)

var ErrUserExists = errors.New("a user with that username already exists")
var ErrUserNotExists = errors.New("user does not exist")
var ErrEmailExists = errors.New("a user with that email address is already registered")

// Inserter is the interface for inserting users to a storage.
//
// Insert inserts the user into the Inserter's underyling storage.
type Inserter interface {
	Insert(user *User) error
}

// Accesser is the interface wrapping methods for accessing user data from its
// underlying storage.
//
// ByID returns the user identified by id or nil, if the user does not exist.
//
// ByUsername returns the user identified by username or nil if no user
// with that name exists.

// ByEmail returns the user identified by the email address or nil if no user
// with that email address exists.
type Accesser interface {
	ByID(id uuid.UUID) (*User, error)
	ByUsername(username string) (*User, error)
	ByEmail(address *mail.Address) (*User, error)
}

// Update is the interface wrapping the Update method.
//
// Update updates user in the Updater's underlying storage.
type Updater interface {
	Update(user *User) error
}

// Deleter is the interface wrapping the Delete method.
//
// Delete deletes a user from the Deleter's underlying storage.
type Deleter interface {
	Delete(user *User) error
}

// Storage is the interface wrapping methods for creating, accessing, updating
// and deleting users from its underlying storage.
type Storage interface {
	Inserter
	Accesser
	Updater
	Deleter
}
