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
	Id                uuid.UUID
	CreatedOnDate     time.Time
	ChangedOnDate     time.Time
	CreatedByPlayerId uuid.UUID
	ChangedByPlayerId uuid.UUID

	Name         string
	PasswordHash sql.NullString
}

func GetLobbies() ([]Lobby, error) {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to connect to database")
	}
	defer db.Close()

	statment, err := db.Prepare(`
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
		log.Println(err)
		return nil, errors.New("failed to prepare database statement")
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

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return lobby, errors.New("failed to connect to database")
	}
	defer db.Close()

	statment, err := db.Prepare(`
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
	`)
	if err != nil {
		log.Println(err)
		return lobby, errors.New("failed to prepare database statement")
	}
	defer statment.Close()

	rows, err := statment.Query(id)
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
			return lobby, err
		}
	}

	return lobby, nil
}

func CreateLobby(playerId uuid.UUID, name string, password string) (uuid.UUID, error) {
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
		log.Println(err)
		return id, errors.New("failed to connect to database")
	}
	defer db.Close()

	statment, err := db.Prepare(`
		INSERT INTO LOBBY (ID, CREATED_BY_PLAYER_ID, CHANGED_BY_PLAYER_ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to prepare database statement")
	}
	defer statment.Close()

	if password == "" {
		_, err = statment.Exec(id, playerId, playerId, name, nil)
	} else {
		_, err = statment.Exec(id, playerId, playerId, name, passwordHash)
	}
	if err != nil {
		return id, err
	}

	return id, nil
}

func UpdateLobby(playerId uuid.UUID, id uuid.UUID, name string, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statment, err := db.Prepare(`
		UPDATE LOBBY
		SET
			NAME = ?,
			PASSWORD_HASH = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(),
			CHANGED_BY_PLAYER_ID = ?
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statment.Close()

	if password == "" {
		_, err = statment.Exec(name, nil, playerId, id)
	} else {
		_, err = statment.Exec(name, passwordHash, playerId, id)
	}
	if err != nil {
		return err
	}

	return nil
}

func DeleteLobby(id uuid.UUID) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statment, err := db.Prepare(`
		DELETE FROM LOBBY
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statment.Close()

	_, err = statment.Exec(id)
	if err != nil {
		return err
	}

	return nil
}
