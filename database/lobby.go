package database

import (
	"database/sql"
	"errors"
	"log"
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

type LobbyDetails struct {
	Lobby
	UserCount int
}

func SearchLobbies(search string) ([]LobbyDetails, error) {
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
			INNER JOIN PLAYER AS P ON P.LOBBY_ID = L.ID
		WHERE L.NAME LIKE ?
		GROUP BY L.ID
		ORDER BY
			TO_DAYS(L.CHANGED_ON_DATE) DESC,
			L.NAME ASC,
			COUNT(P.ID) DESC
	`
	rows, err := Query(sqlString, search)
	if err != nil {
		return nil, err
	}

	result := make([]LobbyDetails, 0)
	for rows.Next() {
		var lobbyDetails LobbyDetails
		if err := rows.Scan(
			&lobbyDetails.Id,
			&lobbyDetails.CreatedOnDate,
			&lobbyDetails.ChangedOnDate,
			&lobbyDetails.Name,
			&lobbyDetails.PasswordHash,
			&lobbyDetails.HandSize,
			&lobbyDetails.UserCount); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
		result = append(result, lobbyDetails)
	}
	return result, nil
}

type lobbyGameInfo struct {
	Lobby
	CardCount int
	JudgeName string
}

func GetLobbyGameInfo(lobbyId uuid.UUID) (data lobbyGameInfo, err error) {
	sqlString := `
		SELECT
			L.ID,
			L.NAME,
			L.HAND_SIZE,
			COUNT(DP.CARD_ID) AS CARD_COUNT,
			U.NAME AS JUDGE_NAME
		FROM LOBBY AS L
			INNER JOIN DRAW_PILE AS DP ON DP.LOBBY_ID = L.ID
			INNER JOIN JUDGE AS J ON J.LOBBY_ID = L.ID
			INNER JOIN PLAYER AS P ON P.ID = J.PLAYER_ID
			INNER JOIN USER AS U ON U.ID = P.USER_ID
		WHERE L.ID = ?
		GROUP BY L.ID
	`
	rows, err := Query(sqlString, lobbyId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&data.Id,
			&data.Name,
			&data.HandSize,
			&data.CardCount,
			&data.JudgeName); err != nil {
			return data, err
		}
	}

	return data, nil
}

type lobbyGameStats struct {
	UserId   uuid.UUID
	UserName string
	Wins     int
}

func GetLobbyGameStats(lobbyId uuid.UUID) ([]lobbyGameStats, error) {
	sqlString := `
		SELECT
			P.USER_ID,
			U.NAME AS USER_NAME,
			COUNT(W.ID) AS WINS
		FROM PLAYER AS P
			LEFT JOIN WIN AS W ON W.PLAYER_ID = P.ID
			INNER JOIN USER AS U ON U.ID = P.USER_ID
		WHERE P.LOBBY_ID = ?
		GROUP BY P.USER_ID
		ORDER BY
			COUNT(W.ID) DESC,
			U.NAME ASC
	`
	rows, err := Query(sqlString, lobbyId)
	if err != nil {
		return nil, err
	}

	result := make([]lobbyGameStats, 0)
	for rows.Next() {
		var stats lobbyGameStats
		if err := rows.Scan(
			&stats.UserId,
			&stats.UserName,
			&stats.Wins); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
		result = append(result, stats)
	}
	return result, nil
}

func SkipJudgeCard(lobbyId uuid.UUID) error {
	sqlString := `
		CALL SP_SKIP_JUDGE (?)
	`
	return Execute(sqlString, lobbyId)
}

func PickLobbyWinner(lobbyId uuid.UUID, cardId uuid.UUID) (playerName string, err error) {
	sqlString := `
		CALL SP_PICK_WINNER (?, ?)
	`
	rows, err := Query(sqlString, lobbyId, cardId)
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
	rows, err := Query(sqlString, id)
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
	rows, err := Query(sqlString, id)
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

func CreateLobby(name string, password string, handSize int) (uuid.UUID, error) {
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
		INSERT INTO LOBBY (ID, NAME, PASSWORD_HASH, HAND_SIZE)
		VALUES (?, ?, ?, ?)
	`
	if password == "" {
		return id, Execute(sqlString, id, name, nil, handSize)
	} else {
		return id, Execute(sqlString, id, name, passwordHash, handSize)
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
		err := Execute(sqlString, lobbyId, deckId)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddUserToLobby(lobbyId uuid.UUID, userId uuid.UUID) (playerId uuid.UUID, err error) {
	playerId, err = uuid.NewUUID()
	if err != nil {
		log.Println(err)
		return playerId, errors.New("failed to generate new id")
	}

	sqlString := `
		INSERT IGNORE INTO PLAYER (ID, LOBBY_ID, USER_ID)
		VALUES (?, ?, ?)
	`
	err = Execute(sqlString, playerId, lobbyId, userId)
	if err != nil {
		return playerId, err
	}

	return playerId, err
}

func RemoveUserFromLobby(lobbyId uuid.UUID, userId uuid.UUID) error {
	sqlString := `
		DELETE FROM PLAYER
		WHERE LOBBY_ID = ?
			AND USER_ID = ?
	`
	return Execute(sqlString, lobbyId, userId)
}

func GetLobbyId(name string) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM LOBBY
		WHERE NAME = ?
	`
	rows, err := Query(sqlString, name)
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
	return Execute(sqlString, name, id)
}

func SetLobbyHandSize(id uuid.UUID, handSize int) error {
	sqlString := `
		UPDATE LOBBY
		SET
			HAND_SIZE = ?
		WHERE ID = ?
	`
	return Execute(sqlString, handSize, id)
}

func DeleteLobby(lobbyId uuid.UUID) error {
	sqlString := `
		DELETE FROM LOBBY
		WHERE ID = ?
	`
	return Execute(sqlString, lobbyId)
}
