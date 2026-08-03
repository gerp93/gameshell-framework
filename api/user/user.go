package apiUser

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"unicode/utf8"

	"github.com/gerp93/gameshell-framework/api"
	"github.com/gerp93/gameshell-framework/auth"
	"github.com/gerp93/gameshell-framework/database"
	"github.com/google/uuid"
)

// maxWinGifBytes is the per-game-configurable size limit for a win image.
// The image is fetched on every win popup, so games are expected to keep it
// bounded — set via SetMaxWinGifBytes at startup, same as SetBrandName/
// SetCookiePrefix/etc. Defaults conservatively; a game that wants a larger
// (or smaller) limit calls the setter once during init.
var maxWinGifBytes = 60 * 1024

const maxWinMessageRunes = 140

// SetMaxWinGifBytes overrides the default per-image size limit enforced by
// SetWinGif. Call it once at startup, before serving requests.
func SetMaxWinGifBytes(maxBytes int) {
	maxWinGifBytes = maxBytes
}

// MaxWinGifBytes returns the currently configured per-image size limit
// (shared by the win and lose GIF uploads), so pages can display it next to
// the upload control instead of only enforcing it silently.
func MaxWinGifBytes() int {
	return maxWinGifBytes
}

// maxWinGifUploadBytes bounds the multipart read: payload plus framing
// overhead (boundary markers, field headers), not a size a real image is
// meant to approach. SetWinGif rejects on Content-Length before ever
// reading the body when a client honestly declares an oversized upload (the
// normal case — a browser always knows a picked file's size), so this cap
// is just a defense-in-depth backstop for a request that didn't declare its
// size. Keep it just above maxWinGifBytes for framing slack, not materially
// larger — see SetWinGif's Content-Length check for why this doesn't
// reopen the mid-upload connection-reset that a read cap near
// maxWinGifBytes used to cause.
func maxWinGifUploadBytes() int64 {
	return int64(maxWinGifBytes) + 4096
}

// winImageTypes are the accepted upload formats, identified by magic bytes
// rather than by extension or the browser-supplied content type.
var winImageTypes = []struct {
	magic []byte
	mime  string
}{
	{[]byte("GIF87a"), "image/gif"},
	{[]byte("GIF89a"), "image/gif"},
	{[]byte("\x89PNG\r\n\x1a\n"), "image/png"},
}

// winImageMime returns the mime type for the uploaded bytes, or "" when the
// data is not an accepted image format.
func winImageMime(data []byte) string {
	for _, t := range winImageTypes {
		if bytes.HasPrefix(data, t.magic) {
			return t.mime
		}
	}
	return ""
}

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

// SetWinGif stores the image shown when this user wins. Despite the name
// (kept for API stability) it accepts GIF or PNG, detected by magic bytes —
// see winImageMime. This is the only multipart handler in the framework;
// everything else takes a plain form.
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

	// Reject an honestly-oversized upload by its declared Content-Length
	// before parsing any of the body. A browser always knows a picked
	// file's size upfront, so this is the normal path for "you picked too
	// big a file". The client sends the whole multipart body in one write
	// (no Expect: 100-continue), so it's typically still mid-upload when
	// this handler responds; writing the error without also draining the
	// unread body races the client's in-flight send against the server
	// closing the connection, and on the losing side of that race the
	// response never reaches the browser at all — surfaced as
	// net::ERR_CONNECTION_RESET, not this response (reproduced live: a file
	// even modestly over the limit reset the connection instead of showing
	// the message). Draining first lets the server finish reading the
	// declared body before it replies, so the connection closes cleanly.
	sizeLimitMessage := fmt.Sprintf("Image must be %d KB or smaller.", maxWinGifBytes/1024)

	if r.ContentLength > maxWinGifUploadBytes() {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(sizeLimitMessage))
		return
	}

	// Backstop for a request that didn't declare Content-Length honestly
	// (e.g. chunked transfer encoding) — not the normal path, see above.
	r.Body = http.MaxBytesReader(w, r.Body, maxWinGifUploadBytes())
	err = r.ParseMultipartForm(maxWinGifUploadBytes())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(sizeLimitMessage))
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

	data, err := io.ReadAll(io.LimitReader(file, int64(maxWinGifBytes)+1))
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
		_, _ = w.Write([]byte(sizeLimitMessage))
		return
	}

	mime := winImageMime(data)
	if mime == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("That file is not a GIF or PNG."))
		return
	}

	err = database.SetUserWinGif(userId, data, mime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// Deliberately no HX-Refresh: this form lives inside a <details> on the
	// account page, and a full reload snaps it shut so the new image looks
	// like it never saved. The caller updates the preview in place instead.
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Image saved."))
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

	// No HX-Refresh, for the same reason as SetWinGif.
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Image removed."))
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
		mime = "image/gif"
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
		_, _ = w.Write([]byte("Win message must be 140 characters or fewer."))
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

// SetLoseGif stores the image shown when this user loses — the counterpart
// to SetWinGif. Shares the same size limit and format checks; see SetWinGif
// for why the Content-Length check comes before anything is read.
func SetLoseGif(w http.ResponseWriter, r *http.Request) {
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

	sizeLimitMessage := fmt.Sprintf("Image must be %d KB or smaller.", maxWinGifBytes/1024)

	// Drain before responding — see SetWinGif's early-reject branch for why:
	// without this, a client still mid-upload races the server closing the
	// connection and net::ERR_CONNECTION_RESET wins instead of this response.
	if r.ContentLength > maxWinGifUploadBytes() {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(sizeLimitMessage))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxWinGifUploadBytes())
	err = r.ParseMultipartForm(maxWinGifUploadBytes())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(sizeLimitMessage))
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	file, _, err := r.FormFile("loseGif")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("No GIF found."))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, int64(maxWinGifBytes)+1))
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
		_, _ = w.Write([]byte(sizeLimitMessage))
		return
	}

	mime := winImageMime(data)
	if mime == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("That file is not a GIF or PNG."))
		return
	}

	err = database.SetUserLoseGif(userId, data, mime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// Deliberately no HX-Refresh, same reasoning as SetWinGif.
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Image saved."))
}

// ClearLoseGif removes a user's lose GIF, leaving their lose message alone.
func ClearLoseGif(w http.ResponseWriter, r *http.Request) {
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

	err = database.ClearUserLoseGif(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Image removed."))
}

// GetLoseGif serves a user's lose GIF. Unlike the setter this is not limited
// to the current user — every player in a lobby has to render the loser's.
func GetLoseGif(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get user id from path."))
		return
	}

	data, mime, err := database.GetUserLoseGif(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	if len(data) == 0 {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("No lose gif found."))
		return
	}

	if mime == "" {
		mime = "image/gif"
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// SetLoseMessage stores the message shown beneath the lose GIF.
func SetLoseMessage(w http.ResponseWriter, r *http.Request) {
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

	var loseMessage string
	for key, val := range r.Form {
		if key == "loseMessage" {
			loseMessage = val[0]
		}
	}

	if utf8.RuneCountInString(loseMessage) > maxWinMessageRunes {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Lose message must be 140 characters or fewer."))
		return
	}

	err = database.SetUserLoseMessage(userId, loseMessage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Lose message saved."))
}

func isCurrentUser(r *http.Request, checkId uuid.UUID) bool {
	userId := api.GetUserId(r)
	return userId == checkId
}
