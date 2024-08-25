package apiDeck

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
	"github.com/grantfbarnes/card-judge/database"
)

func Access(w http.ResponseWriter, r *http.Request) {
	deckIdString := r.PathValue("deckid")
	deckId, err := uuid.Parse(deckIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed to get deck id"))
		return
	}

	dbcs := database.GetDatabaseConnectionString()
	deck, err := database.GetDeck(dbcs, deckId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed to get deck"))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed to parse form"))
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

	if deck.Password.String != password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("password not valid"))
		return
	}

	err = auth.AddAccessId(w, r, deck.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed to set cookie"))
		return
	}

	w.Header().Add("HX-Refresh", "true")
}

func Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed to parse form"))
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no name found"))
		return
	}

	dbcs := database.GetDatabaseConnectionString()
	id, err := database.CreateDeck(dbcs, name, password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed to create deck"))
		return
	}

	w.Header().Add("HX-Redirect", "/deck/"+id.String())
}
