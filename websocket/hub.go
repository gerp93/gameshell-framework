package websocket

import (
	"log"

	"github.com/gerp93/gameshell-framework"
	"github.com/gerp93/gameshell-framework/database"
	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Lobby ID
	lobbyId uuid.UUID

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub(lobbyId uuid.UUID) *Hub {
	return &Hub{
		lobbyId:    lobbyId,
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
			if len(h.clients) == 0 {
				if g := gameshell.Registered(); g != nil {
					_ = g.OnRoomEmpty(h.lobbyId)
				}
				_ = database.DeleteLobby(h.lobbyId)
				delete(lobbyHubs, h.lobbyId)
				return
			}
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.clients[client] = true
	// Belt-and-braces reactivation, not a duplicate of what the page load
	// already did: closes a race where an old connection's disconnect is
	// detected late (see unregisterClient) and lands in between the page
	// load reactivating this player and this websocket actually
	// registering, leaving them stuck inactive despite this brand new
	// connection. AddUserToLobby is idempotent (a no-op update, no hook
	// fired) when the player is already active.
	if _, err := database.AddUserToLobby(h.lobbyId, client.user.Id); err != nil {
		log.Println(err)
	}
	h.broadcastMessage([]byte("<blue>Player Joined</>: <green>" + client.user.Name + "</>"))
	h.broadcastMessage([]byte("refresh"))
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
		// A fast reconnect (page reload, a flaky connection recovering) can
		// register the user's new client before this one's disconnect is
		// even detected -- readPump only notices a dead connection once its
		// blocking Read finally errors out, which can lag well behind a
		// brand new connection completing its handshake. If that happened,
		// h.clients still holds another live client for the same user right
		// now, and marking them inactive here would stomp the reactivation
		// their reconnect already did (see gsDatabase.AddUserToLobby),
		// leaving them permanently stuck inactive despite being connected.
		if !h.userStillConnected(client.user.Id) {
			_ = database.SetPlayerInactive(h.lobbyId, client.user.Id)
		}
	}
	h.broadcastMessage([]byte("<red>Player Left</>: <green>" + client.user.Name + "</>"))
	h.broadcastMessage([]byte("refresh"))
}

// userStillConnected reports whether any other currently registered client
// in this hub belongs to the given user.
func (h *Hub) userStillConnected(userId uuid.UUID) bool {
	for c := range h.clients {
		if c.user.Id == userId {
			return true
		}
	}
	return false
}

func (h *Hub) broadcastMessage(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func LobbyBroadcast(lobbyId uuid.UUID, message string) {
	if hub, ok := lobbyHubs[lobbyId]; ok {
		hub.broadcastMessage([]byte(message))
	}
}

func PlayerBroadcast(playerId uuid.UUID, message string) {
	player, err := database.GetPlayer(playerId)
	if err != nil {
		return
	}

	if hub, ok := lobbyHubs[player.LobbyId]; ok {
		for client := range hub.clients {
			if client.user.Id == player.UserId {
				client.send <- []byte(message)
			}
		}
	}
}
