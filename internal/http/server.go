// Package server contains Webfoo's web server implementation
package http

import (
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/internal/json"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/internal/session"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/credit"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/feedback"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/parcel"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

type Config struct {
	Address      string
	ReadTimeout  Duration
	WriteTimeout Duration
}

// Duration is an implementation of the TextMarshaler and TextUnmarshaler
// interfaces. It wraps the standard library's duration type, so durations may
// be parsed by the TOML parser.
type Duration struct {
	duration time.Duration
}

func (d *Duration) MarshalText() ([]byte, error) {
	return []byte(d.duration.String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.duration, err = time.ParseDuration(string(text))
	if err != nil {
		return err
	}

	return nil
}

type Server struct {
	AddressStorage  address.Storage
	CreditStorage   credit.Storage
	EventStorage    parcel.EventStorage
	FeedbackStorage feedback.Storage
	ParcelStorage   parcel.Storage
	UserStorage     user.Storage
}

func (s *Server) ListenAndServe(config *Config, sessionConfig *session.Config) error {
	r, err := s.newRouter(sessionConfig)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:         config.Address,
		Handler:      r,
		ReadTimeout:  config.ReadTimeout.duration,
		WriteTimeout: config.WriteTimeout.duration,
	}

	return srv.ListenAndServe()
}

func (s *Server) newRouter(sessionConfig *session.Config) (*mux.Router, error) {
	t, err := template.ParseGlob("web/template/*.html")
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	r.Use(session.NewMiddleware(sessionConfig), authMiddleware(s.UserStorage))
	r.Handle("/", newTemplateHandler(t, "home.html", "Home"))
	r.PathPrefix("/static/").Handler(newFileServer("web/static/", "/static/"))
	r.Handle("/login", loginFormHandler(t)).Methods("GET")
	r.Handle("/login", &loginHandler{UserStorage: s.UserStorage}).Methods("POST")
	r.HandleFunc("/logout", handleLogout)
	r.Handle("/signup", registerFormHandler(t)).Methods("GET")
	r.Handle("/signup", &registerHandler{UserStorage: s.UserStorage}).Methods("POST")
	r.Handle("/feedback", &feedbackHandler{
		Templates: t,
		Storage:   s.FeedbackStorage,
	}).Methods("GET")
	r.Handle("/tracking",
		newTemplateHandler(t, "tracking.html", "Tracking")).Methods("GET")
	r.Handle("/tracking", &findParcelHandler{Storage: s.ParcelStorage}).Methods("POST")
	r.Handle("/tracking/{id}", &trackingHandler{parcelStorage: nil})
	r.Handle("/feedback", &addFeedbackHandler{
		Storage: s.FeedbackStorage,
	}).Methods("POST")

	pr := r.PathPrefix("/profile").Subrouter()
	pr.Use(loginChecker)
	pr.Handle("", newTemplateHandler(t, "profile.html", "Profile"))
	pr.Handle("/update", &updateProfileHandler{UserStorage: s.UserStorage})
	pr.Handle("/addresses", &addressHandler{Templates: t, Storage: s.AddressStorage})
	pr.Handle("/addresses/add", &addAddressHandler{Storage: s.AddressStorage})
	pr.Handle("/payment-options", &paymentOptionsHandler{CardStorage: s.CreditStorage, Templates: t})
	pr.Handle("/add-payment-option", &addPaymantOptionHandler{CardStorage: s.CreditStorage}).
		Methods("POST")

	ar := r.PathPrefix("/api").Subrouter()
	json.AddAPIRoutes(ar, s.AddressStorage, s.CreditStorage, s.FeedbackStorage, s.UserStorage)

	return r, nil
}
