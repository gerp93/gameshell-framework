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
	JudgePlayer  sql.Null[Player]
	JudgeCard    sql.Null[Card]
}

type LobbyDetails struct {
	Lobby
	PlayerCount int
}

func GetLobbies(search string) ([]LobbyDetails, error) {
	sqlString := `
		SELECT
			L.ID,
			L.CREATED_ON_DATE,
			L.CHANGED_ON_DATE,
			L.NAME,
			L.PASSWORD_HASH,
			COUNT(LP.ID) AS PLAYER_COUNT
		FROM LOBBY AS L
			INNER JOIN LOBBY_PLAYER AS LP ON LP.LOBBY_ID = L.ID
		WHERE L.NAME LIKE ?
		GROUP BY L.ID
		ORDER BY L.CHANGED_ON_DATE DESC, L.NAME ASC, COUNT(LP.ID) DESC
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
			&lobbyDetails.PlayerCount); err != nil {
			continue
		}
		result = append(result, lobbyDetails)
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
			PASSWORD_HASH
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
			&lobby.PasswordHash); err != nil {
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

func CreateLobby(name string, password string) (uuid.UUID, error) {
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
		INSERT INTO LOBBY (ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?)
	`
	if password == "" {
		return id, Execute(sqlString, id, name, nil)
	} else {
		return id, Execute(sqlString, id, name, passwordHash)
	}
}

func AddCardsToLobby(lobbyId uuid.UUID, deckIds []uuid.UUID) error {
	for _, deckId := range deckIds {
		sqlString := `
			INSERT INTO LOBBY_CARD (LOBBY_ID, CARD_ID)
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

func AddPlayerToLobby(lobbyId uuid.UUID, playerId uuid.UUID) error {
	sqlString := `
		INSERT IGNORE INTO LOBBY_PLAYER (LOBBY_ID, PLAYER_ID)
		VALUES (?, ?)
	`
	return Execute(sqlString, lobbyId, playerId)
}

func RemovePlayerFromLobby(lobbyId uuid.UUID, playerId uuid.UUID) error {
	sqlString := `
		DELETE FROM LOBBY_PLAYER
		WHERE LOBBY_ID = ?
			AND PLAYER_ID = ?
	`
	return Execute(sqlString, lobbyId, playerId)
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

func SetLobbyPassword(id uuid.UUID, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return errors.New("failed to hash password")
	}

	sqlString := `
		UPDATE LOBBY
		SET
			PASSWORD_HASH = ?
		WHERE ID = ?
	`
	if password == "" {
		return Execute(sqlString, nil, id)
	} else {
		return Execute(sqlString, passwordHash, id)
	}
}

func DeleteLobby(lobbyId uuid.UUID) error {
	sqlString := `
		DELETE FROM LOBBY
		WHERE ID = ?
	`
	return Execute(sqlString, lobbyId)
}
