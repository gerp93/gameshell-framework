package database

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

func getPlayerLobbyAccess(playerId uuid.UUID) (lobbyIds []uuid.UUID, err error) {
	sqlString := `
		SELECT DISTINCT
			L.ID
		FROM LOBBY AS L
			LEFT JOIN PLAYER_ACCESS_LOBBY AS PAL ON PAL.LOBBY_ID = L.ID
		WHERE L.PASSWORD_HASH IS NULL
			OR PAL.PLAYER_ID = ?
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

func getPlayerDeckAccess(playerId uuid.UUID) (deckIds []uuid.UUID, err error) {
	sqlString := `
		SELECT DISTINCT
			D.ID
		FROM DECK AS D
			LEFT JOIN PLAYER_ACCESS_DECK AS PAD ON PAD.DECK_ID = D.ID
		WHERE D.PASSWORD_HASH IS NULL
			OR PAD.PLAYER_ID = ?
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
