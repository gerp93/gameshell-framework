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
	LobbyIds     []uuid.UUID
	DeckIds      []uuid.UUID
}

func (p Player) HasLobbyAccess(lobbyId uuid.UUID) bool {
	return helper.IsIdInArray(lobbyId, p.LobbyIds)
}

func (p Player) HasDeckAccess(deckId uuid.UUID) bool {
	return helper.IsIdInArray(deckId, p.DeckIds)
}

func HasLobbyAccess(playerId uuid.UUID, lobbyId uuid.UUID) bool {
	lobbyIds, err := GetPlayerLobbyAccess(playerId)
	if err != nil {
		return false
	}
	return helper.IsIdInArray(lobbyId, lobbyIds)
}

func HasDeckAccess(playerId uuid.UUID, deckId uuid.UUID) bool {
	deckIds, err := GetPlayerDeckAccess(playerId)
	if err != nil {
		return false
	}
	return helper.IsIdInArray(deckId, deckIds)
}

func GetPlayers() ([]Player, error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			IS_ADMIN
		FROM PLAYER
		ORDER BY CHANGED_ON_DATE DESC
	`)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query()
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to query statement in database")
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

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return player, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
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
	`)
	if err != nil {
		log.Println(err)
		return player, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query(id)
	if err != nil {
		log.Println(err)
		return player, errors.New("failed to query statement in database")
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

	player.LobbyIds, err = GetPlayerLobbyAccess(player.Id)
	if err != nil {
		player.LobbyIds = make([]uuid.UUID, 0)
	}

	player.DeckIds, err = GetPlayerDeckAccess(player.Id)
	if err != nil {
		player.DeckIds = make([]uuid.UUID, 0)
	}

	return player, nil
}

func GetPlayerPasswordHash(id uuid.UUID) (string, error) {
	var passwordHash string

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return passwordHash, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		SELECT
			PASSWORD_HASH
		FROM PLAYER
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return passwordHash, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query(id)
	if err != nil {
		log.Println(err)
		return passwordHash, errors.New("failed to query statement in database")
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

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return isAdmin, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		SELECT
			IS_ADMIN
		FROM PLAYER
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return isAdmin, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query(id)
	if err != nil {
		log.Println(err)
		return isAdmin, errors.New("failed to query statement in database")
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

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		SELECT
			ID
		FROM PLAYER
		WHERE NAME = ?
	`)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query(name)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to query statement in database")
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

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		INSERT INTO PLAYER (ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(id, name, passwordHash)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to execute statement in database")
	}

	return id, nil
}

func SetPlayerName(id uuid.UUID, name string) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		UPDATE PLAYER
		SET
			NAME = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(name, id)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}

func SetPlayerPassword(id uuid.UUID, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return errors.New("failed to hash password")
	}

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		UPDATE PLAYER
		SET
			PASSWORD_HASH = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(passwordHash, id)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}

func SetPlayerColorTheme(id uuid.UUID, colorTheme string) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		UPDATE PLAYER
		SET
			COLOR_THEME = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	if colorTheme == "" {
		_, err = statement.Exec(nil, id)
	} else {
		_, err = statement.Exec(colorTheme, id)
	}
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}

func SetPlayerIsAdmin(id uuid.UUID, isAdmin bool) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		UPDATE PLAYER
		SET
			IS_ADMIN = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(isAdmin, id)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}

func DeletePlayer(id uuid.UUID) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		DELETE FROM PLAYER
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}
