package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"syscall"
	"time"

	"github.com/fausecteam/ctf-gameserver/go/checkerlib"
	"github.com/google/uuid"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/credit"
)

func main() {
	checkerlib.RunCheck(&checker{})
}

type feedback struct {
	Rating int
	Text   string
}

var feedbacks = []feedback{
	{
		Rating: 1,
		Text:   "Your service is the worst! The content of my package was completely broken!",
	},
	{
		Rating: 1,
		Text: "IPPS has lost two of my packages THIS WEEK!!!",
	},
	{
		Rating: 1,
		Text: `I've tried sending a birthday present to Mars and the shop employee promised
               it would be delivered on time, but after two weeks it showed up on Ganymed instead!`,
	},
	{
		Rating: 2,
		Text: `Well, I sent my package via your "express" delivery, which seems to be quite
		expensive considering that you do not deliver on time. I probably could have saved
		a lot of money if I had just used the regular delivery option and still would have
		received my stuff earlier. Thanks for nothing!`,
	},
	{
		Rating: 2,
		Text: `I am a marine biologist working on Titan. I sent an aquarium containing some
               specimens to my colleagues back to earth via express delivery. Well, it actually
			   did arrive on time, however, as a solid large block of ice. On the bright side,
               the aquarium did in fact arrive in one piece and specimens were still intact for
               examination.`,
	},
	{
		Rating: 3,
		Text:   `I have rather mixed feelings about IPPS! Sometimes it works flawless, but
				 every now and then, one of my packages gets lost or they said that delivery 
				 was attempted, when I was home all day.`,
	},
	{
		Rating: 3,
		Text:  `I bought a spice rack second-hand on eBay, but its missing one screw! So I'm
				only giving you idiots three stars!!!1`,
	},
	{
		Rating: 4,
		Text:   `Most of the time it works as advertised! You should really give it a try.
				 I think out of a hundred deliveries, I had problems only on one or two
				 occasions.`,
	},
	{
		Rating: 5,
		Text: `I live on Ceres and I ordered one pack of batteries, from an online shop located
               on earth, for my nose hair trimmer. You know, so I have a good chance with the
			   ladies. It was not supposed to be delivered for 12 months, but they actually did
			   it in 9! I'm just preparing for my first date without nose hair. Thank you IPPS!`,
	},
	{
		Rating: 5,
		Text:   "Great service! Always on time and the delivery personal is super friendly!",
	},
	{
		Rating: 5,
		Text: `I have been living on Mars for quite some time now and, therefore, had the chance
			   to try different parcel delivery companies. In my experience, IPPS is by far the
			   most reliant and the customer support is always very friendly and tries hard to
			   help you with any problems arising, if something does in fact go wrong. Unlike
			   Planet Express, where the customer support usually simply does not listen at all
			   and always claims it is not their fault. Additionally, they always deliver on time
			   and I have not experienced any adventures that people claim to have experienced on
			   this site.`,
	},
	{
		Rating: 5,
		Text: `For a photography competition back on Earth I took a portrait series of workers
			   on the dark side of the moon with analogue film. IPPS delivered the films to the
			   jury in time and I won the competition! They were shielded from radiation during
			   transit and arrived in perfect condition. Thank you so much IPPS! I've gotta be the
			   happiest guy on Phobos AND Deimos combined!`,

	},
}

// Client is the interface for interacting with IPPS's different APIs.
type Client interface {
	Login(username, password string) error
	AddAddress(address *address.Address) error
	HasAddress(address *address.Address) (bool, error)
	AddCreditCard(card *credit.Card) error
	HasCreditCard(card *credit.Card) (bool, error)
}

var (
	ErrAddAddressFailed    = errors.New("failed to add address to user")
	ErrAddCreditCardFailed = errors.New("failed to add credit card number to account")
)

// Type checker implements the checkerlib's Checker interface
// for the service IPPS.
type checker struct{}

func (ch *checker) PlaceFlag(ip string, team int, tick int) (checkerlib.Result, error) {
	log.Println("PlaceFlag: creating new service client...")
	c, err := NewHTTPClient(ip)
	if err != nil {
		return checkerlib.ResultInvalid, err
	}

	log.Println("PlaceFlag: creating new user...")
	var username string
	password, err := newPassword()
	name := newFullName()
	if err != nil {
		return checkerlib.ResultInvalid, err
	}
	for tries := 0; tries < 5; tries++ {
		username = newUsername()
		email := newEmail()
		err = c.RegisterUser(username, password, name, email)
		if err == ErrUserAlreadyRegistered {
			continue
		}
		if err == ErrRegistrationFailed {
			return checkerlib.ResultFaulty, nil
		} else if err != nil {
			return checkerlib.ResultInvalid, wrapTimeoutErr(err)
		}

		break
	}
	if err == ErrUserAlreadyRegistered {
		log.Printf("PlaceFlag: regular user registration tries exceeded, trying to register a random UUID...")
		id, err := uuid.NewRandom()
		if err != nil {
			return checkerlib.ResultInvalid, err
		}
		username = id.String()
		email := fmt.Sprintf("%s@%s.org", username, username)
		err = c.RegisterUser(username, password, name, email)
		if err == ErrUserAlreadyRegistered || err == ErrRegistrationFailed {
			return checkerlib.ResultFaulty, nil
		} else if err != nil {
			return checkerlib.ResultInvalid, wrapTimeoutErr(err)
		}
	}

	log.Println("PlaceFlag: storing user data in checker database...")
	k := fmt.Sprintf(usernameKeyTemplate, tick)
	checkerlib.StoreState(k, username)
	k = fmt.Sprintf(passwordKeyTemplate, tick)
	checkerlib.StoreState(k, password)

	log.Println("PlaceFlag: adding user feedback...")
	rand.Seed(time.Now().Unix())
	n := rand.Intn(len(feedbacks))
	f := feedbacks[n]
	err = c.PostFeedback(f.Rating, f.Text)
	if err == ErrPostFeedbackFailed {
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		return checkerlib.ResultInvalid, wrapTimeoutErr(err)
	}

	log.Println("PlaceFlag: placing flag...")
	flag := checkerlib.GetFlag(tick, nil)
	card := &credit.Card{Number: flag}
	err = c.AddCreditCard(card)
	if err == ErrAddCreditCardFailed {
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		return checkerlib.ResultInvalid, wrapTimeoutErr(err)
	}

	return checkerlib.ResultOk, nil
}

func (ch *checker) CheckService(ip string, team int) (checkerlib.Result, error) {
	log.Println("Starting Service check...")
	username := newUsername()
	password, err := newPassword()
	if err != nil {
		log.Printf("CheckServie: error generating user password\n")
		return checkerlib.ResultInvalid, err
	}

	httpClient, err := NewHTTPClient(ip)
	if err != nil {
		return checkerlib.ResultInvalid, err
	}
	log.Printf("Trying to register user '%s'\n", username)
	err = httpClient.RegisterUser(username, password, newFullName(), newEmail())
	if err == ErrRegistrationFailed {
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		return checkerlib.ResultInvalid, err
	}
	log.Println("User registration successful")

	log.Println("Starting HTTP API check")
	res, err := checkClient(httpClient, username, password)
	if err == io.EOF {
		log.Println("HTTP: unexpected EOF")
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		log.Printf("HTTP check failed: %v\n", err)
		return res, wrapTimeoutErr(err)
	} else if res != checkerlib.ResultOk {
		log.Println("HTTP check result was not OK")
		return res, nil
	}

	log.Println("Starting JSON API check")
	c, err := NewJSONClient(ip)
	if err == io.EOF {
		log.Println("JSON: unexpected EOF")
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		return checkerlib.ResultInvalid, err
	}
	res, err = checkClient(c, username, password)
	if err != nil {
		log.Printf("JSON API check failed: %v\n", err)
		return res, wrapTimeoutErr(err)
	} else if res != checkerlib.ResultOk {
		log.Println("JSON API check result was not OK")
		return res, nil
	}

	log.Println("Starting GRPC API check")
	grcpClient, err := NewGRPCClient(ip)
	if err != nil {
		return checkerlib.ResultInvalid, wrapTimeoutErr(err)
	}
	res, err = checkClient(grcpClient, username, password)
	if err == io.EOF {
		log.Println("GRPC: unexpected EOF")
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		log.Printf("GRPC API check failed: %v\n", err)
		return res, wrapTimeoutErr(err)
	} else if res != checkerlib.ResultOk {
		log.Println("GRPC API check result was not OK")
		return res, nil
	}
	log.Println("Checking, whether PublicKey of GRPC API is correct")
	ok, err := grcpClient.CheckPublicKey()
	if err != nil {
		return checkerlib.ResultInvalid, wrapTimeoutErr(err)
	}
	if !ok {
		return checkerlib.ResultFaulty, nil
	}

	return checkerlib.ResultOk, nil
}

func checkClient(c Client, username, password string) (checkerlib.Result, error) {
	log.Println("Trying to log in as user...")
	err := c.Login(username, password)
	if err == ErrLoginFailed {
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		return checkerlib.ResultInvalid, err
	}
	log.Println("Successfully logged in as user")

	log.Println("Trying to add an address...")
	a := newAddress()
	err = c.AddAddress(a)
	if err == ErrAddAddressFailed {
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		return checkerlib.ResultInvalid, err
	}
	log.Println("Successfully added Address")
	log.Println("Checking, whether added address is in list of addresses...")
	found, err := c.HasAddress(a)
	if err != nil {
		return checkerlib.ResultInvalid, err
	}
	log.Printf("Result: %t\n", found)
	if !found {
		return checkerlib.ResultFaulty, nil
	}

	log.Println("Trying to add a credit card...")
	id, err := uuid.NewRandom()
	if err != nil {
		return checkerlib.ResultInvalid, nil
	}
	card := &credit.Card{Number: id.String()}
	err = c.AddCreditCard(card)
	if err == ErrAddCreditCardFailed {
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		return checkerlib.ResultInvalid, err
	}
	log.Println("Successfully added credit card.")
	log.Println("Checking, whether added credit card is in list of credit cards...")
	found, err = c.HasCreditCard(card)
	log.Printf("Result: %t\n", found)
	if !found {
		return checkerlib.ResultFaulty, nil
	}

	return checkerlib.ResultOk, nil
}

func (ch *checker) CheckFlag(ip string, team int, tick int) (checkerlib.Result, error) {
	c, err := NewHTTPClient(ip)
	if err != nil {
		return checkerlib.ResultInvalid, err
	}
	u, err := usernameForTick(tick)
	if err != nil {
		return checkerlib.ResultFlagNotFound, nil
	}
	pw, err := passwordForTick(tick)
	if err != nil {
		return checkerlib.ResultFlagNotFound, nil
	}

	// Without the user's feedback, teams cannot find the
	// target username. Therefore, we count teams as faulty
	// that try to "fix" their service by removing author names
	// from user feedback.
	log.Println("CheckFlag: Checking, whether flag user's customer feedback is visible...")
	found, err := c.HasUserFeedback(u)
	if err == ErrFeedbackNotFound {
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		return checkerlib.ResultInvalid, err
	}
	log.Printf("CheckFlag: Result: %t.\n", found)
	if !found {
		return checkerlib.ResultFaulty, nil
	}

	log.Println("CheckFlag: Trying to login as flag user...")
	err = c.Login(u, pw)
	if err == ErrLoginFailed {
		return checkerlib.ResultFaulty, nil
	} else if err != nil {
		return checkerlib.ResultInvalid, err
	}
	log.Println("CheckFlag: Searching for flag on user's payment method page...")
	f := checkerlib.GetFlag(tick, nil)
	card := &credit.Card{Number: f}
	found, err = c.HasCreditCard(card)
	if err != nil {
		return checkerlib.ResultInvalid, err
	}
	log.Printf("CheckFlag: Result: %t.\n", found)
	if !found {
		return checkerlib.ResultFlagNotFound, nil
	}

	return checkerlib.ResultOk, nil
}

const (
	usernameKeyTemplate = "tick-%d-username"
	passwordKeyTemplate = "tick-%d-password"
)

type keyNotFoundError struct {
	key string
}

func (err *keyNotFoundError) Error() string {
	return fmt.Sprintf("key-value-storage: key '%s' does not exist", err.key)
}

func usernameForTick(tick int) (string, error) {
	k := fmt.Sprintf(usernameKeyTemplate, tick)
	username, ok := checkerlib.LoadState(k).(string)
	if !ok {
		return "", &keyNotFoundError{key: k}
	}

	return username, nil

}

func passwordForTick(tick int) (string, error) {
	k := fmt.Sprintf(passwordKeyTemplate, tick)
	password, ok := checkerlib.LoadState(k).(string)
	if !ok {
		return "", &keyNotFoundError{key: k}
	}

	return password, nil
}

func isSameAddress(a *address.Address, b *address.Address) bool {
	return a.Street == b.Street && a.Zip == b.Zip && a.City == b.City && a.Country == b.Country &&
		a.Planet == b.Planet
}

func isSameCreditCard(c *credit.Card, b *credit.Card) bool {
	return c.Number == b.Number
}

type timeout interface {
	Timeout() bool
}

func wrapTimeoutErr(err error) error {
	t, ok := err.(timeout)
	if strings.Contains(err.Error(), "request canceled") ||
			err == context.DeadlineExceeded || (ok && t.Timeout()) {
		return &net.OpError{
			Op:     "write",
			Net:    "tcp",
			Source: nil,
			Addr:   nil,
			Err:    syscall.ETIMEDOUT,
		}
	}
	return err
}