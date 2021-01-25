package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fausecteam/ctf-gameserver/go/checkerlib"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/credit"
)

var (
	ErrRegistrationFailed    = errors.New("could not register user")
	ErrUserAlreadyRegistered = errors.New("a user with that username already exists")
)

type NotLoggedInError struct {
	action string
}

func (err *NotLoggedInError) Error() string {
	return fmt.Sprintf("%s: tried to perform action without being logged in", err.action)
}

// HTTPClient implements communication with the IPPS service.
type HTTPClient struct {
	client     *http.Client
	ip         string
	isLoggedIn bool
	username   string
}

func NewHTTPClient(ip string) (*HTTPClient, error) {
	j, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Jar:     j,
		Timeout: checkerlib.Timeout,
	}

	return &HTTPClient{
		ip:         ip,
		client:     c,
		isLoggedIn: false,
	}, nil

}

func (c *HTTPClient) RegisterUser(username, password, name, email string) error {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	form.Set("password-confirm", password)
	form.Set("name", name)
	form.Set("email", email)
	resp, err := c.client.PostForm(c.urlForPath("signup"), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusServiceUnavailable {
		return context.DeadlineExceeded
	} else if resp.StatusCode != http.StatusOK {
		return ErrRegistrationFailed
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if bytes.Contains(body, []byte("a user with that username already exists")) {
		return ErrUserAlreadyRegistered
	}
	if !bytes.Contains(body, []byte("Your account has been created successfully!")) {
		return ErrRegistrationFailed
	}
	c.isLoggedIn = true
	c.username = username

	return nil
}

var ErrLoginFailed = errors.New("logging into tick user failed")

func (c *HTTPClient) Login(username, password string) error {
	if c.isLoggedIn {
		err := c.Logout()
		if err == ErrLogoutFailed {
			return ErrLoginFailed
		} else if err != nil {
			return err
		}
	}

	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	resp, err := c.client.PostForm(c.urlForPath("login"), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusServiceUnavailable {
		return context.DeadlineExceeded
	} else if resp.StatusCode != http.StatusOK {
		return ErrLoginFailed
	}
	d, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	if d.Find(".alert .alert-danger").Length() > 0 {
		return ErrLoginFailed
	}
	c.isLoggedIn = true
	c.username = username

	return nil
}

var ErrLogoutFailed = errors.New("failed to log user out")

func (c *HTTPClient) Logout() error {
	resp, err := c.client.Get(c.urlForPath("logout"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusServiceUnavailable {
		return context.DeadlineExceeded
	} else if resp.StatusCode != http.StatusOK {
		return ErrLogoutFailed
	}
	c.isLoggedIn = false
	c.username = ""

	return nil
}

func (c *HTTPClient) AddAddress(address *address.Address) error {
	if !c.isLoggedIn {
		return &NotLoggedInError{action: "AddAddress"}
	}

	form := url.Values{}
	form.Set("street", address.Street)
	form.Set("zip", address.Zip)
	form.Set("city", address.City)
	form.Set("country", address.Country)
	form.Set("planet", address.Planet)

	resp, err := c.client.PostForm(c.urlForPath("profile/addresses/add"), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusServiceUnavailable {
		return context.DeadlineExceeded
	} else if resp.StatusCode != http.StatusOK {
		return ErrAddAddressFailed
	}
	d, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	s := d.Find(".alert .alert-danger")
	if s.Length() > 0 {
		return ErrAddAddressFailed
	}

	return nil
}

func (c *HTTPClient) HasAddress(addr *address.Address) (bool, error) {
	if !c.isLoggedIn {
		return false, &NotLoggedInError{action: "HasAddress"}
	}

	resp, err := c.client.Get(c.urlForPath("profile/addresses"))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusServiceUnavailable {
		return false, context.DeadlineExceeded
	} else if resp.StatusCode != http.StatusOK {
		return false, nil
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, err
	}

	found := false
	rows := doc.Find("#addresses tbody tr")
	rows.EachWithBreak(func(i int, row *goquery.Selection) bool {
		cols := row.Find("td")
		if cols.Length() < 5 {
			return true
		}
		a := &address.Address{
			Street:  cols.Nodes[0].FirstChild.Data,
			Zip:     cols.Nodes[1].FirstChild.Data,
			City:    cols.Nodes[2].FirstChild.Data,
			Country: cols.Nodes[3].FirstChild.Data,
			Planet:  cols.Nodes[4].FirstChild.Data,
		}
		if isSameAddress(addr, a) {
			found = true
			return false
		}

		return true
	})

	return found, nil
}

func (c *HTTPClient) AddCreditCard(card *credit.Card) error {
	if !c.isLoggedIn {
		return &NotLoggedInError{action: "AddCreditCard"}
	}

	form := url.Values{}
	form.Set("number", card.Number)
	resp, err := c.client.PostForm(c.urlForPath("profile/add-payment-option"), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ErrAddCreditCardFailed
	}
	d, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	s := d.Find(".alert .alert-danger")
	if s.Length() > 0 {
		return ErrAddCreditCardFailed
	}

	return nil
}

func (c *HTTPClient) HasCreditCard(card *credit.Card) (bool, error) {
	if !c.isLoggedIn {
		return false, &NotLoggedInError{action: "HasCreditCard"}
	}

	resp, err := c.client.Get(c.urlForPath("profile/payment-options"))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusServiceUnavailable {
		return false, context.DeadlineExceeded
	} else if resp.StatusCode != http.StatusOK {
		return false, nil
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, err
	}

	found := false
	rows := doc.Find("td")
	rows.EachWithBreak(func(i int, row *goquery.Selection) bool {
		if row.Text() == html.UnescapeString(card.Number) {
			found = true
			return false
		}

		return true
	})

	return found, nil
}

var (
	ErrPostFeedbackFailed = errors.New("posting feedback failed")
	ErrFeedbackNotFound   = errors.New("requested feedback has not been found")
)

func (c *HTTPClient) PostFeedback(rating int, text string) error {
	if rating < 1 || rating > 5 {
		panic("invalid rating (must be between 1 and 5")
	} else if text == "" {
		panic("feedback text may not be empty")
	}

	if !c.isLoggedIn {
		return &NotLoggedInError{action: "PostFeedback"}
	}

	form := url.Values{}
	form.Set("rating", strconv.Itoa(rating))
	form.Set("text", text)
	resp, err := c.client.PostForm(c.urlForPath("feedback"), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusServiceUnavailable {
		return context.DeadlineExceeded
	} else if resp.StatusCode != http.StatusOK {
		return ErrPostFeedbackFailed
	}
	d, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	alerts := d.Find(".alert .alert-danger")
	if alerts.Length() > 0 {
		return ErrPostFeedbackFailed
	}

	return nil
}

func (c *HTTPClient) HasUserFeedback(username string) (bool, error) {
	resp, err := c.client.Get(c.urlForPath("feedback"))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusServiceUnavailable {
		return false, context.DeadlineExceeded
	} else if resp.StatusCode != http.StatusOK {
		return false, ErrFeedbackNotFound
	}
	d, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, err
	}
	found := false
	posts := d.Find(".customer-feedback")
	posts.Each(func(i int, post *goquery.Selection) {
		author := post.Find("p .author").Text()
		if strings.TrimSpace(author) == username {
			found = true
		}
	})

	return found, nil
}

const (
	servicePort = 8000
	urlTemplate = "http://%s:%d/%s"
)

func (c *HTTPClient) urlForPath(path string) string {
	return fmt.Sprintf(urlTemplate, c.ip, servicePort, path)
}
