package api

import (
	"net/http"
	"strings"

	"github.com/grantfbarnes/card-judge/auth"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := auth.GetCookiePlayerId(r)
		loggedIn := err == nil

		// required to be logged in
		if r.URL.Path == "/manage" ||
			strings.HasPrefix(r.URL.Path, "/lobby/") ||
			strings.HasPrefix(r.URL.Path, "/deck/") {
			if !loggedIn {
				auth.SetCookieRedirectURL(w, r.URL.Path)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}

		// required to not be logged in
		if r.URL.Path == "/login" {
			if loggedIn {
				http.Redirect(w, r, auth.GetCookieRedirectURL(r), http.StatusSeeOther)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func WriteGoodHeader(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("HX-Retarget", "find .htmx-result-good")
	writeHeader(w, statusCode, message)
}

func WriteBadHeader(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("HX-Retarget", "find .htmx-result-bad")
	writeHeader(w, statusCode, message)
}

func writeHeader(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
