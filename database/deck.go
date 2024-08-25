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
