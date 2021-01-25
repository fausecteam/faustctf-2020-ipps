// Package json implements the web service's JSON API.
package json

import (
	"github.com/gorilla/mux"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/credit"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/feedback"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

type Response struct {
	Error  string      `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func AddAPIRoutes(r *mux.Router, as address.Storage, cs credit.Storage, fs feedback.Storage, us user.Storage) {
	h := NewAPIHandler(as, cs, fs, us)

	r.HandleFunc("/login", h.login).Methods("POST")
	r.HandleFunc("/recent-feedback", h.serveRecentFeedback).Methods("GET")

	ur := r.PathPrefix("/user/{user}").Subrouter()
	ur.HandleFunc("/add-address", h.addAddress).Methods("POST")
	ur.HandleFunc("/get-addresses", h.serveAddresses).Methods("GET")
	ur.HandleFunc("/add-credit-card", h.addCreditCard).Methods("POST")
	ur.HandleFunc("/get-credit-cards", h.serveCreditCards).Methods("GET")
}
