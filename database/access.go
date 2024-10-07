package database

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

func getUserLobbyAccess(userId uuid.UUID) (lobbyIds []uuid.UUID, err error) {
	sqlString := "CALL SP_GET_LOBBY_ACCESS (?)"
	rows, err := query(sqlString, userId)
	if err != nil {
		return nil, err
	}

	lobbyIds = make([]uuid.UUID, 0)
	for rows.Next() {
		var lobbyId uuid.UUID
		var lobbyName string
		if err := rows.Scan(&lobbyId, &lobbyName); err != nil {
			log.Println(err)
			return lobbyIds, errors.New("failed to scan row in query results")
		}
		lobbyIds = append(lobbyIds, lobbyId)
	}

	return lobbyIds, nil
}

func AddUserLobbyAccess(userId uuid.UUID, lobbyId uuid.UUID) error {
	sqlString := `
		INSERT INTO USER_ACCESS_LOBBY (USER_ID, LOBBY_ID)
		VALUES (?, ?)
	`
	return execute(sqlString, userId, lobbyId)
}

func getUserDeckAccess(userId uuid.UUID) (deckIds []uuid.UUID, err error) {
	sqlString := "CALL SP_GET_DECK_ACCESS (?)"
	rows, err := query(sqlString, userId)
	if err != nil {
		return nil, err
	}

	deckIds = make([]uuid.UUID, 0)
	for rows.Next() {
		var deckId uuid.UUID
		var deckName string
		if err := rows.Scan(&deckId, &deckName); err != nil {
			log.Println(err)
			return deckIds, errors.New("failed to scan row in query results")
		}
		deckIds = append(deckIds, deckId)
	}

	return deckIds, nil
}

func AddUserDeckAccess(userId uuid.UUID, deckId uuid.UUID) error {
	sqlString := `
		INSERT INTO USER_ACCESS_DECK (USER_ID, DECK_ID)
		VALUES (?, ?)
	`
	return execute(sqlString, userId, deckId)
}
