// Package session implements the web application's session management.
package session

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type ctxKey int

const key ctxKey = iota

// Config is the type wrapping a session's configuration options.
type Config struct {
	Name string
	Key  string
}

// NewMiddleware creates and returns a Middleware that stores a user's
// session in the HTTP Request's Context. Additionally, it makes sure
// that the session is saved exactly once, before the the server writes
// its response to the client.
func NewMiddleware(conf *Config) mux.MiddlewareFunc {
	store := sessions.NewCookieStore([]byte(conf.Key))
	store.Options.SameSite = http.SameSiteStrictMode
	store.Options.HttpOnly = true
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s, err := store.Get(r, conf.Name)
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), key, s))
			rw := &responseWriter{
				w: w,
				r: r,
			}
			next.ServeHTTP(rw, r)
			if !rw.saved {
				rw.saveSession()
			}
		})
	}

	return mw
}

type responseWriter struct {
	w     http.ResponseWriter
	r     *http.Request
	saved bool
}

func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.saved {
		w.WriteHeader(http.StatusOK)
	}

	return w.w.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if !w.saved {
		w.saveSession()
	}

	w.w.WriteHeader(statusCode)
}

func (w *responseWriter) saveSession() {
	if w.saved {
		return
	}

	w.saved = true
	s, ok := FromContext(w.r.Context())
	if !ok {
		return
	}
	err := s.Save(w.r, w.w)
	if err != nil {
		// This should never happen, log it just to be safe.
		log.Print(err)
		return
	}
}

// From context returns the session stored in ctx, if any.
func FromContext(ctx context.Context) (*sessions.Session, bool) {
	s, ok := ctx.Value(key).(*sessions.Session)
	return s, ok
}

// MustFromContext is like FromContext, except that it panics
// if no session is stored in ctx.
func MustFromContext(ctx context.Context) *sessions.Session {
	s, ok := FromContext(ctx)
	if !ok {
		panic("session not found in context")
	}

	return s
}
