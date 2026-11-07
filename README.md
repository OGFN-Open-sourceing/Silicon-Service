# Silicon-Service

An admin and moderation panel for OGFN backends, as one small Go exe. It was made for the
Vindertech V2 backend, but any backend that implements the [Silicon API](API.md) can use it.

Your backend can run on a VPS while you run `silicon.exe` on any PC (or on the VPS itself). Silicon
talks to the backend over HTTP(S) with a shared API key, and serves the dashboard on your machine.

## Running it

```
silicon.exe                 # uses ./silicon.json, creating it on the first run
silicon.exe -config other.json
silicon.exe -plain          # plain log lines instead of the console screen (services, log files)
```

In a terminal, Silicon shows a live console screen: the dashboard address, whether the backend
answers (and what to fix when it doesn't), a players-online sparkline, the online players and game
servers as numbered lists, and recent admin actions mixed with its log.

Actions are on the number keys. The ones that need input ask for it on the bottom line (Enter
continues, Esc cancels), and the risky ones ask y/n first:

| Key | Action | Key | Action |
| --- | --- | --- | --- |
| 1 | Announce | 6 | End a match (list number or id) |
| 2 | Kick player (list number or name) | 7 | Restart backend (graceful or now) |
| 3 | Ban account | 8 | Open the dashboard in your browser |
| 4 | Unban account | 9 | Show or hide the API key |
| 5 | Give V-Bucks | 0 | New random owner password |

**q** quits. Console actions show up in the audit log as user `console`.

On the first run Silicon writes `silicon.json` and shows:

- a login for the `admin` owner account (change the password under **Admins**), and
- a generated **API key**. Give the backend the same key. For Vindertech, that's
  `SILICON_API_KEY=<key>` in its `.env`.

Then open http://localhost:8420 (or press **8**).

## silicon.json

```json
{
  "listen": "127.0.0.1:8420",
  "https": { "enabled": false, "certFile": "", "keyFile": "", "hosts": [] },
  "backend": { "name": "Vindertech V2", "url": "http://127.0.0.1:3551", "apiKey": "…", "timeoutSeconds": 8 },
  "admins": [ { "user": "admin", "passwordHash": "pbkdf2-sha256$…", "role": "owner" } ],
  "sessionHours": 12,
  "pollSeconds": 10,
  "auditLog": "silicon-audit.jsonl"
}
```

- **listen**: keep it on `127.0.0.1` to use the panel only from this PC. Use `0.0.0.0:8420` to
  reach it from other machines, and turn HTTPS on if you do.
- **https**: set `enabled` to `true`. With no cert/key files given, Silicon makes a self-signed
  certificate (`silicon-cert.pem` and `silicon-key.pem`) for localhost, 127.0.0.1 and any `hosts`
  you list, and reuses it.
- **backend.url**: where the backend answers `/silicon/v1/...`. If the backend is on a VPS and
  Silicon isn't, use HTTPS or a tunnel, since the API key travels in a header.
Silicon's own messages tell you what's wrong when it can't use the backend: a wrong API key, the
backend's Silicon API switched off (no `SILICON_API_KEY`), or a wrong `backend.url`.

- **admins**: owners can do everything. Moderators can view everything and ban, kick, resolve
  reports and announce, but can't edit files, grant V-Bucks, change playlists or restart.

## What's in the panel

A dark, launcher-style dashboard. The overview shows status, online/queue/in-match counts, a
players-online chart (1 h, 6 h, 24 h), quick actions and recent admin activity. Other pages:
online players by state with queue position and wait time, accounts (recent list, search, and a
detail drawer with V-Bucks, level, battle pass, ban/unban, kick, V-Bucks and founder), reports,
game servers (end match), a playlist editor (tiles, names, categories, order, flags, and adding
modes from the backend's library), chat and announcements (sent under your admin name), backend
restarts (owners), content files, the audit log of every admin action, and admin users.

## Structure

```
main.go                 flags, startup
API.md                  the contract a backend implements
Source/
 Backend/
  config.go             silicon.json
  password.go           pbkdf2 password hashes (stdlib only)
  cert.go               self-signed HTTPS
  backend.go            client for the backend's Silicon API
  heartbeat_sys.go      status polling and activity history
 API/
  auth.go               logins, sessions, rate limit, admin users
  audit.go              audit log (JSON lines)
  manager.go            the dashboard's /api and the proxy to the backend
 Console/
  console.go            the terminal screen and its shortcuts
  logbuffer.go          recent log lines for the screen
  term_*.go             raw keyboard input per platform
 Web/
  site.go               embeds and serves the dashboard
  static/               the built dashboard (generated from Web/, don't edit by hand)
Web/                    dashboard source: Astro pages, TypeScript, CSS
 src/pages/             one page per section
 src/layouts/           shell (sidebar, top bar, dialog, toasts)
 src/lib/               API client, dialogs, table helpers, icons
 src/styles/global.css  the design tokens and components
```

## Building

Go 1.24 or newer, with no dependencies outside the standard library:

```
go build -o silicon.exe .
```

The built dashboard is committed in `Source/Web/static`, so `go build` alone is enough. To change
the dashboard (Node 20 or newer):

```
cd Web
npm install
npm run dev      # live reload on http://localhost:4321, talks to a Silicon running on 8420
npm run build    # writes Source/Web/static; rebuild the exe afterwards
```

## License

BSD-3-Clause.
