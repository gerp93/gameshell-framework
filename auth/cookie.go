package auth

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const cookieNamePlayerToken string = "CARD-JUDGE-PLAYER-TOKEN"
const cookieNameRedirectURL string = "CARD-JUDGE-REDIRECT-URL"

func GetCookiePlayerId(r *http.Request) (uuid.UUID, error) {
	cookieValue, err := getCookie(r, cookieNamePlayerToken)
	if err != nil {
		log.Println(err)
		return uuid.Nil, errors.New("failed to get cookie")
	}

	tokenValue, err := getTokenStringValue(cookieValue)
	if err != nil {
		log.Println(err)
		return uuid.Nil, errors.New("failed to get token")
	}

	playerId, err := uuid.Parse(tokenValue)
	if err != nil {
		log.Println(err)
		return uuid.Nil, errors.New("failed to parse token")
	}

	return playerId, nil
}

func SetCookiePlayerId(w http.ResponseWriter, playerId uuid.UUID) error {
	tokenString, err := getValueTokenString(playerId.String())
	if err != nil {
		log.Println(err)
		return errors.New("failed to create token")
	}

	cookie := http.Cookie{
		Name:    cookieNamePlayerToken,
		Value:   tokenString,
		Path:    "/",
		Expires: time.Now().Add(time.Hour * 12),
	}
	http.SetCookie(w, &cookie)
	return nil
}

func RemoveCookiePlayerId(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:    cookieNamePlayerToken,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(w, &cookie)
}

func GetCookieRedirectURL(r *http.Request) string {
	redirectPath, err := getCookie(r, cookieNameRedirectURL)
	if err != nil {
		return "/"
	}
	return redirectPath
}

func SetCookieRedirectURL(w http.ResponseWriter, url string) {
	cookie := http.Cookie{
		Name:    cookieNameRedirectURL,
		Value:   url,
		Path:    "/",
		Expires: time.Now().Add(time.Hour * 12),
	}
	http.SetCookie(w, &cookie)
}

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
