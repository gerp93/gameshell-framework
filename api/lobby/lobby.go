package apiLobby

import (
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/gerp93/gameshell-framework/api"
	"github.com/gerp93/gameshell-framework/database"
	"github.com/gerp93/gameshell-framework/websocket"
	"github.com/google/uuid"
)

// maxTurnTimerSeconds caps the per-turn timer. Anything longer is
// indistinguishable from no timer at all in practice.
const maxTurnTimerSeconds = 300

// SetTurnTimer sets the lobby's per-turn timer, in seconds. Zero disables it.
// The countdown itself is client-side (see /gs/js/timer.js); this only stores
// the setting and tells the lobby it changed.
func SetTurnTimer(w http.ResponseWriter, r *http.Request) {
	lobbyIdString := r.PathValue("lobbyId")
	lobbyId, err := uuid.Parse(lobbyIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Failed to get lobby id from path."))
		return
	}

	userId := api.GetUserId(r)
	hasAccess, err := database.UserHasLobbyAccess(userId, lobbyId)
	if err != nil || !hasAccess {
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

	var turnTimerSeconds int
	for key, val := range r.Form {
		if key == "turnTimerSeconds" {
			turnTimerSeconds, err = strconv.Atoi(val[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Failed to parse turn timer."))
				return
			}
		}
	}

	if turnTimerSeconds < 0 {
		turnTimerSeconds = 0
	}
	if turnTimerSeconds > maxTurnTimerSeconds {
		turnTimerSeconds = maxTurnTimerSeconds
	}

	err = database.SetLobbyTurnTimerSeconds(lobbyId, turnTimerSeconds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	player, _ := database.GetLobbyUserPlayer(lobbyId, userId)
	name := html.EscapeString(player.Name)
	if turnTimerSeconds > 0 {
		websocket.LobbyBroadcast(lobbyId, fmt.Sprintf("<green>%s</>: Turn timer set to %d seconds", name, turnTimerSeconds))
	} else {
		websocket.LobbyBroadcast(lobbyId, fmt.Sprintf("<green>%s</>: Turn timer turned off", name))
	}
	websocket.LobbyBroadcast(lobbyId, fmt.Sprintf("turnTimer:%d", turnTimerSeconds))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Turn timer updated."))
}
