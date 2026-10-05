// Package apiPages provides ready-made page handlers for the pages that are
// identical across every game built on this framework: login, admin user
// management, the deck list, the deck password gate, and the account-page
// chrome. Games mount these directly in main.go, the same way they already
// mount apiDeck's CRUD handlers — no game-side wrapper needed. Pages that
// are only *partly* shared (deck detail) live in the game instead, composed
// with the framework's chrome via ParseGameFragment.
package apiPages

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gerp93/gameshell-framework/api"
	apiUser "github.com/gerp93/gameshell-framework/api/user"
	"github.com/gerp93/gameshell-framework/database"
	"github.com/gerp93/gameshell-framework/static"
	"github.com/google/uuid"
)

func parseChrome(bodyName string) (*template.Template, error) {
	return template.ParseFS(static.StaticFiles, "html/pages/base.html", bodyName)
}

func Login(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = basePageData.BrandName + " - Login"

	tmpl, err := parseChrome("html/pages/body/login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	_ = tmpl.ExecuteTemplate(w, "base", basePageData)
}

func Users(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = basePageData.BrandName + " - Users"

	var name string
	var page int
	params := r.URL.Query()
	for key, val := range params {
		switch key {
		case "name":
			name = val[0]
		case "page":
			page, _ = strconv.Atoi(val[0])
		}
	}

	totalRowCount, err := database.CountUsers(name)
	if err != nil {
		totalRowCount = 0
	}
	totalPageCount := max((totalRowCount+9)/10, 1)

	if page < 1 {
		page = 1
	}
	if page > totalPageCount {
		page = totalPageCount
	}

	users, err := database.SearchUsers(name, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get table rows"))
		return
	}

	tmpl, err := parseChrome("html/pages/body/users.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Name     string
		Page     int
		LastPage int
		RowCount int
		Users    []database.User
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Name:         name,
		Page:         page,
		LastPage:     totalPageCount,
		RowCount:     totalRowCount,
		Users:        users,
	})
}

func Decks(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = basePageData.BrandName + " - Decks"

	var name string
	var page int
	params := r.URL.Query()
	for key, val := range params {
		switch key {
		case "name":
			name = val[0]
		case "page":
			page, _ = strconv.Atoi(val[0])
		}
	}

	totalRowCount, err := database.CountDecks(name)
	if err != nil {
		totalRowCount = 0
	}
	totalPageCount := max((totalRowCount+9)/10, 1)

	if page < 1 {
		page = 1
	}
	if page > totalPageCount {
		page = totalPageCount
	}

	decks, err := database.SearchDecks(name, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to get table rows"))
		return
	}

	tmpl, err := parseChrome("html/pages/body/decks.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Name     string
		Page     int
		LastPage int
		RowCount int
		Decks    []database.DeckDetails
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Name:         name,
		Page:         page,
		LastPage:     totalPageCount,
		RowCount:     totalRowCount,
		Decks:        decks,
	})
}

func DeckAccess(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		http.Redirect(w, r, "/decks", http.StatusSeeOther)
		return
	}

	deck, err := database.GetDeck(deckId)
	if err != nil {
		http.Redirect(w, r, "/decks", http.StatusSeeOther)
		return
	}

	if deck.Id == uuid.Nil {
		http.Redirect(w, r, "/decks", http.StatusSeeOther)
		return
	}

	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = basePageData.BrandName + " - Deck"

	hasDeckAccess, err := database.UserHasDeckAccess(basePageData.User.Id, deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to check deck access"))
		return
	}

	if hasDeckAccess {
		http.Redirect(w, r, "/deck/"+deckId.String(), http.StatusSeeOther)
		return
	}

	tmpl, err := parseChrome("html/pages/body/deck-access.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		Deck database.Deck
	}

	_ = tmpl.ExecuteTemplate(w, "base", data{
		BasePageData: basePageData,
		Deck:         deck,
	})
}

func Account(w http.ResponseWriter, r *http.Request) {
	basePageData := api.GetBasePageData(r)
	basePageData.PageTitle = basePageData.BrandName + " - Account"

	tmpl, err := parseChrome("html/pages/body/account.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to parse HTML"))
		return
	}

	type data struct {
		api.BasePageData
		ThemeGroups         []api.ThemeGroup
		ShowWinCelebration  bool
		WinCelebration      database.UserWinCelebration
		ShowLoseCelebration bool
		LoseCelebration     database.UserLoseCelebration
		ShowWinVideo        bool
		WinVideo            database.UserWinVideo
		MaxGifKB            int
	}

	d := data{
		BasePageData:        basePageData,
		ThemeGroups:         api.ThemeGroups,
		ShowWinCelebration:  accountPageFeatures.WinCelebration,
		ShowLoseCelebration: accountPageFeatures.LoseCelebration,
		ShowWinVideo:        accountPageFeatures.WinVideo,
		MaxGifKB:            apiUser.MaxWinGifBytes() / 1024,
	}

	if accountPageFeatures.WinCelebration {
		// Metadata only — the image bytes are served separately by
		// GetWinGif and never loaded into a page render.
		winCelebration, err := database.GetUserWinCelebration(basePageData.User.Id)
		if err == nil {
			d.WinCelebration = winCelebration
		}
	}

	if accountPageFeatures.LoseCelebration {
		loseCelebration, err := database.GetUserLoseCelebration(basePageData.User.Id)
		if err == nil {
			d.LoseCelebration = loseCelebration
		}
	}

	if accountPageFeatures.WinVideo {
		winVideo, err := database.GetUserWinVideo(basePageData.User.Id)
		if err == nil {
			d.WinVideo = winVideo
		}
	}

	_ = tmpl.ExecuteTemplate(w, "base", d)
}
