package json

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/user"
)

// loginChecker is the middleware that checks, whether the current
// request is from an authorized user, denying access if that is
// not the case or if the user tries to access data of other users.
func loginChecker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := user.FromContext(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		v := mux.Vars(r)
		vu := v["user"]
		if vu != u.Username {
			http.Error(w, "You are note allowed to access other users' data",
				http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
