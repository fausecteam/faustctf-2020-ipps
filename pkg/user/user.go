// Package user contains interfaces and data structures for working with users.
package user

import (
	"context"
	"net/mail"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User is the type representing a single user of the website.
type User struct {
	ID       uuid.UUID     `json:"id"`
	Username string        `json:"username"`
	Password []byte        `json:"-"`
	Email    *mail.Address `json:"email"`
	Name     string        `json:"name"`
}

// New initializes and returns a new User object.
// The user's ID field is a randomly generated UUID.
func New(username, email, password string) (*User, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Username: username,
		Password: hash,
		Email:    addr,
		Name:     "",
	}, nil
}

// PasswordEquals returns, whether the given plaintext password is
// equal to the user's current password.
func (u *User) PasswordEquals(plaintext string) bool {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(plaintext))
	if err != nil {
		return false
	}

	return true
}

// SetPassword updates a user's password hash member, using plaintext
// as the new password.
func (u *User) SetPassword(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = hash

	return nil
}

type ctxKey int

const key ctxKey = iota

// New Context returns a copy of ctx, appending user to ctx's values.
func NewContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, key, user)
}

// FromContext returns the user object stored in ctx. If ctx does not
// have a User object, the second returns value is false.
func FromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(key).(*User)

	return u, ok
}

// MustFromContext is like FromContext, except that it panics if ctx does not
// contain a User object.
func MustFromContext(ctx context.Context) *User {
	u, ok := FromContext(ctx)
	if !ok {
		panic("context has no user")
	}

	return u
}
