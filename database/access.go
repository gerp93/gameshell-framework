package database

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

func UserHasLobbyAccess(userId uuid.UUID, lobbyId uuid.UUID) (bool, error) {
	sqlString := "SELECT FN_USER_HAS_LOBBY_ACCESS (?, ?)"
	rows, err := query(sqlString, userId, lobbyId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	hasAccess := false
	for rows.Next() {
		if err := rows.Scan(&hasAccess); err != nil {
			log.Println(err)
			return false, errors.New("failed to scan row in query results")
		}
	}

	return hasAccess, nil
}

func AddUserLobbyAccess(userId uuid.UUID, lobbyId uuid.UUID) error {
	sqlString := `
		INSERT INTO USER_ACCESS_LOBBY(USER_ID, LOBBY_ID)
		VALUES (?, ?)
	`
	return execute(sqlString, userId, lobbyId)
}

func UserHasDeckAccess(userId uuid.UUID, deckId uuid.UUID) (bool, error) {
	sqlString := "SELECT FN_USER_HAS_DECK_ACCESS (?, ?)"
	rows, err := query(sqlString, userId, deckId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	hasAccess := false
	for rows.Next() {
		if err := rows.Scan(&hasAccess); err != nil {
			log.Println(err)
			return false, errors.New("failed to scan row in query results")
		}
	}

	return hasAccess, nil
}

// UserCanReadDeck reports whether userId may use/view deckId — unlike
// UserHasDeckAccess, this also allows a deck flagged IS_PUBLIC_READONLY, the
// same condition SP_GET_READABLE_DECKS already lists it under. Use this for
// read/use checks (selecting a deck into a lobby, viewing its detail page);
// keep using UserHasDeckAccess for anything that edits or deletes a deck or
// its cards, since public-readonly must never grant that.
func UserCanReadDeck(userId uuid.UUID, deckId uuid.UUID) (bool, error) {
	sqlString := "SELECT FN_USER_CAN_READ_DECK (?, ?)"
	rows, err := query(sqlString, userId, deckId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	canRead := false
	for rows.Next() {
		if err := rows.Scan(&canRead); err != nil {
			log.Println(err)
			return false, errors.New("failed to scan row in query results")
		}
	}

	return canRead, nil
}

func AddUserDeckAccess(userId uuid.UUID, deckId uuid.UUID) error {
	sqlString := `
		INSERT INTO USER_ACCESS_DECK(USER_ID, DECK_ID)
		VALUES (?, ?)
	`
	return execute(sqlString, userId, deckId)
}
