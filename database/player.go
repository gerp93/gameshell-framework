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
			EXISTS(SELECT ID FROM JUDGE WHERE PLAYER_ID = ?) AS PLAYER_IS_JUDGE,
			EXISTS(SELECT ID FROM BOARD WHERE PLAYER_ID = ?) AS PLAYER_PLAYED,
			COALESCE(C.ID, UUID()),
			COALESCE(C.TEXT, '')
		FROM PLAYER AS P
			INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
			LEFT JOIN HAND AS H ON H.PLAYER_ID = P.ID
			LEFT JOIN CARD AS C ON C.ID = H.CARD_ID
		WHERE P.ID = ?
		ORDER BY C.TEXT
	`
	rows, err := Query(sqlString, playerId, playerId, playerId)
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
	JudgeCard     Card
	BoardCards    []Card
	LobbyId       uuid.UUID
	PlayerId      uuid.UUID
	PlayerIsJudge bool
	PlayerCount   int
}

func GetPlayerGameBoard(playerId uuid.UUID) (data playerGameBoard, err error) {
	lobbyId, err := getPlayerLobbyId(playerId)
	if err != nil {
		return data, err
	}

	sqlString := `
		SELECT
			C.ID,
			C.TEXT
		FROM JUDGE AS J
			INNER JOIN CARD AS C ON C.ID = J.CARD_ID
		WHERE J.LOBBY_ID = ?
	`
	rows, err := Query(sqlString, lobbyId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&data.JudgeCard.Id,
			&data.JudgeCard.Text); err != nil {
			return data, err
		}
	}

	sqlString = `
		SELECT
			C.ID,
			C.TEXT
		FROM BOARD AS B
			INNER JOIN CARD AS C ON C.ID = B.CARD_ID
		WHERE B.LOBBY_ID = ?
		ORDER BY C.TEXT
	`
	rows, err = Query(sqlString, lobbyId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var card Card
		if err := rows.Scan(
			&card.Id,
			&card.Text); err != nil {
			continue
		}
		data.BoardCards = append(data.BoardCards, card)
	}

	data.LobbyId, err = getPlayerLobbyId(playerId)
	if err != nil {
		return data, err
	}

	data.PlayerId = playerId

	data.PlayerIsJudge, err = isPlayerJudge(playerId)
	if err != nil {
		return data, err
	}

	data.PlayerCount, err = getLobbyPlayerCount(playerId)
	if err != nil {
		return data, err
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
	lobbyId, err := getPlayerLobbyId(playerId)
	if err != nil {
		return data, err
	}

	sqlString := `
		INSERT INTO BOARD (LOBBY_ID, PLAYER_ID, CARD_ID)
		VALUES (?, ?, ?)
	`
	err = Execute(sqlString, lobbyId, playerId, cardId)
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

func getPlayerLobbyId(playerId uuid.UUID) (lobbyId uuid.UUID, err error) {
	sqlString := `
		SELECT
			LOBBY_ID
		FROM PLAYER
		WHERE ID = ?
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return lobbyId, err
	}

	for rows.Next() {
		if err := rows.Scan(&lobbyId); err != nil {
			return lobbyId, err
		}
	}

	return lobbyId, nil
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

func isPlayerJudge(playerId uuid.UUID) (bool, error) {
	sqlString := `
		SELECT
			ID
		FROM JUDGE
		WHERE PLAYER_ID = ?
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return false, err
	}

	return rows.Next(), nil
}

func getLobbyPlayerCount(playerId uuid.UUID) (playerCount int, err error) {
	sqlString := `
		SELECT
			COUNT(LP.ID)
		FROM PLAYER AS P
			INNER JOIN PLAYER AS LP ON LP.LOBBY_ID = P.LOBBY_ID
		WHERE P.ID = ?
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return playerCount, err
	}

	for rows.Next() {
		if err := rows.Scan(&playerCount); err != nil {
			return playerCount, err
		}
	}

	return playerCount, nil
}
