package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
)

type Lobby struct {
	Id           uuid.UUID
	DateAdded    time.Time
	DateModified time.Time

	Name         string
	PasswordHash sql.NullString
}

func GetLobbies(dbcs string) ([]Lobby, error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		SELECT
			ID,
			DATE_ADDED,
			DATE_MODIFIED,
			NAME,
			PASSWORD_HASH
		FROM LOBBY
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

	result := make([]Lobby, 0)
	for rows.Next() {
		var lobby Lobby
		if err := rows.Scan(
			&lobby.Id,
			&lobby.DateAdded,
			&lobby.DateModified,
			&lobby.Name,
			&lobby.PasswordHash); err != nil {
			continue
		}
		result = append(result, lobby)
	}
	return result, nil
}

func GetLobby(dbcs string, id uuid.UUID) (Lobby, error) {
	var lobby Lobby

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return lobby, err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		SELECT
			ID,
			DATE_ADDED,
			DATE_MODIFIED,
			NAME,
			PASSWORD_HASH
		FROM LOBBY
		WHERE ID = ?
	`)
	if err != nil {
		return lobby, err
	}
	defer statment.Close()

	rows, err := statment.Query(id)
	if err != nil {
		return lobby, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&lobby.Id,
			&lobby.DateAdded,
			&lobby.DateModified,
			&lobby.Name,
			&lobby.PasswordHash); err != nil {
			return lobby, err
		}
	}

	return lobby, nil
}

func CreateLobby(dbcs string, name string, password string) (uuid.UUID, error) {
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
		INSERT INTO LOBBY (ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return id, err
	}
	defer statment.Close()

	if password == "" {
		_, err = statment.Exec(id, name, nil)
	} else {
		_, err = statment.Exec(id, name, passwordHash)
	}
	if err != nil {
		return id, err
	}

	return id, nil
}

func UpdateLobby(dbcs string, id uuid.UUID, name string, password string) error {
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
		UPDATE LOBBY
		SET
			NAME = ?,
			PASSWORD_HASH = ?,
			DATE_MODIFIED = CURRENT_TIMESTAMP()
		WHERE ID = ?
	`)
	if err != nil {
		return err
	}
	defer statment.Close()

	if password == "" {
		_, err = statment.Exec(name, nil, id)
	} else {
		_, err = statment.Exec(name, passwordHash, id)
	}
	if err != nil {
		return err
	}

	return nil
}

func DeleteLobby(dbcs string, id uuid.UUID) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		return err
	}
	defer db.Close()

	statment, err := db.Prepare(`
		DELETE FROM LOBBY
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
