package http

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/internal/session"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

// authMiddleware returns a middleware that stores the request session's
// user as a User struct in the HTTP request's context
func authMiddleware(us user.Storage) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s := session.MustFromContext(r.Context())
			v, ok := s.Values["user"]
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			username, ok := v.(string)
			if !ok {
				http.Error(w, "Session cookie is corrupt", http.StatusBadRequest)
				return
			}
			u, err := us.ByUsername(username)
			if err != nil {
				log.Print(err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			r = r.WithContext(user.NewContext(r.Context(), u))
			next.ServeHTTP(w, r)
		})
	}
}

// loginChecker is the middleware that checks, whether the current
// request is from an authorized user, sending a redirect to the login
// page, if that is not the case
func loginChecker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := user.FromContext(r.Context())
		if !ok {
			sess := session.MustFromContext(r.Context())
			sess.AddFlash("You must be logged in in order to view this page", "errors")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
