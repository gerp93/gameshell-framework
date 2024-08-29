package apiDeck

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/api"
	"github.com/grantfbarnes/card-judge/auth"
	"github.com/grantfbarnes/card-judge/database"
)

func Access(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := uuid.Parse(idString)
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to get id from path.")
		return
	}

	dbcs := database.GetDatabaseConnectionString()
	deck, err := database.GetDeck(dbcs, id)
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to get deck from database.")
		return
	}

	err = r.ParseForm()
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to parse form.")
		return
	}

	var password string
	for key, val := range r.Form {
		if key != "password" {
			continue
		}
		password = val[0]
		break
	}

	if !auth.PasswordMatchesHash(password, deck.PasswordHash.String) {
		api.WriteBadHeader(w, http.StatusBadRequest, "Provided password is not valid.")
		return
	}

	err = auth.AddAccessId(w, r, deck.Id)
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to set cookie in browser.")
		return
	}

	w.Header().Add("HX-Refresh", "true")
	api.WriteGoodHeader(w, http.StatusCreated, "Success")
}

func Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to parse form.")
		return
	}

	var name string
	var password string
	for key, val := range r.Form {
		if key == "name" {
			name = val[0]
		} else if key == "password" {
			password = val[0]
		}
	}

	if name == "" {
		api.WriteBadHeader(w, http.StatusBadRequest, "No name found.")
		return
	}

	dbcs := database.GetDatabaseConnectionString()
	id, err := database.CreateDeck(dbcs, name, password)
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to create deck in database.")
		return
	}

	auth.AddAccessId(w, r, id)

	w.Header().Add("HX-Redirect", "/deck/"+id.String())
	api.WriteGoodHeader(w, http.StatusCreated, "Success")
}

func Update(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := uuid.Parse(idString)
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to get id from path.")
		return
	}

	err = r.ParseForm()
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to parse form.")
		return
	}

	var name string
	var password string
	for key, val := range r.Form {
		if key == "newDeckName" {
			name = val[0]
		} else if key == "newDeckPassword" {
			password = val[0]
		}
	}

	if name == "" {
		api.WriteBadHeader(w, http.StatusBadRequest, "No name found.")
		return
	}

	dbcs := database.GetDatabaseConnectionString()
	err = database.UpdateDeck(dbcs, id, name, password)
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to update deck in database.")
		return
	}

	auth.AddAccessId(w, r, id)

	w.Header().Add("HX-Refresh", "true")
	api.WriteGoodHeader(w, http.StatusCreated, "Success")
}

func Delete(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := uuid.Parse(idString)
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to get id from path.")
		return
	}

	dbcs := database.GetDatabaseConnectionString()
	err = database.DeleteDeck(dbcs, id)
	if err != nil {
		api.WriteBadHeader(w, http.StatusBadRequest, "Failed to delete deck in database.")
		return
	}

	w.Header().Add("HX-Redirect", "/decks")
	api.WriteGoodHeader(w, http.StatusCreated, "Success")
}
