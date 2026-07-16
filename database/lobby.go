package database

import (
	"database/sql"
	"errors"
	"log"

	gameshell "github.com/gerp93/gameshell-framework"
	"github.com/gerp93/gameshell-framework/auth"
	"github.com/google/uuid"
)

func GetLobbyPasswordHash(id uuid.UUID) (sql.NullString, error) {
	var passwordHash sql.NullString

	sqlString := `
		SELECT
			PASSWORD_HASH
		FROM LOBBY
		WHERE ID = ?
	`
	rows, err := query(sqlString, id)
	if err != nil {
		return passwordHash, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&passwordHash); err != nil {
			log.Println(err)
			return passwordHash, errors.New("failed to scan row in query results")
		}
	}

	return passwordHash, nil
}

func CreateLobby(name string, message string, password string) (uuid.UUID, error) {
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
		INSERT INTO LOBBY(
			ID,
			NAME,
			MESSAGE,
			PASSWORD_HASH
		)
		VALUES (?, ?, ?, ?)
	`
	if message == "" {
		if password == "" {
			err = execute(sqlString, id, name, nil, nil)
		} else {
			err = execute(sqlString, id, name, nil, passwordHash)
		}
	} else {
		if password == "" {
			err = execute(sqlString, id, name, message, nil)
		} else {
			err = execute(sqlString, id, name, message, passwordHash)
		}
	}
	if err != nil {
		return id, err
	}

	if g := gameshell.Registered(); g != nil {
		err = g.OnRoomCreated(id)
		if err != nil {
			return id, err
		}
	}

	return id, nil
}

func AddUserToLobby(lobbyId uuid.UUID, userId uuid.UUID) (uuid.UUID, error) {
	player, err := GetLobbyUserPlayer(lobbyId, userId)
	if err != nil {
		log.Println(err)
		return player.Id, errors.New("failed to get player")
	}

	isNewPlayer := player.Id == uuid.Nil
	if isNewPlayer {
		player.Id, err = uuid.NewUUID()
		if err != nil {
			log.Println(err)
			return player.Id, errors.New("failed to generate new player id")
		}
	}

	sqlString := "CALL SP_SET_PLAYER_ACTIVE (?, ?, ?)"
	err = execute(sqlString, player.Id, lobbyId, userId)
	if err != nil {
		return player.Id, err
	}

	if g := gameshell.Registered(); g != nil {
		if isNewPlayer {
			err = g.OnPlayerJoined(player.Id)
		} else if !player.IsActive {
			err = g.OnPlayerActive(player.Id)
		}
	}

	return player.Id, err
}

func SetPlayerInactive(lobbyId uuid.UUID, userId uuid.UUID) error {
	player, err := GetLobbyUserPlayer(lobbyId, userId)
	if err != nil {
		return err
	}

	sqlString := "CALL SP_SET_PLAYER_INACTIVE (?, ?)"
	err = execute(sqlString, lobbyId, userId)
	if err != nil {
		return err
	}

	if player.Id != uuid.Nil && player.IsActive {
		if g := gameshell.Registered(); g != nil {
			return g.OnPlayerInactive(player.Id)
		}
	}

	return nil
}

func GetLobbyId(name string) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM LOBBY
		WHERE NAME = ?
	`
	rows, err := query(sqlString, name)
	if err != nil {
		return id, err
	}
	defer rows.Close()

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
		SET NAME = ?
		WHERE ID = ?
	`
	return execute(sqlString, name, id)
}

func SetLobbyMessage(id uuid.UUID, message string) error {
	sqlString := `
		UPDATE LOBBY
		SET MESSAGE = ?
		WHERE ID = ?
	`
	if message == "" {
		return execute(sqlString, nil, id)
	} else {
		return execute(sqlString, message, id)
	}
}

func DeleteLobby(lobbyId uuid.UUID) error {
	sqlString := `
		DELETE
		FROM LOBBY
		WHERE ID = ?
	`
	return execute(sqlString, lobbyId)
}
