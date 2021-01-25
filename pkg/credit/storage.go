package credit

import (
	"errors"

	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

var ErrNoCards = errors.New("user does not have any credit cards")

// Inserter is the interfaces for insertying credit cards into
// a persistent storage.
//
// Insert inserts c into the Inserter's underlying storage.
type Inserter interface {
	Insert(c *Card) error
}

// Accesser is the interface wrapping the ByUser method.
//
// ByUser returns all of a users credit cards.
type Accesser interface {
	ByUser(u *user.User) ([]*Card, error)
}

// Updater is the interfaces wrapping the Update method.
//
// Update updates m in the Updater's underlying storage.
type Updater interface {
	Update(c *Card) error
}

// Deleter is the interfaces wrapping the Delete method.
//
// Delete removes m from the Deleter's underlying storage.
type Deleter interface {
	Delete(c *Card) error
}

// Storage is the interface wrapping all interfaces for inserting,
// retrieving, updating and deleting credit card information.
type Storage interface {
	Inserter
	Accesser
	Updater
	Deleter
}
