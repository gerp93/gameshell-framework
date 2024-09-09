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

type Player struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	Name         string
	PasswordHash string
	ColorTheme   sql.NullString
	IsAdmin      bool
}

func HasLobbyAccess(playerId uuid.UUID, lobbyId uuid.UUID) bool {
	lobbyIds, err := getPlayerLobbyAccess(playerId)
	if err != nil {
		return false
	}
	return helper.IsIdInArray(lobbyId, lobbyIds)
}

func HasDeckAccess(playerId uuid.UUID, deckId uuid.UUID) bool {
	deckIds, err := getPlayerDeckAccess(playerId)
	if err != nil {
		return false
	}
	return helper.IsIdInArray(deckId, deckIds)
}

func GetPlayers(search string) ([]Player, error) {
	if search == "" {
		search = "%"
	}

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			IS_ADMIN
		FROM PLAYER
		WHERE NAME LIKE ?
		ORDER BY CHANGED_ON_DATE DESC
	`
	rows, err := Query(sqlString, search)
	if err != nil {
		return nil, err
	}

	result := make([]Player, 0)
	for rows.Next() {
		var player Player
		if err := rows.Scan(
			&player.Id,
			&player.CreatedOnDate,
			&player.ChangedOnDate,
			&player.Name,
			&player.IsAdmin); err != nil {
			continue
		}
		result = append(result, player)
	}
	return result, nil
}

func GetPlayer(id uuid.UUID) (Player, error) {
	var player Player

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			PASSWORD_HASH,
			COLOR_THEME,
			IS_ADMIN
		FROM PLAYER
		WHERE ID = ?
	`
	rows, err := Query(sqlString, id)
	if err != nil {
		return player, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&player.Id,
			&player.CreatedOnDate,
			&player.ChangedOnDate,
			&player.Name,
			&player.PasswordHash,
			&player.ColorTheme,
			&player.IsAdmin); err != nil {
			log.Println(err)
			return player, errors.New("failed to scan row in query results")
		}
	}

	return player, nil
}

func GetPlayerName(id uuid.UUID) (Player, error) {
	var player Player

	sqlString := `
		SELECT
			ID,
			NAME
		FROM PLAYER
		WHERE ID = ?
	`
	rows, err := Query(sqlString, id)
	if err != nil {
		return player, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&player.Id,
			&player.Name); err != nil {
			log.Println(err)
			return player, errors.New("failed to scan row in query results")
		}
	}

	return player, nil
}

func GetPlayerPasswordHash(id uuid.UUID) (string, error) {
	var passwordHash string

	sqlString := `
		SELECT
			PASSWORD_HASH
		FROM PLAYER
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

func GetPlayerIsAdmin(id uuid.UUID) (bool, error) {
	var isAdmin bool = false

	sqlString := `
		SELECT
			IS_ADMIN
		FROM PLAYER
		WHERE ID = ?
	`
	rows, err := Query(sqlString, id)
	if err != nil {
		return isAdmin, err
	}

	for rows.Next() {
		if err := rows.Scan(&isAdmin); err != nil {
			log.Println(err)
			return isAdmin, errors.New("failed to scan row in query results")
		}
	}

	return isAdmin, nil
}

func GetPlayerId(name string) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM PLAYER
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

func CreatePlayer(name string, password string) (uuid.UUID, error) {
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
		INSERT INTO PLAYER (ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?)
	`
	return id, Execute(sqlString, id, name, passwordHash)
}

func SetPlayerName(id uuid.UUID, name string) error {
	sqlString := `
		UPDATE PLAYER
		SET
			NAME = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`
	return Execute(sqlString, name, id)
}

func SetPlayerPassword(id uuid.UUID, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return errors.New("failed to hash password")
	}

	sqlString := `
		UPDATE PLAYER
		SET
			PASSWORD_HASH = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`
	return Execute(sqlString, passwordHash, id)
}

func SetPlayerColorTheme(id uuid.UUID, colorTheme string) error {
	sqlString := `
		UPDATE PLAYER
		SET
			COLOR_THEME = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`
	if colorTheme == "" {
		return Execute(sqlString, nil, id)
	} else {
		return Execute(sqlString, colorTheme, id)
	}
}

func SetPlayerIsAdmin(id uuid.UUID, isAdmin bool) error {
	sqlString := `
		UPDATE PLAYER
		SET
			IS_ADMIN = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`
	return Execute(sqlString, isAdmin, id)
}

func DeletePlayer(id uuid.UUID) error {
	sqlString := `
		DELETE FROM PLAYER
		WHERE ID = ?
	`
	return Execute(sqlString, id)
}
