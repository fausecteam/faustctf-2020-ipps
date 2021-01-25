package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"

	"github.com/fausecteam/ctf-gameserver/go/checkerlib"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/credit"
)

// JSONClient is an implementation of the Client interface that interacts
// with IPPS's JSON API.
type JSONClient struct {
	httpClient *http.Client
	ip         string
	username   string
}

func NewJSONClient(ip string) (*JSONClient, error) {
	j, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Jar:     j,
		Timeout: checkerlib.Timeout,
	}

	return &JSONClient{
		httpClient: c,
		ip:         ip,
	}, nil
}

type loginResponse struct {
	Error    string `json:"error"`
	Username string `json:"result"`
}

func (c *JSONClient) Login(username, password string) error {
	form := map[string]string{
		"username": username,
		"password": password,
	}
	resp, err := c.postMultipartForm(c.getAPIURL("login"), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)
	response := &loginResponse{}
	err = d.Decode(response)
	if err != nil {
		return ErrLoginFailed
	}
	if response.Error != "" {
		log.Printf("login(): returned error message: %s\n", response.Error)
		return ErrLoginFailed
	}
	if response.Username != username {
		return ErrLoginFailed
	}
	c.username = username

	return nil
}

type addAddressResponse struct {
	Error   string           `json:"error"`
	Address *address.Address `json:"result"`
}

func (c *JSONClient) AddAddress(a *address.Address) error {
	if c.username == "" {
		return &NotLoggedInError{action: "AddAddress"}
	}

	form := map[string]string{
		"street":  a.Street,
		"zip":     a.Zip,
		"city":    a.City,
		"country": a.Country,
		"planet":  a.Planet,
	}
	resp, err := c.postMultipartForm(c.getAPIUserURL("add-address"), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	response := &addAddressResponse{}
	err = dec.Decode(response)
	if err != nil || response.Error != "" {
		return ErrAddAddressFailed
	}
	if !isSameAddress(a, response.Address) {
		log.Println("Address returned by API call is not the same as address previously added")
		return ErrAddAddressFailed
	}

	return nil
}

type getAddressesResponse struct {
	Error     string
	Addresses []*address.Address `json:"result"`
}

func (c *JSONClient) HasAddress(a *address.Address) (bool, error) {
	if c.username == "" {
		return false, &NotLoggedInError{action: "HasAddress"}
	}

	resp, err := c.httpClient.Get(c.getAPIUserURL("get-addresses"))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	response := &getAddressesResponse{}
	err = dec.Decode(response)
	if err != nil {
		return false, nil
	}
	for _, ra := range response.Addresses {
		if isSameAddress(a, ra) {
			return true, nil
		}
	}

	return false, nil
}

type addCreditCardResponse struct {
	Error string
	Card  *credit.Card `json:"result"`
}

func (c *JSONClient) AddCreditCard(card *credit.Card) error {
	if c.username == "" {
		return &NotLoggedInError{action: "AddCreditCard"}
	}

	form := map[string]string{"number": card.Number}

	resp, err := c.postMultipartForm(c.getAPIUserURL("add-credit-card"), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	response := &addCreditCardResponse{}
	err = dec.Decode(response)
	if err != nil {
		return ErrAddCreditCardFailed
	}
	if response.Error != "" {
		return ErrAddCreditCardFailed
	}
	if !isSameCreditCard(card, response.Card) {
		return ErrAddCreditCardFailed
	}

	return nil
}

type getCreditCardsResponse struct {
	Error string
	Cards []*credit.Card `json:"result"`
}

func (c *JSONClient) HasCreditCard(card *credit.Card) (bool, error) {
	if c.username == "" {
		return false, &NotLoggedInError{action: "HasCreditCard"}
	}

	resp, err := c.httpClient.Get(c.getAPIUserURL("get-credit-cards"))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	response := &getCreditCardsResponse{}
	err = dec.Decode(response)
	if err != nil {
		return false, nil
	}
	if response.Error != "" {
		log.Printf("Server error response: %s\n", response.Error)
		return false, nil
	}
	for _, rc := range response.Cards {
		if isSameCreditCard(card, rc) {
			return true, nil
		}
	}

	return false, nil
}

func (c *JSONClient) getAPIURL(relpath string) string {
	return fmt.Sprintf("http://%s:%d/api/%s", c.ip, servicePort, relpath)
}

func (c *JSONClient) getAPIUserURL(relpath string) string {
	return fmt.Sprintf("http://%s:%d/api/user/%s/%s",
		c.ip, servicePort, c.username, relpath)
}

func (c *JSONClient) postMultipartForm(url string, values map[string]string) (*http.Response, error) {
	b := bytes.NewBuffer(nil)
	mw := multipart.NewWriter(b)
	for k, v := range values {
		err := mw.WriteField(k, v)
		if err != nil {
			return nil, err
		}
	}
	err := mw.Close()
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", url, b)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := c.httpClient.Do(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
