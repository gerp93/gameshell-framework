package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
	"github.com/grantfbarnes/card-judge/database"
)

type RequestContextKey string

const BasePageDataRequestContextKey RequestContextKey = "basePageDataRequestContextKey"
const PlayerIdRequestContextKey RequestContextKey = "playerIdRequestContextKey"

type BasePageData struct {
	PageTitle string
	Player    database.Player
	LoggedIn  bool
}

func PageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		basePageData := BasePageData{
			PageTitle: "Card Judge",
			Player:    database.Player{},
			LoggedIn:  false,
		}

		playerId, err := auth.GetCookiePlayerId(r)
		if err == nil {
			dbcs := database.GetDatabaseConnectionString()
			player, err := database.GetPlayer(dbcs, playerId)
			if err == nil {
				basePageData.Player = player
				basePageData.LoggedIn = true
			}
		}

		// required to be logged in
		if r.URL.Path == "/manage" ||
			r.URL.Path == "/admin" ||
			r.URL.Path == "/lobbies" ||
			r.URL.Path == "/decks" ||
			strings.HasPrefix(r.URL.Path, "/lobby/") ||
			strings.HasPrefix(r.URL.Path, "/deck/") {
			if !basePageData.LoggedIn {
				auth.SetCookieRedirectURL(w, r.URL.Path)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}

		// required to not be logged in
		if r.URL.Path == "/login" {
			if basePageData.LoggedIn {
				http.Redirect(w, r, auth.GetCookieRedirectURL(r), http.StatusSeeOther)
				return
			}
		}

		// required to be admin
		if r.URL.Path == "/admin" {
			if !basePageData.Player.IsAdmin {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}

		r = r.WithContext(context.WithValue(r.Context(), BasePageDataRequestContextKey, basePageData))

		next.ServeHTTP(w, r)
	})
}

func GetBasePageData(r *http.Request) BasePageData {
	return r.Context().Value(BasePageDataRequestContextKey).(BasePageData)
}

func ApiMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		playerId, _ := auth.GetCookiePlayerId(r)
		r = r.WithContext(context.WithValue(r.Context(), PlayerIdRequestContextKey, playerId))
		next.ServeHTTP(w, r)
	})
}

func GetPlayerId(r *http.Request) uuid.UUID {
	return r.Context().Value(PlayerIdRequestContextKey).(uuid.UUID)
}
