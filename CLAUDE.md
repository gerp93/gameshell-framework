# CLAUDE.md — gameshell-framework

Guidance for working in this repository. This module is the reusable platform
extracted from [card-judge](https://github.com/gerp93/card-judge); the two
repos share one style so they read as a single author's codebase. **This file
is a style guide first, a contract second.** Match the surrounding code; do
not introduce new styles, formatters, or abstractions.

## What this is

A Go library (`module github.com/gerp93/gameshell-framework`, go.mod at repo
root) providing the platform for multiplayer party games: **users and
accounts**, auth, lobbies/rooms, player presence, **generic deck
management**, **shared chat rendering**, **the color-theme system**,
websocket realtime, and the framework schema. Games live in their own repos
(card-judge, timeline-trivia), pin a version of this module, and plug in via
the `Game` interface.

Stack: **Go (stdlib `net/http`) + `gorilla/websocket` + MariaDB.** No web
framework, no ORM.

## The contract (must not break)

- **`Game` interface** (`gameshell.go`, root package `gameshell`): lifecycle
  hooks `OnRoomCreated`, `OnPlayerJoined`, `OnPlayerActive`,
  `OnPlayerInactive`, `OnRoomEmpty`. Games call `gameshell.Register` once at
  startup; the framework invokes hooks through `gameshell.Registered()`.
- **One-way dependency:** the framework must NEVER import a game. Game
  behavior enters only through the `Game` interface.
- **Base tables are game-free:** `LOBBY(ID, CREATED_ON_DATE, NAME, MESSAGE,
  PASSWORD_HASH)` and `PLAYER(ID, CREATED_ON_DATE, LOBBY_ID, USER_ID,
  JOIN_ORDER, IS_ACTIVE)`. Games extend via their own 1:1 FK tables — never
  add game columns or game triggers to framework tables.
- **Schema ordering:** the host game runs `static.SQLFiles` (via
  `database.RunFile`) BEFORE its own schema; game extension-table FKs depend
  on the framework tables. `SQLFiles` order is manual and matters
  (settings → tables → functions → procedures → events → triggers).
- **Parameterization points** a game sets at startup: `database.SetEnvPrefix`
  (`<PREFIX>_SQL_HOST/_DATABASE/_USER/_PASSWORD`), `auth.SetCookiePrefix`,
  `api.SetBrandName`, `api.SetPagePolicy`. Generalize by adding parameters to
  existing code, not by rewriting it.
- **Realtime model:** short control strings over the socket, never HTML; the
  game's client re-fetches HTML fragments in response. `LobbyBroadcast` /
  `PlayerBroadcast` (`websocket/hub.go`) are the public realtime API. Presence
  is `PLAYER.IS_ACTIVE`, flipped on ws connect/disconnect; the hub deletes the
  lobby (after `OnRoomEmpty`) when its last client disconnects.
- **Game DB access:** games use the exported `database.Query` /
  `database.Execute` and keep hand-written SQL.
- **Decks are framework-owned, cards are not:** `DECK(ID, CREATED_ON_DATE,
  CHANGED_ON_DATE, NAME UNIQUE, PASSWORD_HASH, IS_PUBLIC_READONLY)` +
  `USER_ACCESS_DECK` + `AUDIT_DECK` (`database/deck.go`, `api/deck/deck.go`)
  are the full deck lifecycle — creation, name/password/visibility changes,
  access grants, deletion. Every `DECK` row is a real, user-facing deck; games
  needing per-room internal cards model them in their own tables (e.g.
  card-judge's wild cards are `CARD` rows keyed by `LOBBY_ID`, not decks).
  Games own their own `CARD`-equivalent table, FK'd to `DECK.ID`, with their
  own schema/columns/CRUD — the framework never touches card rows directly.
- **`OnDeckDeleting(deckId uuid.UUID) error`** (part of the `Game` interface,
  `gameshell.go`): `database.DeleteDeck` calls this **before** deleting the
  `DECK` row. MariaDB's `ON DELETE CASCADE` from `DECK` to a game's `CARD`
  table does **not** fire that table's own triggers, so games use this hook
  to audit/clean up their cards themselves. Do not assume cascade alone
  keeps a game's audit trail correct.
- **Chat rendering is shared, not game-specific:** `static/js/chat.js`
  (`window.gsChat.append`/`.wireForm`) parses the `<red>`/`<green>`/`<blue>`/
  `</>` color tokens used in lobby broadcast messages, adds a timestamp, and
  trims history; `static/css/chat.css` styles those tokens via
  `--color-accent-*` (with fallbacks) so it adapts to each game's active
  theme. Games mount the framework's `static.StaticFiles` under `/gs/` and
  include `/gs/js/chat.js` + `/gs/css/chat.css` instead of writing their own
  chat renderer.
- **User accounts and the theme system are framework-owned:** `USER` (incl.
  `COLOR_THEME`), auth, and the full account lifecycle — create, admin-create,
  login/logout, name/password change, admin password-reset/approve/is-admin,
  delete, color-theme — live in `api/user/user.go` (package `apiUser`), mounted
  by games at the same 11 routes card-judge originally defined
  (`/api/user/...`). `static/css/colors.css` (22 themes, served at
  `/gs/css/colors.css` via the same `/gs/` mount as chat) and `api.ThemeGroups`
  (`api/theme.go` — the canonical `[]ThemeGroup{Name, Themes []Theme{Value,
  Label}}` list, grouped "Classic"/"Visual Assault"/"Tractor") are the single
  source of truth for available themes. Games render their own `account.html`
  and range over `api.ThemeGroups` to build the theme `<select>` instead of
  hardcoding `<option>` tags — the framework ships the data, never the HTML
  (same split as deck management).

## Style (same as card-judge — match exactly)

- `gofmt`/tabs; sparse comments (the `websocket/` files keep their
  gorilla-example doc comments).
- DB layer: raw SQL strings, backtick literals for multi-line; row-by-row
  `Scan` with `defer rows.Close()`; on scan error
  `log.Println(err); return ..., errors.New("failed to scan row in query results")`;
  structs mirror table columns; `CALL SP_...` wrappers. No ORM.
- SQL files: UPPERCASE keywords AND identifiers, one object per file, `SP_/FN_/
  TR_/EVT_/V_/AUDIT_/LOG_` prefixes, `VAR_` local variables,
  `CREATE TABLE IF NOT EXISTS` / `CREATE OR REPLACE`, formatted with
  `sqlfmt --newlines --upper --spaces 4 --comment-pre-space`.
- Handlers (in `api/`): `func Name(w http.ResponseWriter, r *http.Request)`,
  plain-text sentence responses via `w.WriteHeader(...)` +
  `_, _ = w.Write([]byte("Human sentence."))`.
- IDs are `uuid.UUID` (`uuid.NewUUID()` in Go, `UUID()` in SQL).

## Versioning / release

- Semver git tags `vMAJOR.MINOR.PATCH`; bump with
  `version_bump.sh {major|minor|patch}` (updates README version line).
- The Go API and the framework schema move together per tag. Games pin a tag
  in `go.mod`.
- **Upgrade caveat:** `SQLFiles` only creates/replaces; removing an object
  from the manifest does not drop it from existing databases — document any
  required manual `DROP` in the release notes.

## Build / verify

- `go build ./...` + `go vet ./...` (also run on tag by the release workflow).
- There is no test suite; verify changes by running a consuming game
  (card-judge or timeline-trivia) against a local MariaDB and playing through
  lobby join/leave, a full round, and a websocket disconnect (see card-judge's
  `docs/local-mariadb-setup.md`). If the change touches decks, verify both
  games since deck management is shared code.

## Known quirks (preserved from card-judge by design)

- The auth signing secret is process-random (`auth/cookie.go`): sessions do
  not survive restarts and cannot be shared across instances.
- The full framework schema re-runs on every consuming app's startup
  (idempotent by design).
