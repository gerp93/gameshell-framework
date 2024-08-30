package database

import (
	"database/sql"
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
}

func GetPlayer(dbcs string, id uuid.UUID) (Player, error) {
	var player Player

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return player, err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		SELECT ID
			 , DATE_ADDED
			 , DATE_MODIFIED
			 , NAME
			 , PASSWORD_HASH
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
			&player.PasswordHash); err != nil {
			return player, err
		}
	}

	return player, nil
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

func UpdatePlayer(dbcs string, id uuid.UUID, name string, password string) error {
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
		SET NAME = ?,
		    PASSWORD_HASH = ?
		WHERE ID = ?
	`)
	if err != nil {
		return err
	}
	defer statment.Close()

	_, err = statment.Exec(name, passwordHash, id)
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
