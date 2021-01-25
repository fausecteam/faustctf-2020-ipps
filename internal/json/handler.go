package json

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/internal/session"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/address"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/credit"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/feedback"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

type APIHandler struct {
	as address.Storage
	cs credit.Storage
	fs feedback.Storage
	us user.Storage
}

func NewAPIHandler(as address.Storage, cs credit.Storage, fs feedback.Storage, us user.Storage) *APIHandler {
	return &APIHandler{
		as: as,
		cs: cs,
		fs: fs,
		us: us,
	}
}

func (h *APIHandler) login(w http.ResponseWriter, r *http.Request) {
	u, ok := user.FromContext(r.Context())
	if ok {
		sendResult(w, u.Username)
		return
	}

	err := r.ParseMultipartForm(0)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	lf, err := user.ParseLoginForm(r)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}
	u, err = h.us.ByUsername(lf.Username)
	if err == user.ErrUserNotExists {
		sendError(w, http.StatusBadRequest, err)
		return
	} else if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	if !u.PasswordEquals(lf.Password) {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	sess := session.MustFromContext(r.Context())
	sess.Values["user"] = u.Username
	sendResult(w, u.Username)
}

func (h *APIHandler) serveRecentFeedback(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	os := r.Form.Get("offset")
	offset := uint(0)
	if os != "" {
		pos, err := strconv.ParseUint(os, 10, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		offset = uint(pos)
	}
	ff, err := h.fs.Multiple(20, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResult(w, ff)
}

func (h *APIHandler) addCreditCard(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	u, err := h.us.ByUsername(v["user"])
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	err = r.ParseMultipartForm(0)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}
	c, err := credit.NewCardFromForm(u, r)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}
	err = h.cs.Insert(c)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendResult(w, c)
}

func (h *APIHandler) serveCreditCards(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	u, err := h.us.ByUsername(v["user"])
	if err == user.ErrUserNotExists {
		sendError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	cc, err := h.cs.ByUser(u)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendResult(w, cc)
}

func (h *APIHandler) addAddress(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	u, err := h.us.ByUsername(v["user"])
	if err == user.ErrUserNotExists {
		sendError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	err = r.ParseMultipartForm(0)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	a, err := address.NewFromFormForUser(r, u)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}
	err = h.as.Insert(a)
	if err == address.ErrAddressAlreadyAdded {
		sendError(w, http.StatusBadRequest, err)
		return
	} else if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	a, err = h.as.ByID(a.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendResult(w, a)
}

func (h *APIHandler) serveAddresses(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	u, err := h.us.ByUsername(v["user"])
	if err == user.ErrUserNotExists {
		sendError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	aa, err := h.as.ByUser(u)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendResult(w, aa)
}

func sendResult(w http.ResponseWriter, result interface{}) {
	jw := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	err := jw.Encode(&Response{Result: result})
	if err != nil {
		log.Println(err)
	}
}

func sendError(w http.ResponseWriter, status int, err error) {
	if status == http.StatusInternalServerError {
		log.Println(err)
	}

	jw := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err = jw.Encode(&Response{Error: err.Error()})
	if err != nil {
		log.Println(err)
	}
}
