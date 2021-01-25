package user

import (
	"errors"
	"net/http"
	"net/mail"

	"github.com/gorilla/schema"
)

type LoginForm struct {
	Username string `schema:"username,required"`
	Password string `schema:"password,required"`
}

var formDecoder = schema.NewDecoder()

func ParseLoginForm(r *http.Request) (*LoginForm, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	lf := &LoginForm{}
	err = formDecoder.Decode(lf, r.PostForm)
	if err != nil {
		return nil, err
	}

	return lf, nil
}

var ErrPasswdConfirmMismatch = errors.New("password and its confirmation do not match")

type registrationForm struct {
	Username             string `schema:"username,required"`
	Email                string `schema:"email,required"`
	Password             string `schema:"password,required"`
	PasswordConfirmation string `schema:"password-confirm,required"`
	Name                 string `schema:"name,required"`
}

// ParseRegistrationForm parses a POST request's form data and
// returns a new User initialized with values from the form.
func ParseRegistrationForm(r *http.Request) (*User, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	f := &registrationForm{}
	err = formDecoder.Decode(f, r.PostForm)
	if err != nil {
		return nil, err
	}
	if f.Password != f.PasswordConfirmation {
		return nil, ErrPasswdConfirmMismatch
	}
	u, err := New(f.Username, f.Email, f.Password)
	if err != nil {
		return nil, err
	}
	u.Name = f.Name

	return u, nil
}

var (
	ErrPasswordRequired = errors.New("current password is required to update password")
)

type editForm struct {
	Name                    string `schema:"name,required"`
	Email                   string `schema:"email,required"`
	CurrentPassword         string `schema:"current-password"`
	NewPassword             string `schema:"new-password"`
	NewPasswordConfirmation string `schema:"new-password-confirmation"`
}

func UpdateFromEditForm(u *User, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	f := &editForm{}
	err = formDecoder.Decode(f, r.PostForm)
	if err != nil {
		return err
	}
	u.Name = f.Name
	m, err := mail.ParseAddress(f.Email)
	if err != nil {
		return err
	}
	u.Email = m

	if f.NewPassword == "" {
		return nil
	}
	if f.CurrentPassword == "" {
		return ErrPasswordRequired
	}
	if f.NewPassword != f.NewPasswordConfirmation {
		return ErrPasswdConfirmMismatch
	}
	err = u.SetPassword(f.NewPassword)
	if err != nil {
		return err
	}

	return nil
}
