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
- **Parameterization points** a game sets at startup: `database.SetEnvVarPrefix`
  (`<PREFIX>_SQL_HOST/_DATABASE/_USER/_PASSWORD`), `auth.SetCookiePrefix`,
  `api.SetBrandName`, `api.SetPagePolicy`, `apiUser.SetMaxWinGifBytes`.
  Generalize by adding parameters to existing code, not by rewriting it.
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
- **Win celebrations are per-user personalization, so they're framework-owned:**
  `USER_WIN_CELEBRATION(USER_ID PK→USER, CHANGED_ON_DATE, GIF_DATA MEDIUMBLOB,
  GIF_MIME, MESSAGE VARCHAR(1000))` — an optional image (≤`apiUser.SetMaxWinGifBytes`,
  60 KB by default, GIF or PNG — columns/handler names keep the `WinGif` name
  for API stability, but `winImageMime` accepts either) and message (≤140 runes, enforced by
  `apiUser.SetWinMessage`; the column is wider only as headroom, not the
  active limit) a game shows when that player wins. It is deliberately a
  **side table, not `USER` columns**: the `AUDIT_USER` triggers enumerate
  every `USER` column and would copy the blob on every user update, and
  `GetUser` runs on every page request via `MiddlewareForPages`.
  `database.GetUserWinCelebration` returns metadata only (`HasGif` +
  message); the bytes come from `GetUserWinGif`/`apiUser.GetWinGif`.
  `apiUser.SetWinGif` is the framework's only multipart handler — it caps the
  body with `http.MaxBytesReader` and checks the `GIF87a`/`GIF89a`/PNG magic
  rather than trusting the extension. Its response deliberately carries no
  `HX-Refresh`: the account-page form lives inside a `<details>`, and a full
  reload collapses it right after the upload — the caller updates the preview
  in place with JS instead.
- **Lobby settings live in `LOBBY_SETTINGS`, not on `LOBBY`:**
  `LOBBY_SETTINGS(LOBBY_ID PK→LOBBY, TURN_TIMER_SECONDS)` holds
  framework-level, game-agnostic lobby configuration. Getter/setter
  (`database.GetLobbyTurnTimerSeconds`/`SetLobbyTurnTimerSeconds`) upsert and
  `COALESCE`, so no row has to exist and `CreateLobby`'s signature is
  unchanged. `apiLobby.SetTurnTimer` (`api/lobby/lobby.go`) is the handler,
  mounted by games at `PUT /api/lobby/{lobbyId}/turn-timer`. Add new
  framework-level lobby settings as columns here; game-specific ones still
  belong in the game's own 1:1 table.
- **Navigation is markup, not framework JS:** anything that navigates on click
  is a real `<a href>` — never `onclick="location.href=..."` and never a JS
  click-interceptor. The browser then gives ctrl/cmd-click, middle-click,
  right-click "Open in new tab", the hover URL preview and link semantics for
  free, none of which a script can fully reproduce. Wrap the visual element:
  `<a href="/decks"><button>Card Decks</button></a>`, or
  `<a href="/account" class="no-style"><div class="top-bar-menu-link">…</div></a>`.
- **Pages that are identical across every game are framework-owned HTML +
  handler, not just framework-owned data:** `api/pages` (package `apiPages`)
  ships `Login`, `Users`, `Decks`, `DeckAccess`, and `Account`, each backed by
  a template under `static/html/pages/body/` and `static/html/pages/base.html`.
  Games mount these directly — `http.Handle("GET /login", ...MiddlewareForPages(http.HandlerFunc(gsApiPages.Login)))`
  — the same zero-wrapper pattern already used for `gsApiDeck`'s CRUD
  handlers. `Account`'s optional win-celebration section is gated by
  `apiPages.SetAccountPageFeatures(apiPages.AccountPageFeatures{WinCelebration: bool})`
  (default off — the same safe-default `Set*` pattern as `SetBrandName`/
  `SetMaxWinGifBytes`; a game must opt in, and only after it has also mounted
  `apiUser`'s win-gif/win-message handlers, or the section renders pointing
  at routes that don't exist).
- **Pages that are *mostly* shared use a chrome-plus-slot split, not a full
  page:** `static/html/pages/body/deck-detail-chrome.html` renders everything
  about a deck's detail page that's identical between games (header, Export
  Deck, the Edit Deck dialog, the danger-zone delete) and leaves three named
  blocks — `card-header-actions`, `card-search-controls`, `card-management`
  — for the genuinely per-game part (each game's own card schema/dialogs).
  The game supplies its own small fragment file defining those three block
  names and composes it with the chrome via `apiPages.ParseGameFragment`.
  **Critical constraint:** every page body template in this framework and
  every consuming game defines the *same* Go template name, `{{define "body"}}`
  — this only works because exactly one body file is ever parsed per
  request. A composed parse must never include two files that both define
  `"body"`; `text/template` silently lets the second overwrite the first,
  with no compile-time signal. A game's slot-filling fragment must define
  distinctly-named blocks, never `"body"`.
- **The turn countdown is shared client-side plumbing:** `static/js/timer.js`
  (`window.gsTimer.start(el, seconds, onExpire)`/`.stop()`/`.reset()`) +
  `static/css/timer.css`, served at `/gs/js/timer.js` and `/gs/css/timer.css`.
  What happens at zero is game-specific and is passed in as `onExpire` — the
  framework runs no server-side timer goroutine and stores no deadline.
- **Assets found identical across both consuming games move here too, not
  just page templates:** `static/js/deck.js` (the CSV-export download
  handler for `deck-detail-chrome.html`'s Export Deck button — it reads the
  filename from the export target's `data-filename` attribute rather than a
  hardcoded per-game string, so it needed no game-specific branch to become
  shared) and `static/css/about.css`/`home.css` (styling for `about.html`/
  `home.html`, which stay game-owned since their *content* genuinely
  differs — only the CSS was byte-identical). Before assuming something is
  game-specific, diff it against the other consuming game first; several
  "obviously per-game" files turned out identical.
- **`bootstrap` package removes main()'s repeated startup sequence:**
  `bootstrap.ConnectWithRetry`, `.ApplySchema`, `.MountStaticAssets`,
  `.Serve` — every consuming game's `main()` had byte-identical DB-connect-
  retry, schema-application, static-mounting, and port/cert/log-file/listen
  logic, differing only in the env-var prefix and which `embed.FS` to read
  from. These functions `log.Fatalln` and exit the process on failure
  (matching what every call site already did) rather than returning an
  error — this package exists only to be called from `main()`, never from
  request-serving code, so that's an intentional, scoped exception to the
  rest of the framework's error-returning convention.

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
