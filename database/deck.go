package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
)

type Deck struct {
	Id                uuid.UUID
	CreatedOnDate     time.Time
	ChangedOnDate     time.Time
	CreatedByPlayerId uuid.UUID
	ChangedByPlayerId uuid.UUID

	Name         string
	PasswordHash sql.NullString
}

func GetDecks() ([]Deck, error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			CREATED_BY_PLAYER_ID,
			CHANGED_BY_PLAYER_ID,
			NAME,
			PASSWORD_HASH
		FROM DECK
		ORDER BY CHANGED_ON_DATE DESC
	`)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query()
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to query statement in database")
	}

	result := make([]Deck, 0)
	for rows.Next() {
		var deck Deck
		if err := rows.Scan(
			&deck.Id,
			&deck.CreatedOnDate,
			&deck.ChangedOnDate,
			&deck.CreatedByPlayerId,
			&deck.ChangedByPlayerId,
			&deck.Name,
			&deck.PasswordHash); err != nil {
			continue
		}
		result = append(result, deck)
	}
	return result, nil
}

func GetDeck(id uuid.UUID) (Deck, error) {
	var deck Deck

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return deck, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			CREATED_BY_PLAYER_ID,
			CHANGED_BY_PLAYER_ID,
			NAME,
			PASSWORD_HASH
		FROM DECK
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return deck, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query(id)
	if err != nil {
		log.Println(err)
		return deck, errors.New("failed to query statement in database")
	}

	for rows.Next() {
		if err := rows.Scan(
			&deck.Id,
			&deck.CreatedOnDate,
			&deck.ChangedOnDate,
			&deck.CreatedByPlayerId,
			&deck.ChangedByPlayerId,
			&deck.Name,
			&deck.PasswordHash); err != nil {
			return deck, err
		}
	}

	return deck, nil
}

func CreateDeck(playerId uuid.UUID, name string, password string) (uuid.UUID, error) {
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
		log.Println(err)
		return id, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		INSERT INTO DECK (ID, CREATED_BY_PLAYER_ID, CHANGED_BY_PLAYER_ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	if password == "" {
		_, err = statement.Exec(id, playerId, playerId, name, nil)
	} else {
		_, err = statement.Exec(id, playerId, playerId, name, passwordHash)
	}
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to execute statement in database")
	}

	return id, nil
}

func UpdateDeck(playerId uuid.UUID, id uuid.UUID, name string, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		UPDATE DECK
		SET
			NAME = ?,
			PASSWORD_HASH = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(),
			CHANGED_BY_PLAYER_ID = ?
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	if password == "" {
		_, err = statement.Exec(name, nil, playerId, id)
	} else {
		_, err = statement.Exec(name, passwordHash, playerId, id)
	}
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}

func DeleteDeck(id uuid.UUID) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		DELETE FROM DECK
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}
