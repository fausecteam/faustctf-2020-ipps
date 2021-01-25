package credit

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

var formDecoder = schema.NewDecoder()

// Card is the representation of a single credit card.
type Card struct {
	// ID is the (internal) unique identifier of the credit card.
	ID uuid.UUID `schema:"-" json:"id"`
	// Number is the credit card's number.
	Number string `schema:"number,required" json:"number"`
	// User is the user to which the credit card belongs.
	User *user.User `schema:"-" json:"-"`
}

// NewCard returns a new credit card for user
func NewCard(user *user.User) (*Card, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &Card{
		ID:   id,
		User: user,
	}, nil
}

// NewCard parses the request's form and fills a new credit card
// according to the form's values for user.
func NewCardFromForm(user *user.User, r *http.Request) (*Card, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	c, err := NewCard(user)
	if err != nil {
		return nil, err
	}
	err = formDecoder.Decode(c, r.PostForm)
	if err != nil {
		return nil, err
	}

	return c, nil
}
