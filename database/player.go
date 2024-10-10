package database

import (
	"errors"
	"log"
	"regexp"
	"sort"
	"time"

	"github.com/google/uuid"
)

type Player struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	Name     string
	LobbyId  uuid.UUID
	UserId   uuid.UUID
	IsActive bool
}

type gameData struct {
	LobbyId            uuid.UUID
	LobbyName          string
	LobbyHandSize      int
	LobbyDrawPileCount int

	JudgeName     string
	JudgeCardText string

	BoardReady bool
	BoardPlays []boardPlay

	PlayerId      uuid.UUID
	PlayerIsJudge bool
	PlayerIsReady bool
	PlayerHand    []handCard
	PlayerPlays   []Card

	CardsToPlayCount int

	Wins []winDetail
}

type boardPlay struct {
	PlayerId       uuid.UUID
	PlayerUserName string
	Cards          []Card
}

type handCard struct {
	Card
	IsLocked bool
}

type winDetail struct {
	UserName string
	WinCount int
}

func GetPlayerGameData(playerId uuid.UUID) (data gameData, err error) {
	data.PlayerId = playerId

	sqlString := `
		SELECT
			L.ID AS LOBBY_ID,
			L.NAME AS LOBBY_NAME,
			L.HAND_SIZE AS LOBBY_HAND_SIZE,
			(SELECT COUNT(*) FROM DRAW_PILE WHERE LOBBY_ID = L.ID) AS LOBBY_DRAW_PILE_COUNT,
			JU.NAME AS JUDGE_NAME,
			JC.TEXT AS JUDGE_CARD_TEXT,
			EXISTS(SELECT ID FROM JUDGE WHERE PLAYER_ID = P.ID) AS PLAYER_IS_JUDGE
		FROM PLAYER AS P
			INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = P.LOBBY_ID
			INNER JOIN CARD AS JC ON JC.ID = J.CARD_ID
			INNER JOIN PLAYER AS JP ON JP.ID = J.PLAYER_ID
			INNER JOIN USER AS JU ON JU.ID = JP.USER_ID
		WHERE P.ID = ?
	`
	rows, err := query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&data.LobbyId,
			&data.LobbyName,
			&data.LobbyHandSize,
			&data.LobbyDrawPileCount,
			&data.JudgeName,
			&data.JudgeCardText,
			&data.PlayerIsJudge); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
	}

	sqlString = `
		SELECT
			P.ID AS PLAYER_ID,
			U.NAME AS PLAYER_USER_NAME
		FROM LOBBY AS L
			INNER JOIN PLAYER AS P ON P.LOBBY_ID = L.ID
			INNER JOIN USER AS U ON U.ID = P.USER_ID
			LEFT JOIN JUDGE AS J ON J.PLAYER_ID = P.ID
		WHERE L.ID = ?
			AND P.IS_ACTIVE = 1
			AND J.ID IS NULL
		ORDER BY U.NAME ASC
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var bp boardPlay
		if err := rows.Scan(
			&bp.PlayerId,
			&bp.PlayerUserName); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.BoardPlays = append(data.BoardPlays, bp)
	}

	totalCardsPlayedCount := 0
	playerCardsPlayedCount := 0
	for i, bp := range data.BoardPlays {
		sqlString = `
			SELECT
				C.ID AS CARD_ID,
				C.TEXT AS CARD_TEXT
			FROM BOARD AS B
				INNER JOIN CARD AS C ON C.ID = B.CARD_ID
			WHERE B.PLAYER_ID = ?
			ORDER BY B.CREATED_ON_DATE ASC
		`
		rows, err = query(sqlString, bp.PlayerId)
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
			data.BoardPlays[i].Cards = append(data.BoardPlays[i].Cards, card)

			totalCardsPlayedCount += 1
			if bp.PlayerId == playerId {
				playerCardsPlayedCount += 1
				data.PlayerPlays = append(data.PlayerPlays, card)
			}
		}
	}

	blankRegExp := regexp.MustCompile(`__+`)
	data.CardsToPlayCount = len(blankRegExp.FindAllString(data.JudgeCardText, -1))
	if data.CardsToPlayCount < 1 {
		data.CardsToPlayCount = 1
	}
	data.PlayerIsReady = playerCardsPlayedCount == data.CardsToPlayCount
	data.BoardReady = totalCardsPlayedCount == len(data.BoardPlays)*data.CardsToPlayCount

	if data.BoardReady {
		sort.Slice(data.BoardPlays, func(i, j int) bool {
			if len(data.BoardPlays[i].Cards) == 0 {
				return true
			}
			if len(data.BoardPlays[j].Cards) == 0 {
				return false
			}
			return data.BoardPlays[i].Cards[0].Text < data.BoardPlays[j].Cards[0].Text
		})
	}

	sqlString = `
		SELECT
			C.ID,
			C.TEXT,
			H.IS_LOCKED
		FROM HAND AS H
			INNER JOIN CARD AS C ON C.ID = H.CARD_ID
		WHERE H.PLAYER_ID = ?
		ORDER BY C.TEXT
	`
	rows, err = query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var card handCard
		if err := rows.Scan(
			&card.Id,
			&card.Text,
			&card.IsLocked); err != nil {
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
			AND LP.IS_ACTIVE = 1
		GROUP BY LP.USER_ID
		ORDER BY
			COUNT(W.ID) DESC,
			U.NAME ASC
	`
	rows, err = query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var win winDetail
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
	sqlString := "CALL SP_DRAW_HAND (?)"
	return execute(sqlString, playerId)
}

func PlayPlayerCard(playerId uuid.UUID, cardId uuid.UUID) error {
	sqlString := "CALL SP_PLAY_CARD (?, ?)"
	return execute(sqlString, playerId, cardId)
}

func WithdrawalPlayerCard(playerId uuid.UUID, cardId uuid.UUID) error {
	sqlString := "CALL SP_WITHDRAWAL_CARD (?, ?)"
	return execute(sqlString, playerId, cardId)
}

func DiscardPlayerHand(playerId uuid.UUID) error {
	sqlString := "CALL SP_DISCARD_HAND (?)"
	return execute(sqlString, playerId)
}

func DiscardPlayerCard(playerId uuid.UUID, cardId uuid.UUID) error {
	sqlString := "CALL SP_DISCARD_CARD (?, ?)"
	return execute(sqlString, playerId, cardId)
}

func LockPlayerCard(playerId uuid.UUID, cardId uuid.UUID, isLocked bool) error {
	sqlString := `
		UPDATE HAND
		SET IS_LOCKED = ?
		WHERE PLAYER_ID = ?
			AND CARD_ID = ?
	`
	return execute(sqlString, isLocked, playerId, cardId)
}

func GetPlayer(playerId uuid.UUID) (Player, error) {
	var player Player

	sqlString := `
		SELECT
			P.ID,
			P.CREATED_ON_DATE,
			P.CHANGED_ON_DATE,
			U.NAME,
			P.LOBBY_ID,
			P.USER_ID,
			P.IS_ACTIVE
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
		WHERE P.ID = ?
	`
	rows, err := query(sqlString, playerId)
	if err != nil {
		return player, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&player.Id,
			&player.CreatedOnDate,
			&player.ChangedOnDate,
			&player.Name,
			&player.LobbyId,
			&player.UserId,
			&player.IsActive); err != nil {
			log.Println(err)
			return player, errors.New("failed to scan row in query results")
		}
	}

	return player, nil
}

func getPlayerId(lobbyId uuid.UUID, userId uuid.UUID) (uuid.UUID, error) {
	var playerId uuid.UUID

	sqlString := `
		SELECT ID
		FROM PLAYER
		WHERE LOBBY_ID = ?
			AND USER_ID = ?
	`
	rows, err := query(sqlString, lobbyId, userId)
	if err != nil {
		return playerId, err
	}

	for rows.Next() {
		if err := rows.Scan(&playerId); err != nil {
			log.Println(err)
			return playerId, errors.New("failed to scan row in query results")
		}
	}

	if playerId == uuid.Nil {
		playerId, err = uuid.NewUUID()
		if err != nil {
			log.Println(err)
			return playerId, errors.New("failed to generate new player id")
		}
	}

	return playerId, nil
}
