package database

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

type playerData struct {
	PlayerId uuid.UUID
	Cards    []Card
	HandSize int
}

func GetPlayerData(playerId uuid.UUID) (data playerData, err error) {
	data.PlayerId = playerId

	data.Cards, err = getPlayerHand(playerId)
	if err != nil {
		return data, err
	}

	data.HandSize, err = getPlayerHandSize(playerId)
	if err != nil {
		return data, err
	}

	return data, nil
}

func DrawPlayerHand(playerId uuid.UUID) (data playerData, err error) {
	handCount, err := getPlayerHandCount(playerId)
	if err != nil {
		return data, err
	}

	handSize, err := getPlayerHandSize(playerId)
	if err != nil {
		return data, err
	}

	cardsToDraw := handSize - handCount
	if cardsToDraw > 0 {
		sqlString := `
			INSERT INTO HAND
				(
					PLAYER_ID,
					CARD_ID
				)
			SELECT DISTINCT
				P.ID AS PLAYER_ID,
				C.ID AS CARD_ID
			FROM DRAW_PILE AS DP
				INNER JOIN PLAYER AS P ON P.LOBBY_ID = DP.LOBBY_ID
				INNER JOIN CARD AS C ON C.ID = DP.CARD_ID
				INNER JOIN CARD_TYPE AS CT ON CT.ID = C.CARD_TYPE_ID
			WHERE CT.NAME = 'Player'
				AND P.ID = ?
			ORDER BY RAND()
			LIMIT ?
		`
		err = Execute(sqlString, playerId, cardsToDraw)
		if err != nil {
			return data, err
		}

		err = removeUserHandFromLobbyCards()
		if err != nil {
			return data, err
		}
	}

	return GetPlayerData(playerId)
}

func PlayPlayerCard(playerId uuid.UUID, cardId uuid.UUID) (data playerData, err error) {
	sqlString := `
		INSERT INTO BOARD (PLAYER_ID, CARD_ID)
		VALUES (?, ?)
	`
	err = Execute(sqlString, playerId, cardId)
	if err != nil {
		return data, err
	}

	return DiscardPlayerCard(playerId, cardId)
}

func DiscardPlayerHand(playerId uuid.UUID) (data playerData, err error) {
	sqlString := `
		DELETE FROM HAND
		WHERE PLAYER_ID = ?
	`
	err = Execute(sqlString, playerId)
	if err != nil {
		return data, err
	}

	return GetPlayerData(playerId)
}

func DiscardPlayerCard(playerId uuid.UUID, cardId uuid.UUID) (data playerData, err error) {
	sqlString := `
		DELETE FROM HAND
		WHERE PLAYER_ID = ?
			AND CARD_ID = ?
	`
	err = Execute(sqlString, playerId, cardId)
	if err != nil {
		return data, err
	}

	return GetPlayerData(playerId)
}

func getPlayerHandSize(playerId uuid.UUID) (handSize int, err error) {
	sqlString := `
		SELECT
			L.HAND_SIZE
		FROM LOBBY AS L
			INNER JOIN PLAYER AS P ON P.LOBBY_ID = L.ID
		WHERE P.ID = ?
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return handSize, err
	}

	for rows.Next() {
		if err := rows.Scan(&handSize); err != nil {
			log.Println(err)
			return handSize, errors.New("failed to scan row in query results")
		}
	}

	return handSize, nil
}

func getPlayerHandCount(playerId uuid.UUID) (handCount int, err error) {
	sqlString := `
		SELECT
			COUNT(CARD_ID)
		FROM HAND
		WHERE PLAYER_ID = ?
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return handCount, err
	}

	for rows.Next() {
		if err := rows.Scan(&handCount); err != nil {
			return handCount, err
		}
	}

	return handCount, nil
}

func removeUserHandFromLobbyCards() error {
	sqlString := `
		DELETE DP
		FROM DRAW_PILE AS DP
			INNER JOIN PLAYER AS P ON P.LOBBY_ID = DP.LOBBY_ID
			INNER JOIN HAND AS H ON H.PLAYER_ID = P.ID AND H.CARD_ID = DP.CARD_ID
	`
	return Execute(sqlString)
}

func getPlayerHand(playerId uuid.UUID) ([]Card, error) {
	sqlString := `
		SELECT
			C.ID,
			C.TEXT
		FROM HAND AS H
			INNER JOIN CARD AS C ON C.ID = H.CARD_ID
		WHERE H.PLAYER_ID = ?
		ORDER BY C.TEXT
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return nil, err
	}

	result := make([]Card, 0)
	for rows.Next() {
		var card Card
		if err := rows.Scan(
			&card.Id,
			&card.Text); err != nil {
			continue
		}
		result = append(result, card)
	}
	return result, nil
}
