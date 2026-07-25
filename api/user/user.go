package apiUser

import (
	"bytes"
	"io"
	"net/http"
	"unicode/utf8"

	"github.com/gerp93/gameshell-framework/api"
	"github.com/gerp93/gameshell-framework/auth"
	"github.com/gerp93/gameshell-framework/database"
	"github.com/google/uuid"
)

// Limits on the per-user win celebration. The GIF rides inside the HTTP
// response of every win popup, so it is kept small on purpose.
const (
	maxWinGifBytes      = 60 * 1024
	maxWinMessageRunes  = 1000
	winGifMimeType      = "image/gif"
	winGifMultipartSize = maxWinGifBytes + 4096 // payload plus multipart framing
)

func Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var name string
	var password string
	var passwordConfirm string
	for key, val := range r.Form {
		switch key {
		case "name":
			name = val[0]
		case "password":
			password = val[0]
		case "passwordConfirm":
			passwordConfirm = val[0]
		}
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No name found."))
		return
	}

	if password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No password found."))
		return
	}

	if password != passwordConfirm {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Passwords do not match."))
		return
	}

	if database.UserNameExists(name) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("User name already exists."))
		return
	}

	err = database.CreateUser(name, password, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Your request has been submitted. Please wait for an administrator to approve this account."))
}

func CreateAdmin(w http.ResponseWriter, r *http.Request) {
	if !api.UserIsAdmin(r) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var name string
	var password string
	var passwordConfirm string
	for key, val := range r.Form {
		switch key {
		case "name":
			name = val[0]
		case "password":
			password = val[0]
		case "passwordConfirm":
			passwordConfirm = val[0]
		}
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No name found."))
		return
	}

	if password == "" {
		password = "password"
	} else if password != passwordConfirm {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Passwords do not match."))
		return
	}

	if database.UserNameExists(name) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("User name already exists."))
		return
	}

	err = database.CreateUser(name, password, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusCreated)
}

func Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var name string
	var password string
	for key, val := range r.Form {
		switch key {
		case "name":
			name = val[0]
		case "password":
			password = val[0]
		}
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No name found."))
		return
	}

	if password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No password found."))
		return
	}

	allowLogin, err := database.AllowUserLoginAttempt(r.RemoteAddr, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if !allowLogin {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Too many login attempts, please wait an hour to try again."))
		return
	}

	err = database.AddUserLoginAttempt(r.RemoteAddr, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if !database.UserNameExists(name) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("User name does not exist."))
		return
	}

	userId, err := database.GetUserIdByName(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	isApproved, err := database.GetUserIsApproved(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if !isApproved {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("User account is not yet approved by an administrator."))
		return
	}

	passwordHash, err := database.GetUserPasswordHash(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if !auth.PasswordMatchesHash(password, passwordHash) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Provided password is not valid."))
		return
	}

	auth.SetUserId(w, userId)

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func Logout(w http.ResponseWriter, _ *http.Request) {
	auth.RemoveUserId(w)
	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func SetName(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	if !isCurrentUser(r, userId) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var name string
	for key, val := range r.Form {
		if key == "name" {
			name = val[0]
		}
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No name found."))
		return
	}

	if database.UserNameExists(name) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("User name already exists."))
		return
	}

	err = database.SetUserName(userId, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func SetPassword(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	if !isCurrentUser(r, userId) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var currentPassword string
	var newPassword string
	var newPasswordConfirm string
	for key, val := range r.Form {
		switch key {
		case "currentPassword":
			currentPassword = val[0]
		case "newPassword":
			newPassword = val[0]
		case "newPasswordConfirm":
			newPasswordConfirm = val[0]
		}
	}

	if currentPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No current password found."))
		return
	}

	passwordHash, err := database.GetUserPasswordHash(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if !auth.PasswordMatchesHash(currentPassword, passwordHash) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Provided current password is not valid."))
		return
	}

	if newPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No new password found."))
		return
	}

	if newPassword != newPasswordConfirm {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("New passwords do not match."))
		return
	}

	err = database.SetUserPassword(userId, newPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	if !api.UserIsAdmin(r) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = database.SetUserPassword(userId, "password")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("<span class='bi bi-check-square'></span>"))
}

func Approve(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	if !api.UserIsAdmin(r) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = database.ApproveUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("<span class='bi bi-check-square'></span>"))
}

func SetColorTheme(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	if !isCurrentUser(r, userId) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var colorTheme string
	for key, val := range r.Form {
		if key == "colorTheme" {
			colorTheme = val[0]
		}
	}

	if colorTheme == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No color theme found."))
		return
	}

	err = database.SetUserColorTheme(userId, colorTheme)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func SetIsAdmin(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	if !api.UserIsAdmin(r) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var isAdmin bool
	for key, val := range r.Form {
		if key == "isAdmin" {
			isAdmin = val[0] == "1"
		}
	}

	err = database.SetUserIsAdmin(userId, isAdmin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	isCurrentUser := isCurrentUser(r, userId)
	if !isCurrentUser && !api.UserIsAdmin(r) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = database.DeleteUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if isCurrentUser {
		auth.RemoveUserId(w)
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

// SetWinGif stores the GIF shown when this user wins. This is the only
// multipart handler in the framework; everything else takes a plain form.
func SetWinGif(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	if !isCurrentUser(r, userId) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	// Cap the request before anything is buffered, so an oversized upload is
	// refused rather than read into memory.
	r.Body = http.MaxBytesReader(w, r.Body, winGifMultipartSize)
	err = r.ParseMultipartForm(winGifMultipartSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("GIF must be 60 KB or smaller."))
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	file, _, err := r.FormFile("winGif")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No GIF found."))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, maxWinGifBytes+1))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to read the uploaded file."))
		return
	}

	if len(data) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("The uploaded file is empty."))
		return
	}
	if len(data) > maxWinGifBytes {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("GIF must be 60 KB or smaller."))
		return
	}

	// Check the magic bytes rather than trusting the extension or the
	// browser-supplied content type.
	if !bytes.HasPrefix(data, []byte("GIF87a")) && !bytes.HasPrefix(data, []byte("GIF89a")) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("That file is not a GIF."))
		return
	}

	err = database.SetUserWinGif(userId, data, winGifMimeType)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

// ClearWinGif removes a user's win GIF, leaving their win message alone.
func ClearWinGif(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	if !isCurrentUser(r, userId) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = database.ClearUserWinGif(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

// GetWinGif serves a user's win GIF. Unlike the setters this is not limited
// to the current user — every player in a lobby has to render the winner's.
func GetWinGif(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	data, mime, err := database.GetUserWinGif(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	if len(data) == 0 {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("No win gif found."))
		return
	}

	if mime == "" {
		mime = winGifMimeType
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// SetWinMessage stores the message shown beneath the win GIF.
func SetWinMessage(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	if !isCurrentUser(r, userId) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("User does not have access."))
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to parse form."))
		return
	}

	var winMessage string
	for key, val := range r.Form {
		if key == "winMessage" {
			winMessage = val[0]
		}
	}

	if utf8.RuneCountInString(winMessage) > maxWinMessageRunes {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Win message must be 1000 characters or fewer."))
		return
	}

	err = database.SetUserWinMessage(userId, winMessage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Win message saved."))
}

func isCurrentUser(r *http.Request, checkId uuid.UUID) bool {
	userId := api.GetUserId(r)
	return userId == checkId
}
