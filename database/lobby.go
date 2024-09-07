package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
	"github.com/grantfbarnes/card-judge/helper"
)

type Lobby struct {
	Id                uuid.UUID
	CreatedOnDate     time.Time
	ChangedOnDate     time.Time
	CreatedByPlayerId uuid.UUID
	ChangedByPlayerId uuid.UUID

	Name         string
	PasswordHash sql.NullString
	DeckIds      []uuid.UUID
}

func (l Lobby) HasDeck(deckId uuid.UUID) bool {
	return helper.IsIdInArray(deckId, l.DeckIds)
}

func LobbyHasDeck(lobbyId uuid.UUID, deckId uuid.UUID) bool {
	deckIds, err := getLobbyDecks(lobbyId)
	if err != nil {
		return false
	}
	return helper.IsIdInArray(deckId, deckIds)
}

func GetLobbies() ([]Lobby, error) {
	rows, err := Query(`
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			CREATED_BY_PLAYER_ID,
			CHANGED_BY_PLAYER_ID,
			NAME,
			PASSWORD_HASH
		FROM LOBBY
		ORDER BY CHANGED_ON_DATE DESC
	`)
	if err != nil {
		return nil, err
	}

	result := make([]Lobby, 0)
	for rows.Next() {
		var lobby Lobby
		if err := rows.Scan(
			&lobby.Id,
			&lobby.CreatedOnDate,
			&lobby.ChangedOnDate,
			&lobby.CreatedByPlayerId,
			&lobby.ChangedByPlayerId,
			&lobby.Name,
			&lobby.PasswordHash); err != nil {
			continue
		}
		result = append(result, lobby)
	}
	return result, nil
}

func SearchLobbies(search string) ([]Lobby, error) {
	rows, err := Query(`
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			CREATED_BY_PLAYER_ID,
			CHANGED_BY_PLAYER_ID,
			NAME,
			PASSWORD_HASH
		FROM LOBBY
		WHERE NAME LIKE ?
		ORDER BY CHANGED_ON_DATE DESC
	`, search)
	if err != nil {
		return nil, err
	}

	result := make([]Lobby, 0)
	for rows.Next() {
		var lobby Lobby
		if err := rows.Scan(
			&lobby.Id,
			&lobby.CreatedOnDate,
			&lobby.ChangedOnDate,
			&lobby.CreatedByPlayerId,
			&lobby.ChangedByPlayerId,
			&lobby.Name,
			&lobby.PasswordHash); err != nil {
			continue
		}
		result = append(result, lobby)
	}
	return result, nil
}

func GetLobby(id uuid.UUID) (Lobby, error) {
	var lobby Lobby

	rows, err := Query(`
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			CREATED_BY_PLAYER_ID,
			CHANGED_BY_PLAYER_ID,
			NAME,
			PASSWORD_HASH
		FROM LOBBY
		WHERE ID = ?
	`, id)
	if err != nil {
		return lobby, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&lobby.Id,
			&lobby.CreatedOnDate,
			&lobby.ChangedOnDate,
			&lobby.CreatedByPlayerId,
			&lobby.ChangedByPlayerId,
			&lobby.Name,
			&lobby.PasswordHash); err != nil {
			log.Println(err)
			return lobby, errors.New("failed to scan row in query results")
		}
	}

	lobby.DeckIds, err = getLobbyDecks(lobby.Id)
	if err != nil {
		lobby.DeckIds = make([]uuid.UUID, 0)
	}

	return lobby, nil
}

func CreateLobby(playerId uuid.UUID, name string, password string) (uuid.UUID, error) {
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
		INSERT INTO LOBBY (ID, CREATED_BY_PLAYER_ID, CHANGED_BY_PLAYER_ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?, ?, ?)
	`
	if password == "" {
		return id, Execute(sqlString, id, playerId, playerId, name, nil)
	} else {
		return id, Execute(sqlString, id, playerId, playerId, name, passwordHash)
	}
}

func GetLobbyId(name string) (uuid.UUID, error) {
	var id uuid.UUID

	rows, err := Query(`
		SELECT
			ID
		FROM LOBBY
		WHERE NAME = ?
	`, name)
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

func SetLobbyName(playerId uuid.UUID, id uuid.UUID, name string) error {
	sqlString := `
		UPDATE LOBBY
		SET
			NAME = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(),
			CHANGED_BY_PLAYER_ID = ?
		WHERE ID = ?
	`
	return Execute(sqlString, name, playerId, id)
}

func SetLobbyPassword(playerId uuid.UUID, id uuid.UUID, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return errors.New("failed to hash password")
	}

	sqlString := `
		UPDATE LOBBY
		SET
			PASSWORD_HASH = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(),
			CHANGED_BY_PLAYER_ID = ?
		WHERE ID = ?
	`
	if password == "" {
		return Execute(sqlString, nil, playerId, id)
	} else {
		return Execute(sqlString, passwordHash, playerId, id)
	}
}

func AddDeckToLobby(playerId uuid.UUID, id uuid.UUID, deckId uuid.UUID) error {
	sqlString := `
		INSERT INTO LOBBY_DECK (LOBBY_ID, DECK_ID)
		VALUES (?, ?)
	`
	return Execute(sqlString, id, deckId)
}

func RemoveDeckFromLobby(playerId uuid.UUID, id uuid.UUID, deckId uuid.UUID) error {
	sqlString := `
		DELETE FROM LOBBY_DECK
		WHERE LOBBY_ID = ?
			AND DECK_ID = ?
	`
	return Execute(sqlString, id, deckId)
}

func DeleteLobby(id uuid.UUID) error {
	sqlString := `
		DELETE FROM LOBBY
		WHERE ID = ?
	`
	return Execute(sqlString, id)
}

func getLobbyDecks(lobbyId uuid.UUID) (deckIds []uuid.UUID, err error) {
	rows, err := Query(`
		SELECT
			DECK_ID
		FROM LOBBY_DECK
		WHERE LOBBY_ID = ?
	`, lobbyId)
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
