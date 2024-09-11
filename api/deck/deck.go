package apiDeck

import (
	"net/http"
	"text/template"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/api"
	"github.com/grantfbarnes/card-judge/database"
)

func Search(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse form."))
		return
	}

	var search string
	for key, val := range r.Form {
		if key == "search" {
			search = val[0]
		}
	}

	search = "%" + search + "%"

	decks, err := database.GetDecks(search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/components/table-rows/deck-table-rows.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to parse HTML."))
		return
	}

	tmpl.ExecuteTemplate(w, "deck-table-rows", decks)
}

func Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse form."))
		return
	}

	var name string
	var password string
	var passwordConfirm string
	for key, val := range r.Form {
		if key == "name" {
			name = val[0]
		} else if key == "password" {
			password = val[0]
		} else if key == "passwordConfirm" {
			passwordConfirm = val[0]
		}
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No name found."))
		return
	}

	if password != "" {
		if password != passwordConfirm {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Passwords do not match."))
			return
		}
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get user id."))
		return
	}

	existingDeckId, err := database.GetDeckId(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if existingDeckId != uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Deck name already exists."))
		return
	}

	id, err := database.CreateDeck(name, password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = database.AddUserDeckAccess(userId, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Redirect", "/deck/"+id.String())
	w.WriteHeader(http.StatusCreated)
}

func SetName(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get deck id from path."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get user id."))
		return
	}

	if !database.HasDeckAccess(userId, deckId) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User does not have access."))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse form."))
		return
	}

	var name string
	for key, val := range r.Form {
		if key == "name" {
			name = val[0]
		}
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No name found."))
		return
	}

	existingDeckId, err := database.GetDeckId(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if existingDeckId != uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Deck name already exists."))
		return
	}

	err = database.SetDeckName(deckId, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func SetPassword(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get deck id from path."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get user id."))
		return
	}

	if !database.HasDeckAccess(userId, deckId) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User does not have access."))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse form."))
		return
	}

	var password string
	var passwordConfirm string
	for key, val := range r.Form {
		if key == "password" {
			password = val[0]
		} else if key == "passwordConfirm" {
			passwordConfirm = val[0]
		}
	}

	if password != "" {
		if password != passwordConfirm {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Passwords do not match."))
			return
		}
	}

	err = database.SetDeckPassword(deckId, password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckId")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get deck id from path."))
		return
	}

	userId := api.GetUserId(r)
	if userId == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to get user id."))
		return
	}

	if !database.HasDeckAccess(userId, deckId) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User does not have access."))
		return
	}

	err = database.DeleteDeck(deckId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Redirect", "/decks")
	w.WriteHeader(http.StatusOK)
}
