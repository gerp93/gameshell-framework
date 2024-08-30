package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
)

type Player struct {
	Id           uuid.UUID
	DateAdded    time.Time
	DateModified time.Time

	Name         string
	PasswordHash string
	ColorTheme   sql.NullString
	LobbyIds     []uuid.UUID
	DeckIds      []uuid.UUID
}

func GetPlayer(dbcs string, id uuid.UUID) (Player, error) {
	var player Player

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return player, err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		SELECT
			ID,
			DATE_ADDED,
			DATE_MODIFIED,
			NAME,
			PASSWORD_HASH,
			COLOR_THEME
		FROM PLAYER
		WHERE ID = ?
	`)
	if err != nil {
		return player, err
	}
	defer statment.Close()

	rows, err := statment.Query(id)
	if err != nil {
		return player, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&player.Id,
			&player.DateAdded,
			&player.DateModified,
			&player.Name,
			&player.PasswordHash,
			&player.ColorTheme); err != nil {
			return player, err
		}
	}

	player.LobbyIds, err = GetPlayerLobbyAccess(dbcs, player.Id)
	if err != nil {
		player.LobbyIds = make([]uuid.UUID, 0)
	}

	player.DeckIds, err = GetPlayerDeckAccess(dbcs, player.Id)
	if err != nil {
		player.DeckIds = make([]uuid.UUID, 0)
	}

	return player, nil
}

func GetPlayerId(dbcs string, name string, password string) (uuid.UUID, error) {
	id := uuid.Nil

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return id, err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		SELECT
			ID,
			PASSWORD_HASH
		FROM PLAYER
		WHERE NAME = ?
	`)
	if err != nil {
		return id, err
	}
	defer statment.Close()

	rows, err := statment.Query(name)
	if err != nil {
		return id, err
	}

	var passwordHash string
	for rows.Next() {
		if err := rows.Scan(&id, &passwordHash); err != nil {
			return id, err
		}
	}

	if !auth.PasswordMatchesHash(password, passwordHash) {
		return id, errors.New("invalid password")
	}

	return id, nil
}

func CreatePlayer(dbcs string, name string, password string) (uuid.UUID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return id, err
	}

	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		return id, err
	}

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return id, err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		INSERT INTO PLAYER (ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return id, err
	}
	defer statment.Close()

	_, err = statment.Exec(id, name, passwordHash)
	if err != nil {
		return id, err
	}

	return id, nil
}

func SetPlayerName(dbcs string, id uuid.UUID, name string) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		UPDATE PLAYER
		SET
			NAME = ?,
			DATE_MODIFIED = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`)
	if err != nil {
		return err
	}
	defer statment.Close()

	_, err = statment.Exec(name, id)
	if err != nil {
		return err
	}

	return nil
}

func SetPlayerPassword(dbcs string, id uuid.UUID, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		UPDATE PLAYER
		SET
			PASSWORD_HASH = ?,
			DATE_MODIFIED = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`)
	if err != nil {
		return err
	}
	defer statment.Close()

	_, err = statment.Exec(passwordHash, id)
	if err != nil {
		return err
	}

	return nil
}

func SetPlayerColorTheme(dbcs string, id uuid.UUID, colorTheme string) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		UPDATE PLAYER
		SET
			COLOR_THEME = ?,
			DATE_MODIFIED = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`)
	if err != nil {
		return err
	}
	defer statment.Close()

	if colorTheme == "" {
		_, err = statment.Exec(nil, id)
	} else {
		_, err = statment.Exec(colorTheme, id)
	}
	if err != nil {
		return err
	}

	return nil
}

func DeletePlayer(dbcs string, id uuid.UUID) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		DELETE FROM PLAYER
		WHERE ID = ?
	`)
	if err != nil {
		return err
	}
	defer statment.Close()

	_, err = statment.Exec(id)
	if err != nil {
		return err
	}

	return nil
}
