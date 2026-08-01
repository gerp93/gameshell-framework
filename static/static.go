package static

import "embed"

//go:embed *
var StaticFiles embed.FS

// SQLFiles is the ordered list of core SQL files to execute for framework
// database setup — everything every game needs regardless of which
// bootstrap.Features it enables. Order matters: settings -> tables ->
// functions -> procedures -> events -> triggers. The host game runs these
// before its own schema files (game extension-table foreign keys depend on
// the framework tables). Deck-related objects live in DeckSQLFiles instead,
// applied separately (via bootstrap.ApplyFeatureSchema) only when a game
// enables Features.Decks — a deckless game shouldn't get DECK/
// USER_ACCESS_DECK/AUDIT_DECK tables just because it linked this module.
var SQLFiles = []string{
	// database
	"sql/settings.sql",

	// tables
	"sql/tables/USER.sql",
	"sql/tables/USER_WIN_CELEBRATION.sql",
	"sql/tables/USER_LOSE_CELEBRATION.sql",
	"sql/tables/LOBBY.sql",
	"sql/tables/LOBBY_SETTINGS.sql",
	"sql/tables/PLAYER.sql",
	"sql/tables/USER_ACCESS_LOBBY.sql",
	"sql/tables/LOGIN_ATTEMPT.sql",
	"sql/tables/AUDIT_USER.sql",

	// functions
	"sql/functions/FN_GET_LOGIN_ATTEMPT_IS_ALLOWED.sql",
	"sql/functions/FN_GET_PLAYER_LOBBY_ID.sql",
	"sql/functions/FN_USER_HAS_LOBBY_ACCESS.sql",

	// procedures
	"sql/procedures/SP_SET_PLAYER_ACTIVE.sql",
	"sql/procedures/SP_SET_PLAYER_INACTIVE.sql",

	// events
	"sql/events/EVT_CLEAN_AUDIT_USER.sql",
	"sql/events/EVT_CLEAN_LOGIN_ATTEMPTS.sql",

	// triggers
	"sql/triggers/TR_AUDIT_USER_DELETE.sql",
	"sql/triggers/TR_AUDIT_USER_UPDATE.sql",
	"sql/triggers/TR_PLAYER_BEFORE_INSERT.sql",
	"sql/triggers/TR_REVOKE_ACCESS_AF_UP_LOBBY.sql",
	"sql/triggers/TR_SET_CHANGED_ON_DATE_BF_UP_USER.sql",
}

// DeckSQLFiles is the ordered list of deck-related SQL files, applied only
// when a game enables bootstrap.Features.Decks. USER_ACCESS_DECK FKs to
// USER (in SQLFiles, applied first), but nothing in SQLFiles depends on
// anything here — safe to skip entirely for a deckless game.
var DeckSQLFiles = []string{
	// tables
	"sql/tables/DECK.sql",
	"sql/tables/USER_ACCESS_DECK.sql",
	"sql/tables/AUDIT_DECK.sql",

	// functions
	"sql/functions/FN_USER_HAS_DECK_ACCESS.sql",

	// procedures
	"sql/procedures/SP_GET_READABLE_DECKS.sql",

	// triggers
	"sql/triggers/TR_AUDIT_DECK_DELETE.sql",
	"sql/triggers/TR_AUDIT_DECK_UPDATE.sql",
	"sql/triggers/TR_REVOKE_ACCESS_AF_UP_DECK.sql",
	"sql/triggers/TR_SET_CHANGED_ON_DATE_BF_UP_DECK.sql",

	// migrations (idempotent ALTERs for pre-existing databases; run last so
	// objects that referenced a dropped column are already replaced above)
	"sql/migrations/MIG_DECK_DROP_IS_HIDDEN.sql",
	"sql/migrations/MIG_AUDIT_DECK_DROP_IS_HIDDEN.sql",
}
