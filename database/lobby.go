package database

import (
	"database/sql"
	"errors"
	"log"
	"regexp"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
)

type Lobby struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	Name         string
	PasswordHash sql.NullString
	HandSize     int
}

type lobbyDetails struct {
	Lobby
	UserCount int
}

type gameData struct {
	LobbyId            uuid.UUID
	LobbyName          string
	LobbyHandSize      int
	LobbyDrawPileCount int

	JudgeName     string
	JudgeCardText string

	BoardIsReady bool
	BoardPlays   []boardPlay

	PlayerIsJudge                bool
	PlayerIsReady                bool
	PlayerHand                   []handCard
	PlayerPlays                  []playCard
	PlayerSurpriseCardsRemaining int
	PlayerWildCardsRemaining     int

	CardsToPlayCount int

	Wins []winDetail
}

type boardPlay struct {
	PlayerId       uuid.UUID
	PlayerUserName string
	Cards          []playCard
}

type playCard struct {
	Card
	IsSurprise bool
}

type handCard struct {
	Card
	IsLocked bool
}

type winDetail struct {
	UserName string
	WinCount int
}

func SearchLobbies(search string) ([]lobbyDetails, error) {
	sqlString := `
		SELECT
			L.ID,
			L.CREATED_ON_DATE,
			L.CHANGED_ON_DATE,
			L.NAME,
			L.PASSWORD_HASH,
			L.HAND_SIZE,
			COUNT(P.ID) AS USER_COUNT
		FROM LOBBY AS L
			INNER JOIN PLAYER AS P ON P.LOBBY_ID = L.ID AND P.IS_ACTIVE = 1
		WHERE L.NAME LIKE ?
		GROUP BY L.ID
		ORDER BY
			TO_DAYS(L.CHANGED_ON_DATE) DESC,
			L.NAME ASC,
			COUNT(P.ID) DESC
	`
	rows, err := query(sqlString, search)
	if err != nil {
		return nil, err
	}

	result := make([]lobbyDetails, 0)
	for rows.Next() {
		var ld lobbyDetails
		if err := rows.Scan(
			&ld.Id,
			&ld.CreatedOnDate,
			&ld.ChangedOnDate,
			&ld.Name,
			&ld.PasswordHash,
			&ld.HandSize,
			&ld.UserCount); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
		result = append(result, ld)
	}
	return result, nil
}

func GetLobby(id uuid.UUID) (Lobby, error) {
	var lobby Lobby

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			PASSWORD_HASH,
			HAND_SIZE
		FROM LOBBY
		WHERE ID = ?
	`
	rows, err := query(sqlString, id)
	if err != nil {
		return lobby, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&lobby.Id,
			&lobby.CreatedOnDate,
			&lobby.ChangedOnDate,
			&lobby.Name,
			&lobby.PasswordHash,
			&lobby.HandSize); err != nil {
			log.Println(err)
			return lobby, errors.New("failed to scan row in query results")
		}
	}

	return lobby, nil
}

func GetLobbyPasswordHash(id uuid.UUID) (sql.NullString, error) {
	var passwordHash sql.NullString

	sqlString := `
		SELECT
			PASSWORD_HASH
		FROM LOBBY
		WHERE ID = ?
	`
	rows, err := query(sqlString, id)
	if err != nil {
		return passwordHash, err
	}

	for rows.Next() {
		if err := rows.Scan(&passwordHash); err != nil {
			log.Println(err)
			return passwordHash, errors.New("failed to scan row in query results")
		}
	}

	return passwordHash, nil
}

func CreateLobby(name string, password string, handSize int, surpriseCardLimit int, wildCardLimit int) (uuid.UUID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to generate new id")
	}

	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to hash password")
	}

	sqlString := `
		INSERT INTO LOBBY (ID, NAME, PASSWORD_HASH, HAND_SIZE, SURPRISE_CARD_LIMIT, WILD_CARD_LIMIT)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	if password == "" {
		return id, execute(sqlString, id, name, nil, handSize, surpriseCardLimit, wildCardLimit)
	} else {
		return id, execute(sqlString, id, name, passwordHash, handSize, surpriseCardLimit, wildCardLimit)
	}
}

func AddCardsToLobby(lobbyId uuid.UUID, deckIds []uuid.UUID) error {
	for _, deckId := range deckIds {
		sqlString := `
			INSERT INTO DRAW_PILE (LOBBY_ID, CARD_ID)
			SELECT
				? AS LOBBY_ID,
				ID AS CARD_ID
			FROM CARD
			WHERE DECK_ID = ?
		`
		err := execute(sqlString, lobbyId, deckId)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddUserToLobby(lobbyId uuid.UUID, userId uuid.UUID) (uuid.UUID, error) {
	player, err := GetPlayer(lobbyId, userId)
	if err != nil {
		log.Println(err)
		return player.Id, errors.New("failed to get player")
	}

	if player.Id == uuid.Nil {
		player.Id, err = uuid.NewUUID()
		if err != nil {
			log.Println(err)
			return player.Id, errors.New("failed to generate new player id")
		}
	}

	sqlString := "CALL SP_SET_PLAYER_ACTIVE (?, ?, ?)"
	err = execute(sqlString, player.Id, lobbyId, userId)
	return player.Id, err
}

func RemoveUserFromLobby(lobbyId uuid.UUID, userId uuid.UUID) error {
	sqlString := "CALL SP_SET_PLAYER_INACTIVE (?, ?)"
	return execute(sqlString, lobbyId, userId)
}

func GetLobbyId(name string) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM LOBBY
		WHERE NAME = ?
	`
	rows, err := query(sqlString, name)
	if err != nil {
		return id, err
	}

	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			log.Println(err)
			return id, errors.New("failed to scan row in query results")
		}
	}

	return id, nil
}

func SetLobbyName(id uuid.UUID, name string) error {
	sqlString := `
		UPDATE LOBBY
		SET
			NAME = ?
		WHERE ID = ?
	`
	return execute(sqlString, name, id)
}

func SetLobbyHandSize(id uuid.UUID, handSize int) error {
	sqlString := `
		UPDATE LOBBY
		SET
			HAND_SIZE = ?
		WHERE ID = ?
	`
	return execute(sqlString, handSize, id)
}

func SetLobbySurpriseCardLimit(id uuid.UUID, surpriseCardLimit int) error {
	sqlString := `
		CALL SP_SET_SURPRISE_CARD_LIMIT(?,?);
	`
	return execute(sqlString, id, surpriseCardLimit)

}

func SetLobbyWildCardLimit(id uuid.UUID, wildCardLimit int) error {
	sqlString := `
		CALL SP_SET_WILD_CARD_LIMIT(?,?);
	`
	return execute(sqlString, id, wildCardLimit)

}

func DeleteLobby(lobbyId uuid.UUID) error {
	sqlString := `
		DELETE FROM LOBBY
		WHERE ID = ?
	`
	return execute(sqlString, lobbyId)
}

func GetPlayerGameData(playerId uuid.UUID) (gameData, error) {
	var data gameData

	sqlString := `
		SELECT
			L.ID AS LOBBY_ID,
			L.NAME AS LOBBY_NAME,
			L.HAND_SIZE AS LOBBY_HAND_SIZE,
			(SELECT COUNT(*) FROM DRAW_PILE WHERE LOBBY_ID = L.ID) AS LOBBY_DRAW_PILE_COUNT,
			JU.NAME AS JUDGE_NAME,
			JC.TEXT AS JUDGE_CARD_TEXT,
			EXISTS(SELECT ID FROM JUDGE WHERE PLAYER_ID = P.ID) AS PLAYER_IS_JUDGE,
			P.SURPRISE_CARDS_REMAINING,
			P.WILD_CARDS_REMAINING
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
			&data.PlayerIsJudge,
			&data.PlayerSurpriseCardsRemaining,
			&data.PlayerWildCardsRemaining); err != nil {
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
				C.TEXT AS CARD_TEXT,
				B.IS_SURPRISE
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
			var playCard playCard
			if err := rows.Scan(
				&playCard.Id,
				&playCard.Text,
				&playCard.IsSurprise); err != nil {
				log.Println(err)
				return data, errors.New("failed to scan row in query results")
			}
			data.BoardPlays[i].Cards = append(data.BoardPlays[i].Cards, playCard)

			totalCardsPlayedCount += 1
			if bp.PlayerId == playerId {
				playerCardsPlayedCount += 1
				data.PlayerPlays = append(data.PlayerPlays, playCard)
			}
		}
	}

	blankRegExp := regexp.MustCompile(`__+`)
	data.CardsToPlayCount = len(blankRegExp.FindAllString(data.JudgeCardText, -1))
	if data.CardsToPlayCount < 1 {
		data.CardsToPlayCount = 1
	}
	data.PlayerIsReady = playerCardsPlayedCount == data.CardsToPlayCount
	data.BoardIsReady = totalCardsPlayedCount == len(data.BoardPlays)*data.CardsToPlayCount

	if data.BoardIsReady {
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
	sqlString := "CALL SP_PLAY_CARD (?, ?, 0, 0)"
	return execute(sqlString, playerId, cardId)
}

func PlaySurpriseCard(playerId uuid.UUID) error {
	sqlString := "CALL SP_PLAY_SURPRISE_CARD (?)"
	return execute(sqlString, playerId)
}

func PlayWildCard(playerId uuid.UUID, text string) error {
	sqlString := "CALL SP_PLAY_WILD_CARD (?, ?)"
	return execute(sqlString, playerId, text)
}

func WithdrawalPlayerCard(playerId uuid.UUID, cardId uuid.UUID) error {
	sqlString := "CALL SP_WITHDRAWAL_CARD (?, ?)"
	return execute(sqlString, playerId, cardId)
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

func PickLobbyWinner(lobbyId uuid.UUID, cardId uuid.UUID) (string, error) {
	var playerName string
	sqlString := "CALL SP_PICK_WINNER (?, ?)"
	rows, err := query(sqlString, lobbyId, cardId)
	if err != nil {
		return playerName, err
	}

	for rows.Next() {
		if err := rows.Scan(&playerName); err != nil {
			log.Println(err)
			return playerName, errors.New("failed to scan row in query results")
		}
	}

	return playerName, nil
}

func DiscardPlayerHand(playerId uuid.UUID) error {
	sqlString := "CALL SP_DISCARD_HAND (?)"
	return execute(sqlString, playerId)
}

func SkipJudgeCard(lobbyId uuid.UUID) error {
	sqlString := "CALL SP_SKIP_JUDGE_CARD (?)"
	return execute(sqlString, lobbyId)
}
