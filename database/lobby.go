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
	JudgePlayer  sql.Null[Player]
	JudgeCard    sql.Null[Card]
	Cards        []Card
	Players      []Player
}

func GetLobbies() ([]Lobby, error) {
	sqlString := `
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
	`
	rows, err := Query(sqlString)
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

func SearchLobbies(search string) ([]Lobby, error) {
	sqlString := `
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
	`
	rows, err := Query(sqlString, search)
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

	sqlString := `
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
	`
	rows, err := Query(sqlString, id)
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
			log.Println(err)
			return lobby, errors.New("failed to scan row in query results")
		}
	}

	lobby.Cards, err = getLobbyCards(lobby.Id)
	if err != nil {
		lobby.Cards = make([]Card, 0)
	}

	lobby.Players, err = GetLobbyPlayers(lobby.Id)
	if err != nil {
		lobby.Players = make([]Player, 0)
	}

	return lobby, nil
}

func GetLobbyPasswordHash(id uuid.UUID) (sql.NullString, error) {
	var passwordHash sql.NullString

	sqlString := `
		SELECT
			PASSWORD_HASH
		FROM LOBBY
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

func GetLobbyPlayers(lobbyId uuid.UUID) (players []Player, err error) {
	sqlString := `
		SELECT
			P.ID,
			P.NAME
		FROM PLAYER AS P
			INNER JOIN LOBBY_PLAYER AS LP ON LP.PLAYER_ID = P.ID
		WHERE LP.LOBBY_ID = ?
	`
	rows, err := Query(sqlString, lobbyId)
	if err != nil {
		return nil, err
	}

	players = make([]Player, 0)
	for rows.Next() {
		var player Player
		if err := rows.Scan(
			&player.Id,
			&player.Name); err != nil {
			continue
		}
		players = append(players, player)
	}

	return players, nil
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

	sqlString := `
		INSERT INTO LOBBY (ID, CREATED_BY_PLAYER_ID, CHANGED_BY_PLAYER_ID, NAME, PASSWORD_HASH)
		VALUES (?, ?, ?, ?, ?)
	`
	if password == "" {
		return id, Execute(sqlString, id, playerId, playerId, name, nil)
	} else {
		return id, Execute(sqlString, id, playerId, playerId, name, passwordHash)
	}
}

func AddCardsToLobby(lobbyId uuid.UUID, deckIds []uuid.UUID) error {
	for _, deckId := range deckIds {
		sqlString := `
			INSERT INTO LOBBY_CARD (LOBBY_ID, CARD_ID)
			SELECT
				? AS LOBBY_ID,
				ID AS CARD_ID
			FROM CARD
			WHERE DECK_ID = ?
		`
		err := Execute(sqlString, lobbyId, deckId)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddPlayerToLobby(lobbyId uuid.UUID, playerId uuid.UUID) error {
	sqlString := `
		INSERT IGNORE INTO LOBBY_PLAYER (LOBBY_ID, PLAYER_ID)
		VALUES (?, ?)
	`
	return Execute(sqlString, lobbyId, playerId)
}

func RemovePlayerFromLobby(lobbyId uuid.UUID, playerId uuid.UUID) error {
	sqlString := `
		DELETE FROM LOBBY_PLAYER
		WHERE LOBBY_ID = ?
			AND PLAYER_ID = ?
	`
	return Execute(sqlString, lobbyId, playerId)
}

func GetLobbyId(name string) (uuid.UUID, error) {
	var id uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM LOBBY
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

func SetLobbyName(playerId uuid.UUID, id uuid.UUID, name string) error {
	sqlString := `
		UPDATE LOBBY
		SET
			NAME = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(),
			CHANGED_BY_PLAYER_ID = ?
		WHERE ID = ?
	`
	return Execute(sqlString, name, playerId, id)
}

func SetLobbyPassword(playerId uuid.UUID, id uuid.UUID, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return errors.New("failed to hash password")
	}

	sqlString := `
		UPDATE LOBBY
		SET
			PASSWORD_HASH = ?,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(),
			CHANGED_BY_PLAYER_ID = ?
		WHERE ID = ?
	`
	if password == "" {
		return Execute(sqlString, nil, playerId, id)
	} else {
		return Execute(sqlString, passwordHash, playerId, id)
	}
}

func DeleteLobby(lobbyId uuid.UUID) error {
	sqlString := `
		DELETE FROM LOBBY
		WHERE ID = ?
	`
	return Execute(sqlString, lobbyId)
}

func getLobbyCards(lobbyId uuid.UUID) (cards []Card, err error) {
	sqlString := `
		SELECT
			C.ID,
			C.TYPE,
			C.TEXT
		FROM CARD AS C
			INNER JOIN LOBBY_CARD AS LC ON LC.CARD_ID = C.ID
		WHERE LC.LOBBY_ID = ?
	`
	rows, err := Query(sqlString, lobbyId)
	if err != nil {
		return nil, err
	}

	cards = make([]Card, 0)
	for rows.Next() {
		var card Card
		if err := rows.Scan(
			&card.Id,
			&card.Type,
			&card.Text); err != nil {
			continue
		}
		cards = append(cards, card)
	}

	return cards, nil
}
