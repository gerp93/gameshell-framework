package database

import (
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
)

func GetPlayerLobbyAccess(playerId uuid.UUID) (lobbyIds []uuid.UUID, err error) {
	lobbyIds = make([]uuid.UUID, 0)

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return lobbyIds, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		SELECT
			LOBBY_ID
		FROM PLAYER_ACCESS_LOBBY
		WHERE PLAYER_ID = ?
	`)
	if err != nil {
		log.Println(err)
		return lobbyIds, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query(playerId)
	if err != nil {
		log.Println(err)
		return lobbyIds, errors.New("failed to query statement in database")
	}

	for rows.Next() {
		var lobbyId uuid.UUID
		if err := rows.Scan(&lobbyId); err != nil {
			return lobbyIds, err
		}
		lobbyIds = append(lobbyIds, lobbyId)
	}

	return lobbyIds, nil
}

func AddPlayerLobbyAccess(playerId uuid.UUID, lobbyId uuid.UUID) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		INSERT INTO PLAYER_ACCESS_LOBBY (PLAYER_ID, LOBBY_ID)
		VALUES (?, ?)
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(playerId, lobbyId)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}

func GetPlayerDeckAccess(playerId uuid.UUID) (deckIds []uuid.UUID, err error) {
	deckIds = make([]uuid.UUID, 0)

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return deckIds, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		SELECT
			DECK_ID
		FROM PLAYER_ACCESS_DECK
		WHERE PLAYER_ID = ?
	`)
	if err != nil {
		log.Println(err)
		return deckIds, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query(playerId)
	if err != nil {
		log.Println(err)
		return deckIds, errors.New("failed to query statement in database")
	}

	for rows.Next() {
		var deckId uuid.UUID
		if err := rows.Scan(&deckId); err != nil {
			return deckIds, err
		}
		deckIds = append(deckIds, deckId)
	}

	return deckIds, nil
}

func AddPlayerDeckAccess(playerId uuid.UUID, deckId uuid.UUID) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		INSERT INTO PLAYER_ACCESS_DECK (PLAYER_ID, DECK_ID)
		VALUES (?, ?)
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(playerId, deckId)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}
