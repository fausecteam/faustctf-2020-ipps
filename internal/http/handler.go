package http

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/mail"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/internal/session"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/credit"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/feedback"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/parcel"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

func init() {
	// Gorilla sessions gob encodes all data structures, so
	gob.Register(&mail.Address{})
}

type templateHandler struct {
	templates *template.Template
	name      string
	title     string
}

type Page struct {
	Title   string
	Success string
	Errors  []string
	User    *user.User
}

func NewPage(title string, r *http.Request) *Page {
	s := session.MustFromContext(r.Context())
	u, ok := user.FromContext(r.Context())
	if !ok {
		u = nil
	}

	return &Page{
		Title:   title,
		Success: sessionMessage(s, "success"),
		Errors:  sessionMessages(s, "errors"),
		User:    u,
	}
}

func newTemplateHandler(templates *template.Template, name, title string) *templateHandler {
	return &templateHandler{
		templates: templates,
		name:      name,
		title:     title,
	}
}

func (th *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := th.templates.ExecuteTemplate(w, th.name, NewPage(th.title, r))
	if err != nil {
		log.Print(err)
		return
	}
}

func loginFormHandler(t *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, isloggedIn := user.FromContext(r.Context())
		if isloggedIn {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		s := session.MustFromContext(r.Context())
		p := &Page{
			Success: sessionMessage(s, "success"),
			Errors:  sessionMessages(s, "errors"),
		}
		err := t.ExecuteTemplate(w, "login.html", p)
		if err != nil {
			log.Print(err)
			return
		}
	}
}

func registerFormHandler(t *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, isloggedIn := user.FromContext(r.Context())
		if isloggedIn {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		s := session.MustFromContext(r.Context())
		p := &Page{
			Success: sessionMessage(s, "success"),
			Errors:  sessionMessages(s, "errors"),
		}
		err := t.ExecuteTemplate(w, "register.html", p)
		if err != nil {
			log.Print(err)
			return
		}
	}
}

func sessionMessages(s *sessions.Session, key string) []string {
	ff := s.Flashes(key)
	mm := make([]string, 0, len(ff))
	for _, f := range ff {
		str, ok := f.(string)
		if !ok {
			continue
		}
		mm = append(mm, str)
	}

	return mm
}

func sessionMessage(s *sessions.Session, key string) string {
	ff := s.Flashes(key)
	if len(ff) == 0 {
		return ""
	}
	f, ok := ff[0].(string)
	if !ok {
		return ""
	}

	return f
}

func newFileServer(root, prefixToStrip string) http.Handler {
	fs := http.FileServer(http.Dir(root))
	if prefixToStrip != "" {
		fs = http.StripPrefix(prefixToStrip, fs)
	}

	return fs
}

type loginHandler struct {
	UserStorage user.Storage
}

func (h *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess := session.MustFromContext(r.Context())
	f, err := user.ParseLoginForm(r)
	if err != nil {
		log.Print(err)
		sess.AddFlash(err.Error(), "errors")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	u, err := h.UserStorage.ByUsername(f.Username)
	if err == user.ErrUserNotExists {
		sess.AddFlash("User does not exist or password is wrong!", "errors")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	} else if err != nil {
		log.Print(err)
		sess.AddFlash(err.Error(), "errors")
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	if !u.PasswordEquals(f.Password) {
		sess.AddFlash("User does not exist or password is wrong!", "errors")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	sess.Values["user"] = u.Username
	http.Redirect(w, r, "/", http.StatusFound)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	sess := session.MustFromContext(r.Context())
	if _, ok := sess.Values["user"]; !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	delete(sess.Values, "user")
	http.Redirect(w, r, "/", http.StatusFound)
}

type registerHandler struct {
	UserStorage user.Storage
}

func (h *registerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess := session.MustFromContext(r.Context())
	u, err := user.ParseRegistrationForm(r)
	if err != nil {
		sess.AddFlash(err.Error(), "errors")
		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}
	err = h.UserStorage.Insert(u)
	if err != nil {
		sess.AddFlash(err.Error(), "errors")
		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	sess.AddFlash("Your account has been created successfully!", "success")
	sess.Values["user"] = u.Username
	http.Redirect(w, r, "/", http.StatusFound)
}

type updateProfileHandler struct {
	UserStorage user.Storage
}

func (uph *updateProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess := session.MustFromContext(r.Context())
	u := user.MustFromContext(r.Context())
	u, err := uph.UserStorage.ByID(u.ID)
	if err != nil {
		log.Print(err)
		sess.AddFlash(http.StatusText(http.StatusInternalServerError), "errors")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}

	err = user.UpdateFromEditForm(u, r)
	if err != nil {
		sess.AddFlash(err.Error(), "errors")
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}
	err = uph.UserStorage.Update(u)
	if err != nil {
		log.Print(err)
		sess.AddFlash(http.StatusText(http.StatusInternalServerError), "errors")
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}

	sess.AddFlash("Your profile has been updated successfully!", "success")
	sess.Values["user"] = u.Email
	http.Redirect(w, r, "/profile", http.StatusFound)
}

type paymentOptionsHandler struct {
	Templates   *template.Template
	CardStorage credit.Storage
}

type paymentPage struct {
	*Page
	Cards []*credit.Card
}

func (ph *paymentOptionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := user.MustFromContext(r.Context())
	cc, err := ph.CardStorage.ByUser(u)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p := &paymentPage{
		Page:  NewPage("Payment Options", r),
		Cards: cc,
	}
	err = ph.Templates.ExecuteTemplate(w, "payment_options.html", p)
	if err != nil {
		log.Print(err)
		return
	}
}

type addPaymantOptionHandler struct {
	CardStorage credit.Storage
}

func (ph *addPaymantOptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess := session.MustFromContext(r.Context())
	u := user.MustFromContext(r.Context())
	c, err := credit.NewCardFromForm(u, r)
	if err != nil {
		log.Print(err)
		sess.AddFlash(err.Error(), "errors")
		http.Redirect(w, r, "/profile/payment-options", http.StatusFound)
		return
	}
	err = ph.CardStorage.Insert(c)
	if err != nil {
		log.Print(err)
		sess.AddFlash(http.StatusText(http.StatusInternalServerError), "errors")
		http.Redirect(w, r, "/profile/payment-options", http.StatusFound)
		return
	}

	sess.AddFlash("Your credit card has been added successfully!", "success")
	http.Redirect(w, r, "/profile/payment-options", http.StatusFound)
}

type feedbackPage struct {
	*Page
	Feedbacks []feedback.Feedback
}

type feedbackHandler struct {
	Templates *template.Template
	Storage   feedback.Storage
}

func (fh *feedbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ostr := r.PostForm.Get("offset")
	offset, err := strconv.ParseUint(ostr, 10, 64)
	if err != nil {
		offset = 0
	}
	ff, err := fh.Storage.Multiple(20, uint(offset))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	p := &feedbackPage{
		Page:      NewPage("Feedback", r),
		Feedbacks: ff,
	}
	err = fh.Templates.ExecuteTemplate(w, "feedback.html", p)
	if err != nil {
		log.Print(err)
		return
	}
}

type addFeedbackHandler struct {
	Storage feedback.Storage
}

func (fh *addFeedbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess := session.MustFromContext(r.Context())
	u := user.MustFromContext(r.Context())

	err := r.ParseForm()
	if err != nil {
		sess.AddFlash(http.StatusText(http.StatusInternalServerError), "errors")
		http.Redirect(w, r, "/feedback", http.StatusFound)
		return
	}
	rating, err := strconv.ParseUint(r.PostForm.Get("rating"), 10, 8)
	if err != nil {
		sess.AddFlash("Rating must be a number between 1 and 5", "errors")
		http.Redirect(w, r, "/feedback", http.StatusFound)
		return
	}
	f, err := feedback.New(u.Username, uint8(rating), r.PostForm.Get("text"))
	if err != nil {
		sess.AddFlash(err.Error(), "errors")
		http.Redirect(w, r, "/feedback", http.StatusFound)
		return
	}
	err = fh.Storage.Insert(f)
	if err != nil {
		sess.AddFlash(err.Error(), "errors")
		http.Redirect(w, r, "/feedback", http.StatusFound)
		return
	}

	sess.AddFlash("Thank you, for your feedback!", "success")
	http.Redirect(w, r, "/feedback", http.StatusFound)
}

type addressPage struct {
	*Page
	Addresses []*address.Address
}

type addressHandler struct {
	Templates *template.Template
	Storage   address.Storage
}

func (h *addressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := user.MustFromContext(r.Context())
	aa, err := h.Storage.ByUser(u)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	p := &addressPage{
		Page:      NewPage("Addresses", r),
		Addresses: aa,
	}
	err = h.Templates.ExecuteTemplate(w, "addresses.html", p)
	if err != nil {
		log.Println(err)
	}
}

type addAddressHandler struct {
	Storage address.Storage
}

func (h *addAddressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := user.MustFromContext(r.Context())
	sess := session.MustFromContext(r.Context())
	a, err := address.NewFromFormForUser(r, u)
	if err != nil {
		sess.AddFlash(err.Error(), "errors")
		http.Redirect(w, r, "/profile/addresses", http.StatusFound)
		return
	}

	err = h.Storage.Insert(a)
	if err != nil {
		if err == address.ErrAddressAlreadyAdded {
			sess.AddFlash("You have already added this address.", "errors")

		} else {
			sess.AddFlash(err.Error(), "errors")
		}
		http.Redirect(w, r, "/profile/addresses", http.StatusFound)
		return
	}

	sess.AddFlash("Your address has been added successfully!", "success")
	http.Redirect(w, r, "/profile/addresses", http.StatusFound)
}

type findParcelHandler struct {
	Storage parcel.Storage
}

func (h *findParcelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess := session.MustFromContext(r.Context())
	err := r.ParseForm()
	if err != nil {
		sess.AddFlash(err.Error())
		http.Redirect(w, r, "/tracking", http.StatusFound)
	}
	ids, ok := r.PostForm["tracking-id"]
	if !ok || len(ids) < 1 {
		sess.AddFlash("The tracking number is missing", "errors")
		http.Redirect(w, r, "/tracking", http.StatusFound)
		return
	}

	id, err := uuid.Parse(ids[0])
	if err != nil {
		sess.AddFlash("The tracking number you provided is invalid", "errors")
		http.Redirect(w, r, "/tracking", http.StatusFound)
		return
	}
	p, err := h.Storage.ByID(id)
	if err != nil {
		sess.AddFlash("An internal server error occured, please try again later", "errors")
		http.Redirect(w, r, "/tracking", http.StatusFound)
		return
	} else if p == nil {
		sess.AddFlash("A parcel with that tracking number does not exist in our database."+
			" Please make sure that the tracking number you entered is correct or"+
			" contact the sender.", "warnings")
		http.Redirect(w, r, "/tracking", http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/tracking/%s", id.String()), http.StatusFound)
}

type trackingHandler struct {
	templates     *template.Template
	eventStorage  parcel.EventAccesser
	parcelStorage parcel.Accesser
}

type trackingPage struct {
	*Page
	Parcel *parcel.Parcel
	Events []*parcel.Event
}

func (h *trackingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess := session.MustFromContext(r.Context())
	v := mux.Vars(r)
	idStr, ok := v["id"]
	if !ok {
		sess.AddFlash("The tracking number is missing", "errors")
		http.Redirect(w, r, "/tracking", http.StatusFound)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		sess.AddFlash("The tracking number you supplied is invalid", "errors")
		http.Redirect(w, r, "/tracking", http.StatusFound)
		return
	}

	p, err := h.parcelStorage.ByID(id)
	if err != nil {
		log.Println(err)
		sess.AddFlash("An internal server error occured, please try again later", "errors")
		http.Redirect(w, r, "/tracking", http.StatusFound)
		return
	} else if p == nil {
		sess.AddFlash("A parcel with that tracking number does not exist in our database."+
			" Please make sure that the tracking number you entered is correct or"+
			" contact the sender.", "warnings")
		http.Redirect(w, r, "/tracking", http.StatusFound)
		return
	}

	ee, err := h.eventStorage.ByParcel(p)
	if err != nil {
		log.Println(err)
		sess.AddFlash("An internal server error occurred, please try again later",
			"errors")
		http.Redirect(w, r, "/tracking", http.StatusFound)
		return
	}

	err = h.templates.ExecuteTemplate(w, "parcel_events.html", &trackingPage{
		Page:   NewPage("Tracking Information", r),
		Parcel: p,
		Events: ee,
	})
	if err != nil {
		log.Println(err)
	}
}
