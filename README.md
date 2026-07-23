# Gameshell Framework

Version: 0.6.0

A reusable Go + HTMX + MariaDB platform for multiplayer party games. Gameshell
provides the bones every game needs — user accounts, authentication, game
lobbies, player presence, and websocket realtime — so each game repo only has
to implement its own rules, tables, and UI.

Extracted from [card-judge](https://github.com/gerp93/card-judge), which is the
first game built on it.

## What the framework owns

- **Users & auth**: accounts, bcrypt passwords, signed cookie sessions, login
  attempt rate limiting, admin/approval flags (`auth/`, `database/user.go`),
  and the full account-management HTTP layer — create, login/logout,
  name/password change, admin approve/reset/is-admin, delete, color-theme
  (`api/user/user.go`).
- **Theme system**: `static/css/colors.css` (22 switchable themes, served at
  `/gs/css/colors.css`) plus `api.ThemeGroups` (`api/theme.go`), the canonical
  grouped theme list games range over to build their own account-page picker.
- **Lobbies (rooms)**: create/delete, name/message/password, password-gated
  access grants (`database/lobby.go`, `database/access.go`).
- **Players (participants)**: per-lobby membership with join order and
  active/inactive presence (`database/player.go`).
- **Realtime**: one gorilla/websocket hub per lobby with `LobbyBroadcast` and
  `PlayerBroadcast` (`websocket/`). Presence flips on connect/disconnect, and
  the lobby is deleted when its last client disconnects.
- **Page middleware**: login/admin gating from a declarative `PagePolicy`,
  brand-aware `BasePageData` (`api/`).
- **Framework schema**: the base tables and SQL objects (`static/sql/`),
  applied in order via `static.SQLFiles` + `database.RunFile`.

## How a game plugs in

A game implements the `Game` interface (root package) and registers it at
startup. The framework invokes the hooks at room/player lifecycle moments; the
framework never imports a game.

```go
gameshell.Register(myGame{})            // lifecycle hooks
database.SetEnvVarPrefix("MY_GAME")        // MY_GAME_SQL_HOST, _DATABASE, _USER, _PASSWORD
auth.SetCookiePrefix("MY-GAME")         // MY-GAME-USER-TOKEN cookie
api.SetBrandName("My Game")             // top bar + default page title
api.SetPagePolicy(api.PagePolicy{...})  // which paths need login/admin
```

On startup the game runs the framework schema first, then its own:

```go
for _, sqlFile := range static.SQLFiles { // framework static
    database.RunFile(sqlFile)
}
// then the game's own embedded SQL via database.Execute(...)
```

Game state lives in **1:1 extension tables** keyed by `LOBBY.ID` / `PLAYER.ID`
foreign keys — games never add columns to the framework tables. Game queries
use `database.Query` / `database.Execute`.

## Versioning

Semver git tags (`vMAJOR.MINOR.PATCH`) via `version_bump.sh`. The Go API and
the framework schema move together per tag; games pin a version in `go.mod`.
The schema manifest only creates/replaces objects — removing an object from
`SQLFiles` does not drop it from an existing database.

### Migrations

`sql/migrations/` holds idempotent `ALTER`/`DROP` scripts for evolving a
database that was already provisioned by an older version of the schema —
e.g. `MIG_DECK_DROP_IS_HIDDEN.sql` drops the retired `DECK.IS_HIDDEN` column
(`DROP COLUMN IF EXISTS`, safe to run against a database that never had it).
They're registered in `static.SQLFiles` last, after every table/function/
procedure/event/trigger, so anything that referenced the old object has
already been replaced by the time the migration runs.

A migration is a temporary bridge, not a permanent fixture: once every real
deployment has run past the version that introduced it (i.e. no database in
the wild still predates the change), the migration script and its `SQLFiles`
entry get deleted — there's no longer anything left for it to migrate.

## License

AGPL-3.0 (inherited from card-judge, from which this framework was extracted).
