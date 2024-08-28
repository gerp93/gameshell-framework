package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Player struct {
	Id           uuid.UUID
	DateAdded    time.Time
	DateModified time.Time

	Name string
}

func GetPlayers(dbcs string) ([]Player, error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		SELECT ID
			 , DATE_ADDED
			 , DATE_MODIFIED
			 , NAME
	 	FROM PLAYER
		ORDER BY DATE_MODIFIED DESC
	`)
	if err != nil {
		return nil, err
	}
	defer statment.Close()

	rows, err := statment.Query()
	if err != nil {
		return nil, err
	}

	result := make([]Player, 0)
	for rows.Next() {
		var player Player
		if err := rows.Scan(
			&player.Id,
			&player.DateAdded,
			&player.DateModified,
			&player.Name); err != nil {
			continue
		}
		result = append(result, player)
	}
	return result, nil
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
			&player.Name); err != nil {
			return player, err
		}
	}

	return player, nil
}

func CreatePlayer(dbcs string, name string) (uuid.UUID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return id, err
	}

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return id, err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		INSERT INTO PLAYER (ID, NAME)
		VALUES (?, ?)
	`)
	if err != nil {
		return id, err
	}
	defer statment.Close()

	_, err = statment.Exec(id, name)
	if err != nil {
		return id, err
	}

	return id, nil
}

func UpdatePlayer(dbcs string, id uuid.UUID, name string) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		UPDATE PLAYER
		SET NAME = ?
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
