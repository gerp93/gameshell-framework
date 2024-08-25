package api

import (
	"net/http"

	"github.com/grantfbarnes/card-judge/auth"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := auth.GetPlayerName(r)
		loggedIn := err == nil

		if r.URL.Path == "/login" {
			if loggedIn {
				http.Redirect(w, r, auth.GetRedirectURL(r), http.StatusSeeOther)
				return
			}
		} else {
			if !loggedIn {
				auth.SetRedirectURL(w, r.URL.Path)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			auth.RemoveRedirectURL(w)
		}

		next.ServeHTTP(w, r)
	})
}
