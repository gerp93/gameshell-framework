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

	statement, err := db.Prepare(`
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
	defer statement.Close()

	rows, err := statement.Query()
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to query statement in database")
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
			CREATED_BY_PLAYER_ID,
			CHANGED_BY_PLAYER_ID,
			NAME,
			PASSWORD_HASH
		FROM LOBBY
		WHERE NAME LIKE ?
		ORDER BY CHANGED_ON_DATE DESC
	`)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	rows, err := statement.Query(search)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to query statement in database")
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

	statement, err := db.Prepare(`
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
	defer statement.Close()

	rows, err := statement.Query(id)
	if err != nil {
		log.Println(err)
		return lobby, errors.New("failed to query statement in database")
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

	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		INSERT INTO LOBBY (ID, CREATED_BY_PLAYER_ID, CHANGED_BY_PLAYER_ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	if password == "" {
		_, err = statement.Exec(id, playerId, playerId, name, nil)
	} else {
		_, err = statement.Exec(id, playerId, playerId, name, passwordHash)
	}
	if err != nil {
		log.Println(err)
		return id, errors.New("failed to execute statement in database")
	}

	return id, nil
}

func GetLobbyId(name string) (uuid.UUID, error) {
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
		FROM LOBBY
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

func SetLobbyName(playerId uuid.UUID, id uuid.UUID, name string) error {
	db, err := sql.Open("mysql", dbcs)
	if err != nil {
		log.Println(err)
		return errors.New("failed to connect to database")
	}
	defer db.Close()

	statement, err := db.Prepare(`
		UPDATE LOBBY
		SET
			NAME = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(),
			CHANGED_BY_PLAYER_ID = ?
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	_, err = statement.Exec(name, playerId, id)
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
	}

	return nil
}

func SetLobbyPassword(playerId uuid.UUID, id uuid.UUID, password string) error {
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
		UPDATE LOBBY
		SET
			PASSWORD_HASH = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(),
			CHANGED_BY_PLAYER_ID = ?
		WHERE ID = ?
	`)
	if err != nil {
		log.Println(err)
		return errors.New("failed to prepare database statement")
	}
	defer statement.Close()

	if password == "" {
		_, err = statement.Exec(nil, playerId, id)
	} else {
		_, err = statement.Exec(passwordHash, playerId, id)
	}
	if err != nil {
		log.Println(err)
		return errors.New("failed to execute statement in database")
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

	statement, err := db.Prepare(`
		DELETE FROM LOBBY
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
