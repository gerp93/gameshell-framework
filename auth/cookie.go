package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/helper"
)

const cookieNamePlayerToken string = "CARD-JUDGE-PLAYER-TOKEN"
const cookieNameAccessToken string = "CARD-JUDGE-ACCESS-TOKEN"

func GetCookiePlayerId(r *http.Request) (uuid.UUID, error) {
	cookieValue, err := getCookie(r, cookieNamePlayerToken)
	if err != nil {
		return uuid.Nil, err
	}

	tokenValue, err := getTokenStringValue(cookieValue)
	if err != nil {
		return uuid.Nil, err
	}

	playerId, err := uuid.Parse(tokenValue)
	if err != nil {
		return uuid.Nil, err
	}

	return playerId, nil
}

func SetCookiePlayerId(w http.ResponseWriter, playerId uuid.UUID) error {
	tokenString, err := getValueTokenString(playerId.String())
	if err != nil {
		return err
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
	cookie := getRemovalCookie(cookieNamePlayerToken)
	http.SetCookie(w, &cookie)
}

func HasCookieAccess(r *http.Request, id uuid.UUID) (bool, error) {
	cookieValue, err := getCookie(r, cookieNameAccessToken)
	if err != nil {
		return false, err
	}

	tokenValue, err := getTokenStringValue(cookieValue)
	if err != nil {
		return false, err
	}

	return strings.Contains(tokenValue, id.String()), nil
}

func AddCookieAccessId(w http.ResponseWriter, r *http.Request, id uuid.UUID) error {
	cookieValue, err := getCookie(r, cookieNameAccessToken)
	if err != nil {
		return err
	}

	tokenValue, err := getTokenStringValue(cookieValue)
	if err != nil {
		return err
	}

	if strings.Contains(tokenValue, id.String()) {
		// already have access
		return nil
	}

	var accessStrings []string
	var accessIds []uuid.UUID

	accessStrings = strings.Split(tokenValue, " ")
	accessIds = helper.ConvertArrayStringsToUuids(accessStrings)
	accessIds = append(accessIds, id)
	accessStrings = helper.ConvertArrayUuidsToStrings(accessIds)

	tokenString, err := getValueTokenString(strings.Join(accessStrings, " "))
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:    cookieNameAccessToken,
		Value:   tokenString,
		Path:    "/",
		Expires: time.Now().Add(time.Hour * 12),
	}
	http.SetCookie(w, &cookie)
	return nil
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

func getRemovalCookie(cookieName string) http.Cookie {
	return http.Cookie{
		Name:    cookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
}
