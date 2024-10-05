package database

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type Player struct {
	Id uuid.UUID
}

type winDetails struct {
	UserName string
	WinCount int
}

type gameData struct {
	LobbyId            uuid.UUID
	LobbyName          string
	LobbyHandSize      int
	LobbyPlayerCount   int
	LobbyDrawPileCount int

	JudgeId       uuid.UUID
	JudgeName     string
	JudgeCardText string

	BoardCards []Card

	PlayerId      uuid.UUID
	PlayerHand    []Card
	PlayerIsJudge bool
	PlayerPlayed  bool

	Wins []winDetails
}

func GetPlayerGameData(playerId uuid.UUID) (data gameData, err error) {
	data.PlayerId = playerId

	sqlString := `
		SELECT
			L.ID AS LOBBY_ID,
			L.NAME AS LOBBY_NAME,
			L.HAND_SIZE AS LOBBY_HAND_SIZE,
			(SELECT COUNT(*) FROM PLAYER WHERE LOBBY_ID = L.ID AND ACTIVE = 1) AS LOBBY_PLAYER_COUNT,
			(SELECT COUNT(*) FROM DRAW_PILE WHERE LOBBY_ID = L.ID) AS LOBBY_DRAW_PILE_COUNT,
			J.ID AS JUDGE_ID,
			JU.NAME AS JUDGE_NAME,
			JC.TEXT AS JUDGE_CARD_TEXT,
			EXISTS(SELECT ID FROM JUDGE WHERE PLAYER_ID = P.ID) AS PLAYER_IS_JUDGE,
			EXISTS(SELECT ID FROM BOARD WHERE PLAYER_ID = P.ID) AS PLAYER_PLAYED
		FROM PLAYER AS P
			INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = P.LOBBY_ID
			INNER JOIN CARD AS JC ON JC.ID = J.CARD_ID
			INNER JOIN PLAYER AS JP ON JP.ID = J.PLAYER_ID
			INNER JOIN USER AS JU ON JU.ID = JP.USER_ID
		WHERE P.ID = ?
	`
	rows, err := Query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&data.LobbyId,
			&data.LobbyName,
			&data.LobbyHandSize,
			&data.LobbyPlayerCount,
			&data.LobbyDrawPileCount,
			&data.JudgeId,
			&data.JudgeName,
			&data.JudgeCardText,
			&data.PlayerIsJudge,
			&data.PlayerPlayed); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
	}

	data.LobbyPlayerCount -= 1 // do not count judge

	sqlString = `
		SELECT
			C.ID,
			C.TEXT
		FROM BOARD AS B
			INNER JOIN CARD AS C ON C.ID = B.CARD_ID
			INNER JOIN PLAYER AS P ON P.LOBBY_ID = B.LOBBY_ID
		WHERE P.ID = ?
		ORDER BY C.TEXT
	`
	rows, err = Query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var card Card
		if err := rows.Scan(
			&card.Id,
			&card.Text); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.BoardCards = append(data.BoardCards, card)
	}

	sqlString = `
		SELECT
			C.ID,
			C.TEXT
		FROM HAND AS H
			INNER JOIN CARD AS C ON C.ID = H.CARD_ID
		WHERE H.PLAYER_ID = ?
		ORDER BY C.TEXT
	`
	rows, err = Query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var card Card
		if err := rows.Scan(
			&card.Id,
			&card.Text); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.PlayerHand = append(data.PlayerHand, card)
	}

	sqlString = `
		SELECT
			U.NAME AS USER_NAME,
			COUNT(W.ID) AS WINS
		FROM PLAYER AS P
			INNER JOIN PLAYER AS LP ON LP.LOBBY_ID = P.LOBBY_ID
			INNER JOIN USER AS U ON U.ID = LP.USER_ID
			LEFT JOIN WIN AS W ON W.PLAYER_ID = LP.ID
		WHERE P.ID = ?
		AND LP.ACTIVE = 1
		GROUP BY LP.USER_ID
		ORDER BY
			COUNT(W.ID) DESC,
			U.NAME ASC
	`
	rows, err = Query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var win winDetails
		if err := rows.Scan(
			&win.UserName,
			&win.WinCount); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.Wins = append(data.Wins, win)
	}

	return data, nil
}

func DrawPlayerHand(playerId uuid.UUID) error {
	sqlString := `
		CALL SP_DRAW_HAND (?)
	`
	return Execute(sqlString, playerId)
}

func PlayPlayerCard(playerId uuid.UUID, cardId uuid.UUID) error {

	sqlString := `
		INSERT INTO CARD_PLAY (CARD_ID, PLAYER_USER_ID, JUDGE_USER_ID)
		SELECT
			H.CARD_ID,
			PLAYER_USER.ID,
			JUDGE_USER.ID
		FROM HAND H
			INNER JOIN PLAYER P on P.ID = H.PLAYER_ID
			INNER JOIN LOBBY L ON L.ID = P.LOBBY_ID
			INNER JOIN JUDGE J ON J.LOBBY_ID = L.ID
			INNER JOIN USER PLAYER_USER on P.USER_ID = PLAYER_USER.ID
			INNER JOIN PLAYER JUDGE_PLAYER on J.PLAYER_ID = JUDGE_PLAYER.ID
			INNER JOIN USER JUDGE_USER on judge_player.USER_ID = JUDGE_USER.ID
		WHERE h.CARD_ID = ?
		AND h.PLAYER_ID = ?
	`

	Execute(sqlString, cardId, playerId)

	sqlString = `
		INSERT INTO BOARD (LOBBY_ID, PLAYER_ID, CARD_ID)
		SELECT
			LOBBY_ID,
			ID,
			? AS CARD_ID
		FROM PLAYER
		WHERE ID = ?
	`
	err := Execute(sqlString, cardId, playerId)
	if err != nil {
		return err
	}

	return DiscardPlayerCard(playerId, cardId, false)
}

func DiscardPlayerHand(playerId uuid.UUID) error {

	sqlString := `
	INSERT INTO CARD_DISCARD(CARD_ID, PLAYER_USER_ID)
	SELECT H.CARD_ID, P.USER_ID FROM HAND H
		INNER JOIN PLAYER P ON P.ID = H.PLAYER_ID
	WHERE H.PLAYER_ID = ?
	`
	Execute(sqlString, playerId)

	sqlString = `
		DELETE FROM HAND
		WHERE PLAYER_ID = ?
	`
	return Execute(sqlString, playerId)
}

func DiscardPlayerCard(playerId uuid.UUID, cardId uuid.UUID, recordDiscard bool) error {

	sqlString := ""

	if recordDiscard {

		sqlString = `
		INSERT INTO CARD_DISCARD(CARD_ID, PLAYER_USER_ID)
		SELECT H.CARD_ID, P.USER_ID FROM HAND H
			INNER JOIN PLAYER P ON p.ID = H.PLAYER_ID
		WHERE H.PLAYER_ID = ?
			AND H.CARD_ID = ?
		`

		Execute(sqlString, playerId, cardId)
	}

	sqlString = `
	DELETE FROM HAND
	WHERE PLAYER_ID = ?
		AND CARD_ID = ?	
	`
	return Execute(sqlString, playerId, cardId)
}

func GetPlayerId(lobbyId uuid.UUID, userId uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
	SELECT ID FROM PLAYER
	WHERE LOBBY_ID = ?
	AND USER_ID = ?
	`
	rows, err := Query(sqlString, lobbyId, userId)
	if err != nil {
		fmt.Println("here 1")
		return id, err
	}

	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			fmt.Println("here 2")
			log.Println(err)
			return id, errors.New("failed to scan row in query results")
		}
	}

	fmt.Println(id)
	if id == uuid.Nil {
		id, err = uuid.NewUUID()
		if err != nil {
			log.Println(err)
			return id, errors.New("failed to generate new player id")
		}
	}

	fmt.Println(id)
	return id, nil
}
