package static

import "embed"

//go:embed *
var StaticFiles embed.FS

// SQLFiles is the ordered list of SQL files to execute for framework database
// setup. Order matters: settings -> tables -> functions -> procedures ->
// events -> triggers. The host game runs these before its own schema files
// (game extension-table foreign keys depend on the framework tables).
var SQLFiles = []string{
	// database
	"sql/settings.sql",

	// tables
	"sql/tables/USER.sql",
	"sql/tables/LOBBY.sql",
	"sql/tables/PLAYER.sql",
	"sql/tables/USER_ACCESS_LOBBY.sql",
	"sql/tables/LOGIN_ATTEMPT.sql",
	"sql/tables/AUDIT_USER.sql",
	"sql/tables/DECK.sql",
	"sql/tables/USER_ACCESS_DECK.sql",
	"sql/tables/AUDIT_DECK.sql",

	// functions
	"sql/functions/FN_GET_LOGIN_ATTEMPT_IS_ALLOWED.sql",
	"sql/functions/FN_GET_PLAYER_LOBBY_ID.sql",
	"sql/functions/FN_USER_HAS_LOBBY_ACCESS.sql",
	"sql/functions/FN_USER_HAS_DECK_ACCESS.sql",

	// procedures
	"sql/procedures/SP_SET_PLAYER_ACTIVE.sql",
	"sql/procedures/SP_SET_PLAYER_INACTIVE.sql",
	"sql/procedures/SP_GET_READABLE_DECKS.sql",

	// events
	"sql/events/EVT_CLEAN_AUDIT_USER.sql",
	"sql/events/EVT_CLEAN_LOGIN_ATTEMPTS.sql",

	// triggers
	"sql/triggers/TR_AUDIT_USER_DELETE.sql",
	"sql/triggers/TR_AUDIT_USER_UPDATE.sql",
	"sql/triggers/TR_PLAYER_BEFORE_INSERT.sql",
	"sql/triggers/TR_REVOKE_ACCESS_AF_UP_LOBBY.sql",
	"sql/triggers/TR_SET_CHANGED_ON_DATE_BF_UP_USER.sql",
	"sql/triggers/TR_AUDIT_DECK_DELETE.sql",
	"sql/triggers/TR_AUDIT_DECK_UPDATE.sql",
	"sql/triggers/TR_REVOKE_ACCESS_AF_UP_DECK.sql",
	"sql/triggers/TR_SET_CHANGED_ON_DATE_BF_UP_DECK.sql",
}
