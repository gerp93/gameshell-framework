package database

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

type Player struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	Name         string
	LobbyId      uuid.UUID
	UserId       uuid.UUID
	IsActive     bool
	CreditsSpent int
}

func GetPlayer(lobbyId uuid.UUID, userId uuid.UUID) (Player, error) {
	var player Player

	sqlString := `
		SELECT
			P.ID,
			P.CREATED_ON_DATE,
			P.CHANGED_ON_DATE,
			U.NAME,
			P.LOBBY_ID,
			P.USER_ID,
			P.IS_ACTIVE,
			P.CREDITS_SPENT
		FROM PLAYER AS P
			INNER JOIN USER AS U ON U.ID = P.USER_ID
		WHERE P.LOBBY_ID = ?
			AND P.USER_ID = ?
	`
	rows, err := query(sqlString, lobbyId, userId)
	if err != nil {
		return player, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&player.Id,
			&player.CreatedOnDate,
			&player.ChangedOnDate,
			&player.Name,
			&player.LobbyId,
			&player.UserId,
			&player.IsActive,
			&player.CreditsSpent); err != nil {
			log.Println(err)
			return player, errors.New("failed to scan row in query results")
		}
	}

	return player, nil
}
