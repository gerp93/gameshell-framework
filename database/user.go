package database

import (
	"database/sql"
	"errors"
	"log"
	"strings"
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
			PASSWORD_HASH,
			COLOR_THEME,
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

func getUserInDB(userId uuid.UUID) (user User, err error) {
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
	rows, err := Query(sqlString, userId)
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

func GetUser(id uuid.UUID) (User, error) {
	user, ok := allUsers[id]
	if !ok {
		return user, errors.New("user not found")
	}
	return user, nil
}

func GetUserPasswordHash(id uuid.UUID) (string, error) {
	user, ok := allUsers[id]
	if !ok {
		return "", errors.New("user not found")
	}
	return user.PasswordHash, nil
}

func GetUserIsAdmin(id uuid.UUID) (bool, error) {
	user, ok := allUsers[id]
	if !ok {
		return false, errors.New("user not found")
	}
	return user.IsAdmin, nil
}

func GetUserIdByName(name string) uuid.UUID {
	for id, user := range allUsers {
		if strings.EqualFold(user.Name, name) {
			return id
		}
	}
	return uuid.Nil
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
	err = Execute(sqlString, id, name, passwordHash)
	if err != nil {
		return id, err
	}

	allUsers[id], err = getUserInDB(id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func SetUserName(id uuid.UUID, name string) error {
	user, ok := allUsers[id]
	if !ok {
		return errors.New("user not found")
	}

	sqlString := `
		UPDATE USER
		SET
			NAME = ?
		WHERE ID = ?
	`
	err := Execute(sqlString, name, id)
	if err != nil {
		return err
	}

	user.Name = name
	allUsers[id] = user

	return nil
}

func SetUserPassword(id uuid.UUID, password string) error {
	user, ok := allUsers[id]
	if !ok {
		return errors.New("user not found")
	}

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
	err = Execute(sqlString, passwordHash, id)
	if err != nil {
		return err
	}

	user.PasswordHash = passwordHash
	allUsers[id] = user

	return nil
}

func SetUserColorTheme(id uuid.UUID, colorTheme string) (err error) {
	user, ok := allUsers[id]
	if !ok {
		return errors.New("user not found")
	}

	sqlString := `
		UPDATE USER
		SET
			COLOR_THEME = ?
		WHERE ID = ?
	`
	if colorTheme == "" {
		err = Execute(sqlString, nil, id)
		user.ColorTheme = sql.NullString{}
	} else {
		err = Execute(sqlString, colorTheme, id)
		user.ColorTheme = sql.NullString{String: colorTheme, Valid: true}
	}
	if err != nil {
		return err
	}

	allUsers[id] = user

	return nil
}

func SetUserIsAdmin(id uuid.UUID, isAdmin bool) error {
	user, ok := allUsers[id]
	if !ok {
		return errors.New("user not found")
	}

	sqlString := `
		UPDATE USER
		SET
			IS_ADMIN = ?
		WHERE ID = ?
	`
	err := Execute(sqlString, isAdmin, id)
	if err != nil {
		return err
	}

	user.IsAdmin = isAdmin
	allUsers[id] = user

	return nil
}

func DeleteUser(id uuid.UUID) error {
	sqlString := `
		DELETE FROM USER
		WHERE ID = ?
	`
	err := Execute(sqlString, id)
	if err != nil {
		return err
	}

	delete(allUsers, id)

	return nil
}
