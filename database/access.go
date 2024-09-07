package database

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

func GetPlayerLobbyAccess(playerId uuid.UUID) (lobbyIds []uuid.UUID, err error) {
	sqlString := `
		SELECT
			LOBBY_ID
		FROM PLAYER_ACCESS_LOBBY
		WHERE PLAYER_ID = ?
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return nil, err
	}

	lobbyIds = make([]uuid.UUID, 0)
	for rows.Next() {
		var lobbyId uuid.UUID
		if err := rows.Scan(&lobbyId); err != nil {
			log.Println(err)
			return lobbyIds, errors.New("failed to scan row in query results")
		}
		lobbyIds = append(lobbyIds, lobbyId)
	}

	return lobbyIds, nil
}

func AddPlayerLobbyAccess(playerId uuid.UUID, lobbyId uuid.UUID) error {
	sqlString := `
		INSERT INTO PLAYER_ACCESS_LOBBY (PLAYER_ID, LOBBY_ID)
		VALUES (?, ?)
	`
	return Execute(sqlString, playerId, lobbyId)
}

func GetPlayerDeckAccess(playerId uuid.UUID) (deckIds []uuid.UUID, err error) {
	sqlString := `
		SELECT
			DECK_ID
		FROM PLAYER_ACCESS_DECK
		WHERE PLAYER_ID = ?
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return nil, err
	}

	deckIds = make([]uuid.UUID, 0)
	for rows.Next() {
		var deckId uuid.UUID
		if err := rows.Scan(&deckId); err != nil {
			log.Println(err)
			return deckIds, errors.New("failed to scan row in query results")
		}
		deckIds = append(deckIds, deckId)
	}

	return deckIds, nil
}

func AddPlayerDeckAccess(playerId uuid.UUID, deckId uuid.UUID) error {
	sqlString := `
		INSERT INTO PLAYER_ACCESS_DECK (PLAYER_ID, DECK_ID)
		VALUES (?, ?)
	`
	return Execute(sqlString, playerId, deckId)
}
