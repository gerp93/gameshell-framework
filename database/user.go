package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/gerp93/gameshell-framework/auth"
	"github.com/google/uuid"
)

type User struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	Name         string
	PasswordHash string
	ColorTheme   sql.NullString
	IsApproved   bool
	IsAdmin      bool
}

func SearchUsers(name string, page int) ([]User, error) {
	name = "%" + name + "%"

	if page < 1 {
		page = 1
	}

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			PASSWORD_HASH,
			COLOR_THEME,
			IS_APPROVED,
			IS_ADMIN
		FROM USER
		WHERE NAME LIKE ?
		ORDER BY CHANGED_ON_DATE DESC,
			NAME ASC
		LIMIT 10 OFFSET ?
	`
	rows, err := query(sqlString, name, (page-1)*10)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
			&user.IsApproved,
			&user.IsAdmin,
		); err != nil {
			log.Println(err)
			return nil, errors.New("failed to scan row in query results")
		}
		result = append(result, user)
	}
	return result, nil
}

func CountUsers(name string) (int, error) {
	name = "%" + name + "%"

	sqlString := `
		SELECT
			COUNT(*)
		FROM USER
		WHERE NAME LIKE ?
	`
	rows, err := query(sqlString, name)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Println(err)
			return 0, errors.New("failed to scan row in query results")
		}
	}

	return count, nil
}

func GetUser(userId uuid.UUID) (User, error) {
	var user User

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			PASSWORD_HASH,
			COLOR_THEME,
			IS_APPROVED,
			IS_ADMIN
		FROM USER
		WHERE ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return user, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&user.Id,
			&user.CreatedOnDate,
			&user.ChangedOnDate,
			&user.Name,
			&user.PasswordHash,
			&user.ColorTheme,
			&user.IsApproved,
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
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&passwordHash); err != nil {
			log.Println(err)
			return passwordHash, errors.New("failed to scan row in query results")
		}
	}

	return passwordHash, nil
}

func GetUserIsApproved(userId uuid.UUID) (bool, error) {
	var isApproved bool

	sqlString := `
		SELECT
			IS_APPROVED
		FROM USER
		WHERE ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return isApproved, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&isApproved); err != nil {
			log.Println(err)
			return isApproved, errors.New("failed to scan row in query results")
		}
	}

	return isApproved, nil
}

func GetUserIsAdmin(userId uuid.UUID) (bool, error) {
	var isAdmin bool

	sqlString := `
		SELECT
			IS_ADMIN
		FROM USER
		WHERE ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return isAdmin, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&isAdmin); err != nil {
			log.Println(err)
			return isAdmin, errors.New("failed to scan row in query results")
		}
	}

	return isAdmin, nil
}

func AddUserLoginAttempt(ipAddress string, userName string) error {
	sqlString := `
		INSERT INTO LOGIN_ATTEMPT(IP_ADDRESS, USER_NAME)
		VALUES (?, ?)
	`
	return execute(sqlString, ipAddress, userName)
}

func AllowUserLoginAttempt(ipAddress string, userName string) (bool, error) {
	sqlString := "SELECT FN_GET_LOGIN_ATTEMPT_IS_ALLOWED (?, ?)"
	rows, err := query(sqlString, ipAddress, userName)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	allowLogin := false
	for rows.Next() {
		if err := rows.Scan(&allowLogin); err != nil {
			log.Println(err)
			return false, errors.New("failed to scan row in query results")
		}
	}

	return allowLogin, nil
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
	defer rows.Close()

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
	defer rows.Close()

	return rows.Next()
}

func CreateUser(name string, password string, isApproved bool) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return errors.New("failed to hash password")
	}

	sqlString := `
		INSERT INTO USER (NAME, PASSWORD_HASH, IS_APPROVED)
		VALUES (?, ?, ?)
	`
	return execute(sqlString, name, passwordHash, isApproved)
}

func ApproveUser(id uuid.UUID) error {
	sqlString := `
		UPDATE USER
		SET IS_APPROVED = 1
		WHERE ID = ?
	`
	return execute(sqlString, id)
}

func SetUserName(id uuid.UUID, name string) error {
	sqlString := `
		UPDATE USER
		SET NAME = ?
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
		SET PASSWORD_HASH = ?
		WHERE ID = ?
	`
	err = execute(sqlString, passwordHash, id)
	if err != nil {
		return err
	}

	return nil
}

func SetUserColorTheme(id uuid.UUID, colorTheme string) error {
	sqlString := `
		UPDATE USER
		SET COLOR_THEME = ?
		WHERE ID = ?
	`
	var err error
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

// UserWinCelebration is the personalization a player gets to show off when
// they win — an optional GIF and an optional message shown beneath it. The
// GIF bytes are deliberately not part of this struct (or of User): GetUser
// runs on every page request, and the blob lives in its own table so the
// AUDIT_USER triggers never copy it.
type UserWinCelebration struct {
	UserId  uuid.UUID
	HasGif  bool
	Message sql.NullString
}

// GetUserWinCelebration returns a user's celebration metadata without loading
// the GIF bytes. A user who has never set one comes back zeroed, not an error.
func GetUserWinCelebration(userId uuid.UUID) (UserWinCelebration, error) {
	celebration := UserWinCelebration{UserId: userId}

	sqlString := `
		SELECT
			COALESCE(LENGTH(GIF_DATA), 0) > 0,
			MESSAGE
		FROM USER_WIN_CELEBRATION
		WHERE USER_ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return celebration, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&celebration.HasGif, &celebration.Message); err != nil {
			log.Println(err)
			return celebration, errors.New("failed to scan row in query results")
		}
	}

	return celebration, nil
}

// GetUserWinGif returns the raw GIF bytes and their mime type. Empty bytes
// mean the user has no GIF set.
func GetUserWinGif(userId uuid.UUID) ([]byte, string, error) {
	var data []byte
	var mime sql.NullString

	sqlString := `
		SELECT
			GIF_DATA,
			GIF_MIME
		FROM USER_WIN_CELEBRATION
		WHERE USER_ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&data, &mime); err != nil {
			log.Println(err)
			return nil, "", errors.New("failed to scan row in query results")
		}
	}

	return data, mime.String, nil
}

func SetUserWinGif(id uuid.UUID, data []byte, mime string) error {
	sqlString := `
		INSERT INTO USER_WIN_CELEBRATION(
			USER_ID,
			GIF_DATA,
			GIF_MIME
		)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			GIF_DATA = VALUES(GIF_DATA),
			GIF_MIME = VALUES(GIF_MIME),
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(6)
	`
	return execute(sqlString, id, data, mime)
}

func ClearUserWinGif(id uuid.UUID) error {
	sqlString := `
		UPDATE USER_WIN_CELEBRATION
		SET GIF_DATA = NULL,
			GIF_MIME = NULL,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(6)
		WHERE USER_ID = ?
	`
	return execute(sqlString, id)
}

func SetUserWinMessage(id uuid.UUID, message string) error {
	sqlString := `
		INSERT INTO USER_WIN_CELEBRATION(
			USER_ID,
			MESSAGE
		)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
			MESSAGE = VALUES(MESSAGE),
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(6)
	`
	if message == "" {
		return execute(sqlString, id, nil)
	}
	return execute(sqlString, id, message)
}

// UserLoseCelebration is the opposite of UserWinCelebration — an optional
// GIF and message a player can set to be shown to everyone in the lobby
// when they lose instead of win. See UserWinCelebration for why the GIF
// bytes are kept out of this struct.
type UserLoseCelebration struct {
	UserId  uuid.UUID
	HasGif  bool
	Message sql.NullString
}

// GetUserLoseCelebration returns a user's lose-celebration metadata without
// loading the GIF bytes. A user who has never set one comes back zeroed,
// not an error.
func GetUserLoseCelebration(userId uuid.UUID) (UserLoseCelebration, error) {
	celebration := UserLoseCelebration{UserId: userId}

	sqlString := `
		SELECT
			COALESCE(LENGTH(GIF_DATA), 0) > 0,
			MESSAGE
		FROM USER_LOSE_CELEBRATION
		WHERE USER_ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return celebration, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&celebration.HasGif, &celebration.Message); err != nil {
			log.Println(err)
			return celebration, errors.New("failed to scan row in query results")
		}
	}

	return celebration, nil
}

// GetUserLoseGif returns the raw GIF bytes and their mime type. Empty bytes
// mean the user has no lose GIF set.
func GetUserLoseGif(userId uuid.UUID) ([]byte, string, error) {
	var data []byte
	var mime sql.NullString

	sqlString := `
		SELECT
			GIF_DATA,
			GIF_MIME
		FROM USER_LOSE_CELEBRATION
		WHERE USER_ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&data, &mime); err != nil {
			log.Println(err)
			return nil, "", errors.New("failed to scan row in query results")
		}
	}

	return data, mime.String, nil
}

func SetUserLoseGif(id uuid.UUID, data []byte, mime string) error {
	sqlString := `
		INSERT INTO USER_LOSE_CELEBRATION(
			USER_ID,
			GIF_DATA,
			GIF_MIME
		)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			GIF_DATA = VALUES(GIF_DATA),
			GIF_MIME = VALUES(GIF_MIME),
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(6)
	`
	return execute(sqlString, id, data, mime)
}

func ClearUserLoseGif(id uuid.UUID) error {
	sqlString := `
		UPDATE USER_LOSE_CELEBRATION
		SET GIF_DATA = NULL,
			GIF_MIME = NULL,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(6)
		WHERE USER_ID = ?
	`
	return execute(sqlString, id)
}

func SetUserLoseMessage(id uuid.UUID, message string) error {
	sqlString := `
		INSERT INTO USER_LOSE_CELEBRATION(
			USER_ID,
			MESSAGE
		)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
			MESSAGE = VALUES(MESSAGE),
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(6)
	`
	if message == "" {
		return execute(sqlString, id, nil)
	}
	return execute(sqlString, id, message)
}

// UserWinVideo is the YouTube clip a player can set to force the lobby to
// watch when they win the game overall. Separate from UserWinCelebration —
// that GIF/message fires on every correct placement; this only fires on
// game-over.
type UserWinVideo struct {
	UserId             uuid.UUID
	HasVideo           bool
	YouTubeVideoId     sql.NullString
	StartOffsetSeconds int
}

// GetUserWinVideo returns a user's game-win video metadata. A user who has
// never set one comes back zeroed, not an error.
func GetUserWinVideo(userId uuid.UUID) (UserWinVideo, error) {
	video := UserWinVideo{UserId: userId}

	sqlString := `
		SELECT
			YOUTUBE_VIDEO_ID,
			START_OFFSET_SECONDS
		FROM USER_WIN_VIDEO
		WHERE USER_ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return video, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&video.YouTubeVideoId, &video.StartOffsetSeconds); err != nil {
			log.Println(err)
			return video, errors.New("failed to scan row in query results")
		}
		video.HasVideo = video.YouTubeVideoId.Valid && video.YouTubeVideoId.String != ""
	}

	return video, nil
}

func SetUserWinVideo(id uuid.UUID, youtubeVideoId string, startOffsetSeconds int) error {
	sqlString := `
		INSERT INTO USER_WIN_VIDEO(
			USER_ID,
			YOUTUBE_VIDEO_ID,
			START_OFFSET_SECONDS
		)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			YOUTUBE_VIDEO_ID = VALUES(YOUTUBE_VIDEO_ID),
			START_OFFSET_SECONDS = VALUES(START_OFFSET_SECONDS),
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(6)
	`
	return execute(sqlString, id, youtubeVideoId, startOffsetSeconds)
}

func ClearUserWinVideo(id uuid.UUID) error {
	sqlString := `
		UPDATE USER_WIN_VIDEO
		SET YOUTUBE_VIDEO_ID = NULL,
			START_OFFSET_SECONDS = 0,
			CHANGED_ON_DATE = CURRENT_TIMESTAMP(6)
		WHERE USER_ID = ?
	`
	return execute(sqlString, id)
}

func SetUserIsAdmin(id uuid.UUID, isAdmin bool) error {
	sqlString := `
		UPDATE USER
		SET IS_ADMIN = ?
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
		DELETE
		FROM USER
		WHERE ID = ?
	`
	err := execute(sqlString, id)
	if err != nil {
		return err
	}

	return nil
}
