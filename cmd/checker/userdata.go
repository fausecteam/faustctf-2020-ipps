package main

import (
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
)

func newUsername() string {
	username := randomdata.SillyName()
	username = strings.Replace(username, " ", "_", -1)
	mathrand.Seed(time.Now().Unix())
	n := mathrand.Intn(10000)

	return fmt.Sprintf("%s%d", username, n)
}

const passwordLength = 12

var passwordAlphabet = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyx0123456789+-_$/!@#%&*()[]")

func newPassword() (string, error) {
	pw := make([]byte, passwordLength)
	_, err := rand.Read(pw)
	if err != nil {
		return "", err
	}
	for i, v := range pw {
		pw[i] = passwordAlphabet[int(v)%len(passwordAlphabet)]
	}

	return string(pw), nil
}

var firstNames = []string{
	"Angela",
	"Carla",
	"Cecilia",
	"Christopher",
	"Darlene",
	"Dominique",
	"Edward",
	"Elliott",
	"Fernando",
	"Jessica",
	"Joanna",
	"John",
	"Nicholas",
	"Percival",
	"Philip",
	"Robert",
	"Theodore",
	"Tyrell",
	"Winston",
}

var lastNames = []string{
	"Alderson",
	"Buckland",
	"Bishop",
	"Cox",
	"Day",
	"DiPierro",
	"Dorian",
	"Espinosa",
	"Kelso",
	"Miller",
	"Moss",
	"Price",
	"Reid",
	"Schmidt",
	"Snowden",
	"Turk",
	"Vera",
	"Wellick",
}

// newFullName returns a random first and last name combination from the name database.
func newFullName() string {
	mathrand.Seed(time.Now().Unix())

	i := mathrand.Intn(len(firstNames))
	j := mathrand.Intn(len(lastNames))
	fn := firstNames[i]
	ln := lastNames[j]

	return fmt.Sprintf("%s %s", fn, ln)
}

// newEmail returns a new random email address.
func newEmail() string {
	return randomdata.Email()
}

func newAddress() *address.Address {
	return &address.Address{
		ID:      uuid.UUID{},
		Street:  randomdata.Street(),
		Zip:     randomdata.PostalCode("US"),
		City:    randomdata.City(),
		Country: "USA",
		Planet:  "Earth",
		User:    nil,
	}
}
