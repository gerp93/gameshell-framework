package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

func GetDatabaseConnectionString() string {
	serverHost := os.Getenv("GFB_SQL_HOST")
	userName := os.Getenv("GFB_SQL_USER")
	userPassword := os.Getenv("GFB_SQL_PASSWORD")
	databaseName := os.Getenv("GFB_SQL_DATABASE")
	return fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", userName, userPassword, serverHost, databaseName)
}

func Ping(dbcs string) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}

func GetLobbies(dbcs string) ([]Lobby, error) {
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
	 	FROM LOBBY 
	`)
	if err != nil {
		return nil, err
	}
	defer selectStatment.Close()

	rows, err := selectStatment.Query()
	if err != nil {
		return nil, err
	}

	result := make([]Lobby, 0)
	for rows.Next() {
		var lobby Lobby
		if err := rows.Scan(
			&lobby.Id,
			&lobby.DateAdded,
			&lobby.DateModified,
			&lobby.Name,
			&lobby.Password); err != nil {
			continue
		}
		result = append(result, lobby)
	}
	return result, nil
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

func GetCards(dbcs string, deckId uuid.UUID) ([]Card, error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	selectStatment, err := db.Prepare(`
		SELECT ID
			 , DATE_ADDED
			 , DATE_MODIFIED
			 , TYPE
			 , TEXT
	 	FROM CARD 
		WHERE DECK_ID = ?
	`)
	if err != nil {
		return nil, err
	}
	defer selectStatment.Close()

	rows, err := selectStatment.Query(deckId)
	if err != nil {
		return nil, err
	}

	result := make([]Card, 0)
	for rows.Next() {
		var card Card
		if err := rows.Scan(
			&card.Id,
			&card.DateAdded,
			&card.DateModified,
			&card.Type,
			&card.Text); err != nil {
			continue
		}
		result = append(result, card)
	}
	return result, nil
}
