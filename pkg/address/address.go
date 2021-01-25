// Package address defines interfaces and data structures for working with customer addresses.
package address

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

var ErrAddressAlreadyAdded = errors.New("address: user has already added this address")

type Address struct {
	ID      uuid.UUID  `json:"id" schema:"id"`
	Street  string     `json:"street" schema:"street,required"`
	Zip     string     `json:"zip" schema:"zip,required"`
	City    string     `json:"city" schema:"city,required"`
	Country string     `json:"country" schema:"country,required"`
	Planet  string     `json:"planet" schema:"planet"`
	User    *user.User `json:"-" schema:"-"`
}

// NewForUser creates and returns a new Address, with its User member set to u.
func NewForUser(u *user.User) (*Address, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &Address{
		ID:   id,
		User: u,
	}, nil
}

var formDecoder = schema.NewDecoder()

// NewFromFormForUser parses r's post form, decoding it into an Address,
// using a newly generated ID as the address's ID and setting the address's
// User member to u.
func NewFromFormForUser(r *http.Request, u *user.User) (*Address, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	a := &Address{User: u}
	err = formDecoder.Decode(a, r.PostForm)
	if err != nil {
		return nil, err
	}
	// Ignore user supplied ID
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	a.ID = id

	return a, nil
}

// FromFormForUser parses r's post form into an address, using the id
// provided in the form as the address's ID and setting the address's
// User member to u.
func FromFormForUser(r *http.Request, u *user.User) (*Address, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	a := &Address{User: u}
	err = formDecoder.Decode(a, r.PostForm)
	if err != nil {
		return nil, err
	}

	return a, nil
}
