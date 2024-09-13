package database

import (
	"github.com/google/uuid"
)

type playerData struct {
	PlayerId      uuid.UUID
	LobbyHandSize int
	PlayerIsJudge bool
	PlayerPlayed  bool
	PlayerHand    []Card
}

func GetPlayerData(playerId uuid.UUID) (data playerData, err error) {
	data.PlayerId = playerId

	sqlString := `
		SELECT
			L.HAND_SIZE,
			EXISTS(SELECT ID FROM JUDGE WHERE PLAYER_ID = P.ID) AS PLAYER_IS_JUDGE,
			EXISTS(SELECT ID FROM BOARD WHERE PLAYER_ID = P.ID) AS PLAYER_PLAYED,
			COALESCE(C.ID, UUID()),
			COALESCE(C.TEXT, '')
		FROM PLAYER AS P
			INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
			LEFT JOIN HAND AS H ON H.PLAYER_ID = P.ID
			LEFT JOIN CARD AS C ON C.ID = H.CARD_ID
		WHERE P.ID = ?
		ORDER BY C.TEXT
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var playerCard Card
		if err := rows.Scan(
			&data.LobbyHandSize,
			&data.PlayerIsJudge,
			&data.PlayerPlayed,
			&playerCard.Id,
			&playerCard.Text); err != nil {
			continue
		}
		if playerCard.Text != "" {
			data.PlayerHand = append(data.PlayerHand, playerCard)
		}
	}

	return data, nil
}

type playerGameBoard struct {
	PlayerId      uuid.UUID
	LobbyId       uuid.UUID
	PlayerIsJudge bool
	PlayerCount   int
	JudgeCardText string
	BoardCards    []Card
}

func GetPlayerGameBoard(playerId uuid.UUID) (data playerGameBoard, err error) {
	data.PlayerId = playerId

	sqlString := `
		SELECT
			L.ID AS LOBBY_ID,
			EXISTS(SELECT ID FROM JUDGE WHERE PLAYER_ID = P.ID) AS PLAYER_IS_JUDGE,
			(SELECT COUNT(ID) FROM PLAYER WHERE LOBBY_ID = L.ID) AS PLAYER_COUNT,
			COALESCE((
				SELECT JC.TEXT
				FROM JUDGE AS J
					INNER JOIN CARD AS JC ON JC.ID = J.CARD_ID
				WHERE J.LOBBY_ID = P.LOBBY_ID
			), '') AS JUDGE_CARD_TEXT,
			COALESCE(BC.ID, UUID()) AS BOARD_CARD_ID,
			COALESCE(BC.TEXT, '') AS BOARD_CARD_TEXT
		FROM PLAYER AS P
			INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
			LEFT JOIN BOARD AS B ON B.LOBBY_ID = P.LOBBY_ID
			LEFT JOIN CARD AS BC ON BC.ID = B.CARD_ID
		WHERE P.ID = ?
		ORDER BY BC.TEXT
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var boardCard Card
		if err := rows.Scan(
			&data.LobbyId,
			&data.PlayerIsJudge,
			&data.PlayerCount,
			&data.JudgeCardText,
			&boardCard.Id,
			&boardCard.Text); err != nil {
			return data, err
		}
		if boardCard.Text != "" {
			data.BoardCards = append(data.BoardCards, boardCard)
		}
	}

	data.PlayerCount -= 1 // do not count judge

	return data, nil
}

func DrawPlayerHand(playerId uuid.UUID) (data playerData, err error) {
	sqlString := `
		CALL SP_DRAW_HAND (?)
	`
	err = Execute(sqlString, playerId)
	if err != nil {
		return data, err
	}

	return GetPlayerData(playerId)
}

func PlayPlayerCard(playerId uuid.UUID, cardId uuid.UUID) (data playerData, err error) {
	sqlString := `
		INSERT INTO BOARD (LOBBY_ID, PLAYER_ID, CARD_ID)
		SELECT
			LOBBY_ID,
			ID,
			? AS CARD_ID
		FROM PLAYER
		WHERE ID = ?
	`
	err = Execute(sqlString, cardId, playerId)
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

func getPlayerName(playerId uuid.UUID) (name string, err error) {
	sqlString := `
		SELECT
			U.NAME
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
		WHERE P.ID = ?
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return name, err
	}

	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			return name, err
		}
	}

	return name, nil
}
