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

func UserHasLobbyAccess(userId uuid.UUID, lobbyId uuid.UUID) bool {
	lobbyIds, err := getUserLobbyAccess(userId)
	if err != nil {
		return false
	}
	return helper.IsIdInArray(lobbyId, lobbyIds)
}

func UserHasDeckAccess(userId uuid.UUID, deckId uuid.UUID) bool {
	deckIds, err := getUserDeckAccess(userId)
	if err != nil {
		return false
	}
	return helper.IsIdInArray(deckId, deckIds)
}

func SearchUsers(search string) ([]User, error) {
	if search == "" {
		search = "%"
	}

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
		WHERE NAME LIKE ?
		ORDER BY
			TO_DAYS(CHANGED_ON_DATE) DESC,
			NAME ASC
	`
	rows, err := query(sqlString, search)
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
			&user.PasswordHash,
			&user.ColorTheme,
			&user.IsAdmin); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
		result = append(result, user)
	}
	return result, nil
}

func GetUser(userId uuid.UUID) (user User, err error) {
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
	rows, err := query(sqlString, userId)
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

func GetUserPasswordHash(userId uuid.UUID) (string, error) {
	var passwordHash string

	sqlString := `
		SELECT
			PASSWORD_HASH
		FROM USER
		WHERE ID = ?
	`
	rows, err := query(sqlString, userId)
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

func GetUserIsAdmin(userId uuid.UUID) (bool, error) {
	var isAdmin bool

	sqlString := `
		SELECT
			PASSWORD_HASH
		FROM USER
		WHERE ID = ?
	`
	rows, err := query(sqlString, userId)
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

func GetUserIdByName(name string) (uuid.UUID, error) {
	var userId uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM USER
		WHERE NAME = ?
	`
	rows, err := query(sqlString, name)
	if err != nil {
		return userId, err
	}

	for rows.Next() {
		if err := rows.Scan(&userId); err != nil {
			log.Println(err)
			return userId, errors.New("failed to scan row in query results")
		}
	}

	if userId == uuid.Nil {
		return userId, errors.New("user not found")
	}

	return userId, nil
}

func UserNameExists(name string) bool {
	sqlString := `
		SELECT
			ID
		FROM USER
		WHERE NAME = ?
	`
	rows, err := query(sqlString, name)
	if err != nil {
		return false
	}

	return rows.Next()
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
	err = execute(sqlString, id, name, passwordHash)
	if err != nil {
		return id, err
	}

	return id, nil
}

func SetUserName(id uuid.UUID, name string) error {
	sqlString := `
		UPDATE USER
		SET
			NAME = ?
		WHERE ID = ?
	`
	err := execute(sqlString, name, id)
	if err != nil {
		return err
	}

	return nil
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
	err = execute(sqlString, passwordHash, id)
	if err != nil {
		return err
	}

	return nil
}

func SetUserColorTheme(id uuid.UUID, colorTheme string) (err error) {
	sqlString := `
		UPDATE USER
		SET
			COLOR_THEME = ?
		WHERE ID = ?
	`
	if colorTheme == "" {
		err = execute(sqlString, nil, id)
	} else {
		err = execute(sqlString, colorTheme, id)
	}
	if err != nil {
		return err
	}

	return nil
}

func SetUserIsAdmin(id uuid.UUID, isAdmin bool) error {
	sqlString := `
		UPDATE USER
		SET
			IS_ADMIN = ?
		WHERE ID = ?
	`
	err := execute(sqlString, isAdmin, id)
	if err != nil {
		return err
	}

	return nil
}

func DeleteUser(id uuid.UUID) error {
	sqlString := `
		DELETE FROM USER
		WHERE ID = ?
	`
	err := execute(sqlString, id)
	if err != nil {
		return err
	}

	return nil
}
