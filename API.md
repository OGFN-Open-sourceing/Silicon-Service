# Silicon API v1

What a backend implements so Silicon-Service can manage it. Any backend can implement it, in any
language. Silicon itself only stores its admins, its config and an audit log; every player, account
and server action is a call to these routes.

- Base path: `<backend url>/silicon/v1`
- Auth: every request carries `Authorization: Bearer <apiKey>`, the key from `silicon.json`
  (`backend.apiKey`). Answer `401` when it doesn't match. Compare it in constant time.
- Bodies are JSON. Errors are `{"error": "message"}` with a 4xx/5xx status.
- A route you don't implement answers `404` or `501`. The panel then shows "This backend doesn't
  support that yet" instead of failing.
- Times are ISO 8601 strings or epoch milliseconds.
- Calls made for a signed-in admin also carry `X-Silicon-User` and `X-Silicon-Role`. They are for
  display only (Vindertech names announcements with them). Silicon has already checked the role.

## Status

`GET /status`

```json
{ "name": "Vindertech V2", "version": "2.3.0", "uptimeSeconds": 5321,
  "online": 12, "inQueue": 2, "inMatch": 8, "servers": 1 }
```

Silicon polls this every `pollSeconds` for the overview and the activity chart.

## Players (online now)

`GET /players` returns `{ "players": [ { "accountId", "displayName", "state": "lobby|queue|match", "playlist", "queuePos", "queuedSeconds" } ], "queueTotal" }`

`POST /players/{accountId}/kick` with `{ "reason" }` disconnects the player (XMPP, matchmaking, game).

## Accounts

`GET /accounts?q=<name, email or id>` returns `{ "accounts": [ { "accountId", "displayName", "email", "created", "lastLogin", "banned", "banReason", "founder", "host", "online" } ] }`.
With no `q`, the most recently active accounts.

`GET /accounts/{accountId}` returns one account with its profile summary: the fields above plus
`discordLinked`, `state` (`lobby|queue|match|offline`), `vbucks`, `level`, `xp`, `bookLevel` and
`battlePass`.

`POST /accounts/{accountId}/ban` with `{ "reason" }` bans the account and kicks it if it's online.

`POST /accounts/{accountId}/unban` with `{}`.

`POST /accounts/{accountId}/vbucks` with `{ "amount": 1000 }` adds V-Bucks. A negative amount removes them. Owner only.

`POST /accounts/{accountId}/founder` with `{ "founder": true|false }`. Owner only.

## Reports

`GET /reports` returns `{ "reports": [ { "id", "createdAt", "reporter", "reporterName", "reported", "reportedName", "reason", "details", "status": "open|resolved" } ] }`

`POST /reports/{id}/resolve` with `{ "action": "dismiss|ban|warn" }`.

## Game servers

`GET /servers` returns `{ "servers": [ { "id", "address", "playlist", "state", "players", "maxPlayers", "lastHeartbeat" } ] }`

`POST /servers/{id}/end` ends that match. Owner only.

## Playlists

`GET /playlists` returns `{ "playlists": [ { "name", "displayName", "enabled", "image", "autoImage", "category", "priority", "isDefault", "visibleWhenDisabled", "displayAsNew", "limitedTime", "description", "subname", "violator", "border" } ] }`.
`autoImage` is the tile the backend uses when `image` is empty.

`GET /playlists/library` returns `{ "library": [ { "name", "label", "group", "builds", "players", "squad", "image" } ] }`,
the modes an admin can add.

`PUT /playlists` with `{ "playlists": [ ...rows as above, without autoImage ] }` replaces the whole
list. Names are letters, digits and underscores, each once. Keep a backup. Owner only.

`POST /playlists` with `{ "name", "enabled" }` turns one on or off. Owner only.

## Chat

`GET /chat` returns `{ "messages": [ { "at", "from", "to", "body" } ] }`, recent messages, newest last.

`POST /announce` with `{ "message" }` sends to everyone online.

## Files (content and config)

`GET /files` returns `{ "files": ["news", "catalog", ...], "details": [ { "name", "label", "json", "size" } ] }`,
the files the backend allows editing.

`GET /files/{name}` returns `{ "name", "content" }`.

`PUT /files/{name}` with `{ "content" }`. Validate it, and keep a backup. Owner only.

## Restart

`POST /restart` with `{ "mode": "graceful|now" }`. Owner only.

## Roles

Silicon enforces two roles before a call reaches the backend:

- **owner**: everything.
- **moderator**: everything except `PUT /files/*`, `POST /restart`, `POST|PUT /playlists`,
  `POST /matchmaking*`, `POST /accounts/*/vbucks` and `POST /accounts/*/founder`.

Every non-GET call is written to Silicon's audit log with the admin, their IP, the body and the answer's status.
