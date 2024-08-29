package auth

import (
	"errors"
	"net/http"
	"time"
)

const cookieNamePlayerToken string = "CARD-JUDGE-PLAYER-TOKEN"
const cookieNameAccessToken string = "CARD-JUDGE-ACCESS-TOKEN"

func getCookie(r *http.Request, cookieName string) (string, error) {
	cookieFound := false
	cookieValue := ""
	for _, c := range r.Cookies() {
		if c.Name != cookieName {
			continue
		}
		cookieFound = true
		cookieValue = c.Value
		break
	}

	if !cookieFound {
		return "", errors.New("cookie not found")
	}

	return cookieValue, nil
}

func getRemovalCookie(cookieName string) http.Cookie {
	return http.Cookie{
		Name:    cookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
}
