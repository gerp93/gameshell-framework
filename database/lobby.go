package database

import (
	"database/sql"
	"errors"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
)

type Lobby struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	Name         string
	PasswordHash sql.NullString

	HandSize    int
	CreditLimit int
}

type lobbyDetails struct {
	Lobby
	UserCount int
}

type GameData struct {
	LobbyId          uuid.UUID
	LobbyName        string
	LobbyHandSize    int
	LobbyCreditLimit int

	DrawPilePromptCount   int
	DrawPileResponseCount int

	TotalPlayerCount int

	JudgeName          sql.NullString
	JudgeCardText      sql.NullString
	JudgeBlankCount    int
	JudgeResponseCount int

	BoardIsReady   bool
	BoardIsEmpty   bool
	BoardResponses []boardResponse

	PlayerId               uuid.UUID
	PlayerIsJudge          bool
	PlayerIsReady          bool
	PlayerHand             []handCard
	PlayerResponses        []boardResponse
	PlayerLosingStreak     int
	PlayerCreditsSpent     int
	PlayerExtraResponses   int
	PlayerCreditsRemaining int

	Wins []winDetail

	UpcomingJudges []string
}

type boardResponse struct {
	ResponseId     uuid.UUID
	IsRevealed     bool
	PlayerId       uuid.UUID
	PlayerUserName string
	ResponseCards  []boardResponseCard
}

type boardResponseCard struct {
	ResponseCardId uuid.UUID
	Card
	DeckName        string
	SpecialCategory sql.NullString
}

type handCard struct {
	Card
	DeckName string
	IsLocked bool
}

type winDetail struct {
	UserName string
	WinCount int
}

func SearchLobbies(search string) ([]lobbyDetails, error) {
	sqlString := `
		SELECT
			L.ID,
			L.CREATED_ON_DATE,
			L.CHANGED_ON_DATE,
			L.NAME,
			L.PASSWORD_HASH,
			L.HAND_SIZE,
			L.CREDIT_LIMIT,
			COUNT(P.ID) AS USER_COUNT
		FROM LOBBY AS L
			INNER JOIN PLAYER AS P ON P.LOBBY_ID = L.ID AND P.IS_ACTIVE = 1
		WHERE L.NAME LIKE ?
		GROUP BY L.ID
		ORDER BY
			TO_DAYS(L.CHANGED_ON_DATE) DESC,
			L.NAME ASC,
			COUNT(P.ID) DESC
	`
	rows, err := query(sqlString, search)
	if err != nil {
		return nil, err
	}

	result := make([]lobbyDetails, 0)
	for rows.Next() {
		var ld lobbyDetails
		if err := rows.Scan(
			&ld.Id,
			&ld.CreatedOnDate,
			&ld.ChangedOnDate,
			&ld.Name,
			&ld.PasswordHash,
			&ld.HandSize,
			&ld.CreditLimit,
			&ld.UserCount); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
		result = append(result, ld)
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
			NAME,
			PASSWORD_HASH,
			HAND_SIZE,
			CREDIT_LIMIT
		FROM LOBBY
		WHERE ID = ?
	`
	rows, err := query(sqlString, id)
	if err != nil {
		return lobby, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&lobby.Id,
			&lobby.CreatedOnDate,
			&lobby.ChangedOnDate,
			&lobby.Name,
			&lobby.PasswordHash,
			&lobby.HandSize,
			&lobby.CreditLimit); err != nil {
			log.Println(err)
			return lobby, errors.New("failed to scan row in query results")
		}
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
	rows, err := query(sqlString, id)
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

func CreateLobby(name string, password string, handSize int, creditLimit int) (uuid.UUID, error) {
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
		INSERT INTO LOBBY (ID, NAME, PASSWORD_HASH, HAND_SIZE, CREDIT_LIMIT)
		VALUES (?, ?, ?, ?, ?)
	`
	if password == "" {
		return id, execute(sqlString, id, name, nil, handSize, creditLimit)
	} else {
		return id, execute(sqlString, id, name, passwordHash, handSize, creditLimit)
	}
}

func AddCardsToLobby(lobbyId uuid.UUID, deckIds []uuid.UUID) error {
	for _, deckId := range deckIds {
		sqlString := `
			INSERT INTO DRAW_PILE (LOBBY_ID, CARD_ID)
			SELECT
				? AS LOBBY_ID,
				ID AS CARD_ID
			FROM CARD
			WHERE DECK_ID = ?
		`
		err := execute(sqlString, lobbyId, deckId)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddUserToLobby(lobbyId uuid.UUID, userId uuid.UUID) (uuid.UUID, error) {
	player, err := GetPlayer(lobbyId, userId)
	if err != nil {
		log.Println(err)
		return player.Id, errors.New("failed to get player")
	}

	if player.Id == uuid.Nil {
		player.Id, err = uuid.NewUUID()
		if err != nil {
			log.Println(err)
			return player.Id, errors.New("failed to generate new player id")
		}
	}

	sqlString := "CALL SP_SET_PLAYER_ACTIVE (?, ?, ?)"
	err = execute(sqlString, player.Id, lobbyId, userId)
	return player.Id, err
}

func RemoveUserFromLobby(lobbyId uuid.UUID, userId uuid.UUID) error {
	sqlString := "CALL SP_SET_PLAYER_INACTIVE (?, ?)"
	return execute(sqlString, lobbyId, userId)
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
		SET
			NAME = ?
		WHERE ID = ?
	`
	return execute(sqlString, name, id)
}

func SetLobbyHandSize(id uuid.UUID, handSize int) error {
	sqlString := `
		UPDATE LOBBY
		SET
			HAND_SIZE = ?
		WHERE ID = ?
	`
	return execute(sqlString, handSize, id)
}

func SetLobbyCreditLimit(id uuid.UUID, creditLimit int) error {
	sqlString := `
		UPDATE LOBBY
		SET
			CREDIT_LIMIT = ?
		WHERE ID = ?
	`
	return execute(sqlString, creditLimit, id)
}

func DeleteLobby(lobbyId uuid.UUID) error {
	sqlString := `
		DELETE FROM LOBBY
		WHERE ID = ?
	`
	return execute(sqlString, lobbyId)
}

func GetPlayerGameData(playerId uuid.UUID) (GameData, error) {
	var data GameData

	sqlString := `
		SELECT
			L.ID                                                AS LOBBY_ID,
			L.NAME                                              AS LOBBY_NAME,
			L.HAND_SIZE                                         AS LOBBY_HAND_SIZE,
			L.CREDIT_LIMIT                                      AS LOBBY_CREDIT_LIMIT,
			(SELECT COUNT(*)
				FROM DRAW_PILE AS DP
						INNER JOIN CARD AS DPC ON DPC.ID = DP.CARD_ID
				WHERE DP.LOBBY_ID = L.ID
				AND DPC.CATEGORY = 'PROMPT')                    AS DRAW_PILE_PROMPT_COUNT,
			(SELECT COUNT(*)
				FROM DRAW_PILE AS DP
						INNER JOIN CARD AS DPC ON DPC.ID = DP.CARD_ID
				WHERE DP.LOBBY_ID = L.ID
				AND DPC.CATEGORY = 'RESPONSE')                  AS DRAW_PILE_RESPONSE_COUNT,
			(SELECT COUNT(*) FROM PLAYER WHERE LOBBY_ID = L.ID) AS TOTAL_PLAYER_COUNT,
			(SELECT JU.NAME
				FROM USER AS JU
						INNER JOIN PLAYER AS JP ON JP.USER_ID = JU.ID
				WHERE JP.ID = J.PLAYER_ID)                      AS JUDGE_NAME,
			(SELECT TEXT FROM CARD WHERE ID = J.CARD_ID)        AS JUDGE_CARD_TEXT,
			J.BLANK_COUNT                                       AS JUDGE_BLANK_COUNT,
			J.RESPONSE_COUNT                                    AS JUDGE_RESPONSE_COUNT,
			P.ID                                                AS PLAYER_ID,
			IF(FN_GET_LOBBY_JUDGE_PLAYER_ID(L.ID) = P.ID, 1, 0) AS PLAYER_IS_JUDGE,
			P.LOSING_STREAK                                     AS PLAYER_LOSING_STREAK,
			P.CREDITS_SPENT                                     AS PLAYER_CREDITS_SPENT,
			P.EXTRA_RESPONSES                                   AS PLAYER_EXTRA_RESPONSES
		FROM PLAYER AS P
				INNER JOIN LOBBY AS L ON L.ID = P.LOBBY_ID
				INNER JOIN JUDGE AS J ON J.LOBBY_ID = L.ID
		WHERE P.ID = ?
	`
	rows, err := query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&data.LobbyId,
			&data.LobbyName,
			&data.LobbyHandSize,
			&data.LobbyCreditLimit,
			&data.DrawPilePromptCount,
			&data.DrawPileResponseCount,
			&data.TotalPlayerCount,
			&data.JudgeName,
			&data.JudgeCardText,
			&data.JudgeBlankCount,
			&data.JudgeResponseCount,
			&data.PlayerId,
			&data.PlayerIsJudge,
			&data.PlayerLosingStreak,
			&data.PlayerCreditsSpent,
			&data.PlayerExtraResponses); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
	}

	data.PlayerCreditsRemaining = data.LobbyCreditLimit - data.PlayerCreditsSpent
	if data.PlayerCreditsRemaining < 0 {
		data.PlayerCreditsRemaining = 0
	}

	sqlString = `
		SELECT
			R.ID          AS RESPONSE_ID,
			R.IS_REVEALED AS IS_REVEALED,
			P.ID          AS PLAYER_ID,
			U.NAME        AS PLAYER_USER_NAME
		FROM LOBBY AS L
				INNER JOIN PLAYER AS P ON P.LOBBY_ID = L.ID
				INNER JOIN USER AS U ON U.ID = P.USER_ID
				INNER JOIN RESPONSE AS R ON R.PLAYER_ID = P.ID
				LEFT JOIN JUDGE AS J ON J.PLAYER_ID = P.ID
		WHERE L.ID = ?
			AND P.IS_ACTIVE = 1
			AND J.ID IS NULL
		ORDER BY U.NAME, R.CREATED_ON_DATE
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var br boardResponse
		if err := rows.Scan(
			&br.ResponseId,
			&br.IsRevealed,
			&br.PlayerId,
			&br.PlayerUserName); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.BoardResponses = append(data.BoardResponses, br)
	}

	totalCardsPlayedCount := 0
	for i, br := range data.BoardResponses {
		sqlString = `
			SELECT
				RC.ID  AS RESPONSE_CARD_ID,
				C.ID   AS CARD_ID,
				C.TEXT AS CARD_TEXT,
				D.NAME AS DECK_NAME,
				RC.SPECIAL_CATEGORY
			FROM RESPONSE AS R
					INNER JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
					INNER JOIN CARD AS C ON C.ID = RC.CARD_ID
					INNER JOIN DECK AS D ON D.ID = C.DECK_ID
			WHERE R.ID = ?
			ORDER BY RC.CREATED_ON_DATE
		`
		rows, err = query(sqlString, br.ResponseId)
		if err != nil {
			return data, err
		}

		for rows.Next() {
			var responseCard boardResponseCard
			if err := rows.Scan(
				&responseCard.ResponseCardId,
				&responseCard.Id,
				&responseCard.Text,
				&responseCard.DeckName,
				&responseCard.SpecialCategory); err != nil {
				log.Println(err)
				return data, errors.New("failed to scan row in query results")
			}
			data.BoardResponses[i].ResponseCards = append(data.BoardResponses[i].ResponseCards, responseCard)

			totalCardsPlayedCount += 1
		}

		if br.PlayerId == playerId {
			data.PlayerResponses = append(data.PlayerResponses, data.BoardResponses[i])
		}
	}

	data.BoardIsEmpty = totalCardsPlayedCount == 0
	data.BoardIsReady = !data.BoardIsEmpty

	sqlString = `
		SELECT P.ID AS PLAYER_ID,
			IF(COUNT(RC.ID) = -- CARDS PLAYED
				(J.BLANK_COUNT * -- CARDS PER RESPONSE
				(J.RESPONSE_COUNT + P.EXTRA_RESPONSES)) -- PLAYER RESPONSES
				, 1, 0) AS PLAYER_IS_READY
		FROM RESPONSE AS R
				LEFT JOIN RESPONSE_CARD AS RC ON RC.RESPONSE_ID = R.ID
				INNER JOIN PLAYER AS P ON P.ID = R.PLAYER_ID
				INNER JOIN JUDGE AS J ON J.LOBBY_ID = P.LOBBY_ID
		WHERE J.LOBBY_ID = ?
		GROUP BY P.ID
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var playerId uuid.UUID
		var isReady bool
		if err := rows.Scan(&playerId, &isReady); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}

		if !isReady {
			data.BoardIsReady = false
		}

		if playerId == data.PlayerId {
			data.PlayerIsReady = isReady
		}
	}

	if data.BoardIsReady {
		sort.Slice(data.BoardResponses, func(i, j int) bool {
			if len(data.BoardResponses[i].ResponseCards) == 0 {
				return true
			}
			if len(data.BoardResponses[j].ResponseCards) == 0 {
				return false
			}
			return data.BoardResponses[i].ResponseCards[0].Text < data.BoardResponses[j].ResponseCards[0].Text
		})
	}

	sqlString = `
		SELECT
			C.ID,
			C.TEXT,
			D.NAME,
			H.IS_LOCKED
		FROM HAND AS H
			INNER JOIN CARD AS C ON C.ID = H.CARD_ID
			INNER JOIN DECK AS D ON D.ID = C.DECK_ID
		WHERE H.PLAYER_ID = ?
		ORDER BY H.IS_LOCKED DESC, C.TEXT
	`
	rows, err = query(sqlString, playerId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var card handCard
		if err := rows.Scan(
			&card.Id,
			&card.Text,
			&card.DeckName,
			&card.IsLocked); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.PlayerHand = append(data.PlayerHand, card)
	}

	sqlString = `
		SELECT
			U.NAME      AS USER_NAME,
			COUNT(W.ID) AS WINS
		FROM PLAYER AS P
				INNER JOIN USER AS U ON U.ID = P.USER_ID
				LEFT JOIN WIN AS W ON W.PLAYER_ID = P.ID
		WHERE P.LOBBY_ID = ?
			AND P.IS_ACTIVE = 1
		GROUP BY P.USER_ID
		ORDER BY COUNT(W.ID) DESC, U.NAME ASC
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var win winDetail
		if err := rows.Scan(
			&win.UserName,
			&win.WinCount); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.Wins = append(data.Wins, win)
	}

	sqlString = `
		SELECT U.NAME
		FROM LOBBY AS L
				INNER JOIN JUDGE AS J ON J.LOBBY_ID = L.ID
				INNER JOIN (SELECT
								LOBBY_ID,
								USER_ID,
								RANK() OVER (PARTITION BY LOBBY_ID ORDER BY CREATED_ON_DATE) AS JOIN_ORDER
							FROM PLAYER
							WHERE IS_ACTIVE = 1) AS T ON T.LOBBY_ID = L.ID
				INNER JOIN USER AS U ON U.ID = T.USER_ID
		WHERE L.ID = ?
		ORDER BY T.JOIN_ORDER <= J.POSITION, T.JOIN_ORDER
	`
	rows, err = query(sqlString, data.LobbyId)
	if err != nil {
		return data, err
	}

	for rows.Next() {
		var judgeName string
		if err := rows.Scan(&judgeName); err != nil {
			log.Println(err)
			return data, errors.New("failed to scan row in query results")
		}
		data.UpcomingJudges = append(data.UpcomingJudges, judgeName)
	}

	return data, nil
}

func DrawHand(playerId uuid.UUID) error {
	sqlString := "CALL SP_DRAW_HAND (?)"
	return execute(sqlString, playerId)
}

func PlayCard(playerId uuid.UUID, cardId uuid.UUID) error {
	sqlString := "CALL SP_RESPOND_WITH_CARD (?, ?, NULL)"
	return execute(sqlString, playerId, cardId)
}

func AddExtraResponse(playerId uuid.UUID) error {
	sqlString := "CALL SP_ADD_EXTRA_RESPONSE (?)"
	return execute(sqlString, playerId)
}

func PlayStealCard(playerId uuid.UUID) error {
	sqlString := "CALL SP_RESPOND_WITH_STEAL_CARD (?)"
	return execute(sqlString, playerId)
}

func PlaySurpriseCard(playerId uuid.UUID) error {
	sqlString := "CALL SP_RESPOND_WITH_SURPRISE_CARD (?)"
	return execute(sqlString, playerId)
}

func PlayWildCard(playerId uuid.UUID, text string) error {
	sqlString := "CALL SP_RESPOND_WITH_WILD_CARD (?, ?)"
	return execute(sqlString, playerId, text)
}

func WithdrawCard(responseCardId uuid.UUID) error {
	sqlString := "CALL SP_WITHDRAW_CARD (?)"
	return execute(sqlString, responseCardId)
}

func DiscardCard(playerId uuid.UUID, cardId uuid.UUID) error {
	sqlString := "CALL SP_DISCARD_CARD (?, ?)"
	return execute(sqlString, playerId, cardId)
}

func LockCard(playerId uuid.UUID, cardId uuid.UUID, isLocked bool) error {
	sqlString := `
		UPDATE HAND
		SET IS_LOCKED = ?
		WHERE PLAYER_ID = ?
			AND CARD_ID = ?
	`
	return execute(sqlString, isLocked, playerId, cardId)
}

func RevealResponse(responseId uuid.UUID) error {
	sqlString := `
		UPDATE RESPONSE
		SET IS_REVEALED = 1
		WHERE ID = ?
	`
	return execute(sqlString, responseId)
}

func PickWinner(responseId uuid.UUID) (string, error) {
	var playerName string
	sqlString := "CALL SP_PICK_WINNER (?)"
	rows, err := query(sqlString, responseId)
	if err != nil {
		return playerName, err
	}

	for rows.Next() {
		if err := rows.Scan(&playerName); err != nil {
			log.Println(err)
			return playerName, errors.New("failed to scan row in query results")
		}
	}

	return playerName, nil
}

func PickRandomWinner(lobbyId uuid.UUID) (string, error) {
	var playerName string
	sqlString := "CALL SP_PICK_RANDOM_WINNER (?)"
	rows, err := query(sqlString, lobbyId)
	if err != nil {
		return playerName, err
	}

	for rows.Next() {
		if err := rows.Scan(&playerName); err != nil {
			log.Println(err)
			return playerName, errors.New("failed to scan row in query results")
		}
	}

	return playerName, nil
}

func DiscardHand(playerId uuid.UUID) error {
	sqlString := "CALL SP_DISCARD_HAND (?)"
	return execute(sqlString, playerId)
}

func SkipPrompt(lobbyId uuid.UUID) error {
	sqlString := "CALL SP_SKIP_PROMPT (?)"
	return execute(sqlString, lobbyId)
}

func SetJudgeResponseCount(lobbyId uuid.UUID, responseCount int) error {
	sqlString := "CALL SP_SET_RESPONSE_COUNT (?, ?)"
	return execute(sqlString, lobbyId, responseCount)
}
