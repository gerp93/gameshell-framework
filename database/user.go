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

type User struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	Name         string
	PasswordHash string
	ColorTheme   sql.NullString
	IsAdmin      bool
}

func HasLobbyAccess(userId uuid.UUID, lobbyId uuid.UUID) bool {
	lobbyIds, err := getUserLobbyAccess(userId)
	if err != nil {
		return false
	}
	return helper.IsIdInArray(lobbyId, lobbyIds)
}

func HasDeckAccess(userId uuid.UUID, deckId uuid.UUID) bool {
	deckIds, err := getUserDeckAccess(userId)
	if err != nil {
		return false
	}
	return helper.IsIdInArray(deckId, deckIds)
}

func GetUsers(search string) ([]User, error) {
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
		FROM USER
		WHERE NAME LIKE ?
		ORDER BY
			TO_DAYS(CHANGED_ON_DATE) DESC,
			NAME ASC
	`
	rows, err := Query(sqlString, search)
	if err != nil {
		return nil, err
	}

	result := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.Id,
			&user.CreatedOnDate,
			&user.ChangedOnDate,
			&user.Name,
			&user.IsAdmin); err != nil {
			continue
		}
		result = append(result, user)
	}
	return result, nil
}

func GetUser(id uuid.UUID) (User, error) {
	var user User

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			PASSWORD_HASH,
			COLOR_THEME,
			IS_ADMIN
		FROM USER
		WHERE ID = ?
	`
	rows, err := Query(sqlString, id)
	if err != nil {
		return user, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&user.Id,
			&user.CreatedOnDate,
			&user.ChangedOnDate,
			&user.Name,
			&user.PasswordHash,
			&user.ColorTheme,
			&user.IsAdmin); err != nil {
			log.Println(err)
			return user, errors.New("failed to scan row in query results")
		}
	}

	return user, nil
}

func GetUserName(id uuid.UUID) (User, error) {
	var user User

	sqlString := `
		SELECT
			ID,
			NAME
		FROM USER
		WHERE ID = ?
	`
	rows, err := Query(sqlString, id)
	if err != nil {
		return user, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&user.Id,
			&user.Name); err != nil {
			log.Println(err)
			return user, errors.New("failed to scan row in query results")
		}
	}

	return user, nil
}

func GetUserPasswordHash(id uuid.UUID) (string, error) {
	var passwordHash string

	sqlString := `
		SELECT
			PASSWORD_HASH
		FROM USER
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

func GetUserIsAdmin(id uuid.UUID) (bool, error) {
	var isAdmin bool = false

	sqlString := `
		SELECT
			IS_ADMIN
		FROM USER
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

func GetUserId(name string) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM USER
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

func CreateUser(name string, password string) (uuid.UUID, error) {
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
		INSERT INTO USER (ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?)
	`
	return id, Execute(sqlString, id, name, passwordHash)
}

func SetUserName(id uuid.UUID, name string) error {
	sqlString := `
		UPDATE USER
		SET
			NAME = ?
		WHERE ID = ?
	`
	return Execute(sqlString, name, id)
}

func SetUserPassword(id uuid.UUID, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return errors.New("failed to hash password")
	}

	sqlString := `
		UPDATE USER
		SET
			PASSWORD_HASH = ?
		WHERE ID = ?
	`
	return Execute(sqlString, passwordHash, id)
}

func SetUserColorTheme(id uuid.UUID, colorTheme string) error {
	sqlString := `
		UPDATE USER
		SET
			COLOR_THEME = ?
		WHERE ID = ?
	`
	if colorTheme == "" {
		return Execute(sqlString, nil, id)
	} else {
		return Execute(sqlString, colorTheme, id)
	}
}

func SetUserIsAdmin(id uuid.UUID, isAdmin bool) error {
	sqlString := `
		UPDATE USER
		SET
			IS_ADMIN = ?
		WHERE ID = ?
	`
	return Execute(sqlString, isAdmin, id)
}

func DeleteUser(id uuid.UUID) error {
	sqlString := `
		DELETE FROM USER
		WHERE ID = ?
	`
	return Execute(sqlString, id)
}
