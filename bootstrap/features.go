package bootstrap

import (
	"net/http"

	"github.com/gerp93/gameshell-framework/api"
	apiDeck "github.com/gerp93/gameshell-framework/api/deck"
	apiLobby "github.com/gerp93/gameshell-framework/api/lobby"
	apiPages "github.com/gerp93/gameshell-framework/api/pages"
	apiUser "github.com/gerp93/gameshell-framework/api/user"
	"github.com/gerp93/gameshell-framework/static"
)

// Features is the single, readable list of framework-provided
// functionality a game can turn on. Each field's doc comment names
// exactly what mounting it adds. A game wanting a different limit/behavior
// within an enabled feature still uses that feature's own setter (e.g.
// apiUser.SetMaxWinGifBytes) — Features only controls whether the feature
// exists at all, not how it's configured.
type Features struct {
	// Decks mounts deck CRUD (create/rename/password/visibility/delete) and
	// the /decks list + /deck/{deckId}/access gate pages. The deck *detail*
	// page (GET /deck/{deckId}) stays game-owned regardless — it needs each
	// game's own card schema — so a deckless game still skips wiring that
	// route itself; this flag only covers the framework's half.
	Decks bool
	// WinCelebration mounts the win-gif/win-message routes and enables the
	// account page's Win Celebration section (equivalent to calling
	// apiPages.SetAccountPageFeatures(AccountPageFeatures{WinCelebration: true})).
	WinCelebration bool
	// LobbyTurnTimer mounts PUT /api/lobby/{lobbyId}/turn-timer. Games that
	// have their own timer concept (e.g. card-judge's round timer) leave
	// this off rather than expose a second, competing timer control.
	LobbyTurnTimer bool
}

// MountFeatures wires the framework's core user/account/auth routes
// (always on — every game needs accounts) plus whichever optional routes f
// enables, on the default ServeMux. Call once at startup, before Serve.
func MountFeatures(f Features) {
	http.Handle("POST /api/user/create", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.Create)))
	http.Handle("POST /api/user/create/admin", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.CreateAdmin)))
	http.Handle("POST /api/user/login", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.Login)))
	http.Handle("POST /api/user/logout", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.Logout)))
	http.Handle("PUT /api/user/{userId}/name", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.SetName)))
	http.Handle("PUT /api/user/{userId}/password", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.SetPassword)))
	http.Handle("PUT /api/user/{userId}/password/reset", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.ResetPassword)))
	http.Handle("PUT /api/user/{userId}/color-theme", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.SetColorTheme)))
	http.Handle("PUT /api/user/{userId}/approve", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.Approve)))
	http.Handle("PUT /api/user/{userId}/is-admin", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.SetIsAdmin)))
	http.Handle("DELETE /api/user/{userId}", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.Delete)))

	http.Handle("GET /login", api.MiddlewareForPages(http.HandlerFunc(apiPages.Login)))
	http.Handle("GET /account", api.MiddlewareForPages(http.HandlerFunc(apiPages.Account)))
	http.Handle("GET /users", api.MiddlewareForPages(http.HandlerFunc(apiPages.Users)))

	if f.Decks {
		http.Handle("POST /api/deck/create", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.Create)))
		http.Handle("PUT /api/deck/{deckId}/name", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.SetName)))
		http.Handle("PUT /api/deck/{deckId}/password", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.SetPassword)))
		http.Handle("PUT /api/deck/{deckId}/is-public-read-only", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.SetIsPublicReadOnly)))
		http.Handle("DELETE /api/deck/{deckId}", api.MiddlewareForAPIs(http.HandlerFunc(apiDeck.Delete)))
		http.Handle("GET /decks", api.MiddlewareForPages(http.HandlerFunc(apiPages.Decks)))
		http.Handle("GET /deck/{deckId}/access", api.MiddlewareForPages(http.HandlerFunc(apiPages.DeckAccess)))
	}

	if f.WinCelebration {
		http.Handle("PUT /api/user/{userId}/win-gif", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.SetWinGif)))
		http.Handle("DELETE /api/user/{userId}/win-gif", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.ClearWinGif)))
		http.Handle("GET /api/user/{userId}/win-gif", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.GetWinGif)))
		http.Handle("PUT /api/user/{userId}/win-message", api.MiddlewareForAPIs(http.HandlerFunc(apiUser.SetWinMessage)))
		apiPages.SetAccountPageFeatures(apiPages.AccountPageFeatures{WinCelebration: true})
	}

	if f.LobbyTurnTimer {
		http.Handle("PUT /api/lobby/{lobbyId}/turn-timer", api.MiddlewareForAPIs(http.HandlerFunc(apiLobby.SetTurnTimer)))
	}
}

// ApplyFeatureSchema applies the SQL for whichever optional features f
// enables — currently just static.DeckSQLFiles for f.Decks. Call after
// ApplySchema(static.StaticFiles, static.SQLFiles) (the core schema) and
// before the game's own ApplySchema call, same DB-connect-first ordering.
// A deckless game (Decks: false) never gets DECK/USER_ACCESS_DECK/
// AUDIT_DECK created at all, not just unused — same "doesn't exist unless
// asked for" guarantee MountFeatures gives the routes.
func ApplyFeatureSchema(f Features) {
	if f.Decks {
		ApplySchema(static.StaticFiles, static.DeckSQLFiles)
	}
}
