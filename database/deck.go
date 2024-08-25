package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Deck struct {
	Id           uuid.UUID
	DateAdded    time.Time
	DateModified time.Time

	Name     string
	Password sql.NullString
}

func GetDecks(dbcs string) ([]Deck, error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	selectStatment, err := db.Prepare(`
		SELECT ID
			 , DATE_ADDED
			 , DATE_MODIFIED
			 , NAME
			 , PASSWORD
	 	FROM DECK 
	`)
	if err != nil {
		return nil, err
	}
	defer selectStatment.Close()

	rows, err := selectStatment.Query()
	if err != nil {
		return nil, err
	}

	result := make([]Deck, 0)
	for rows.Next() {
		var deck Deck
		if err := rows.Scan(
			&deck.Id,
			&deck.DateAdded,
			&deck.DateModified,
			&deck.Name,
			&deck.Password); err != nil {
			continue
		}
		result = append(result, deck)
	}
	return result, nil
}

func GetDeck(dbcs string, id uuid.UUID) (Deck, error) {
	var deck Deck

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return deck, err
	}
	defer db.Close()

	selectStatment, err := db.Prepare(`
		SELECT ID
			 , DATE_ADDED
			 , DATE_MODIFIED
			 , NAME
			 , PASSWORD
	 	FROM DECK 
		WHERE ID = ?
	`)
	if err != nil {
		return deck, err
	}
	defer selectStatment.Close()

	rows, err := selectStatment.Query(id)
	if err != nil {
		return deck, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&deck.Id,
			&deck.DateAdded,
			&deck.DateModified,
			&deck.Name,
			&deck.Password); err != nil {
			return deck, err
		}
	}

	return deck, nil
}

func CreateDeck(dbcs string, name string, password string) (uuid.UUID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return id, err
	}

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return id, err
	}
	defer db.Close()

	insertStatment, err := db.Prepare(`
		INSERT INTO DECK (ID, NAME, PASSWORD)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return id, err
	}
	defer insertStatment.Close()

	if password == "" {
		_, err = insertStatment.Exec(id, name, nil)
	} else {
		_, err = insertStatment.Exec(id, name, password)
	}
	if err != nil {
		return id, err
	}

	return id, nil
}
