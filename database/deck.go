package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
)

type Deck struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	Name         string
	PasswordHash sql.NullString
}

func GetDecks(dbcs string) ([]Deck, error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			PASSWORD_HASH
		FROM DECK
		ORDER BY CHANGED_ON_DATE DESC
	`)
	if err != nil {
		return nil, err
	}
	defer statment.Close()

	rows, err := statment.Query()
	if err != nil {
		return nil, err
	}

	result := make([]Deck, 0)
	for rows.Next() {
		var deck Deck
		if err := rows.Scan(
			&deck.Id,
			&deck.CreatedOnDate,
			&deck.ChangedOnDate,
			&deck.Name,
			&deck.PasswordHash); err != nil {
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

	statment, err := db.Prepare(`
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			PASSWORD_HASH
		FROM DECK
		WHERE ID = ?
	`)
	if err != nil {
		return deck, err
	}
	defer statment.Close()

	rows, err := statment.Query(id)
	if err != nil {
		return deck, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&deck.Id,
			&deck.CreatedOnDate,
			&deck.ChangedOnDate,
			&deck.Name,
			&deck.PasswordHash); err != nil {
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

	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		return id, err
	}

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return id, err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		INSERT INTO DECK (ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return id, err
	}
	defer statment.Close()

	if password == "" {
		_, err = statment.Exec(id, name, nil)
	} else {
		_, err = statment.Exec(id, name, passwordHash)
	}
	if err != nil {
		return id, err
	}

	return id, nil
}

func UpdateDeck(dbcs string, id uuid.UUID, name string, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		UPDATE DECK
		SET
			NAME = ?,
			PASSWORD_HASH = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`)
	if err != nil {
		return err
	}
	defer statment.Close()

	if password == "" {
		_, err = statment.Exec(name, nil, id)
	} else {
		_, err = statment.Exec(name, passwordHash, id)
	}
	if err != nil {
		return err
	}

	return nil
}

func DeleteDeck(dbcs string, id uuid.UUID) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		DELETE FROM DECK
		WHERE ID = ?
	`)
	if err != nil {
		return err
	}
	defer statment.Close()

	_, err = statment.Exec(id)
	if err != nil {
		return err
	}

	return nil
}
