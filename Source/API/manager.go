package API

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/OGFN-Open-sourceing/Silicon-Service/Source/Backend"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// SecurityHeaders keeps the panel out of frames and off other origins.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self'")
		next.ServeHTTP(w, r)
	})
}

// sameOrigin rejects cross-site writes (the cookie is SameSite=Strict too; this covers old browsers).
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	return err == nil && strings.EqualFold(u.Host, r.Host)
}

// ownerOnly lists the backend calls a moderator may not make: content files, currency, founder
// status, restarts and match/playlist configuration. Everything else (players, bans, kicks, reports, chat) is
// moderation and open to both roles.
func ownerOnly(method, path string) bool {
	if method == http.MethodGet {
		return false
	}
	switch {
	case strings.HasPrefix(path, "/files"),
		strings.HasPrefix(path, "/restart"),
		strings.HasPrefix(path, "/playlists"),
		strings.HasPrefix(path, "/matchmaking"),
		strings.HasSuffix(path, "/vbucks"),
		strings.HasSuffix(path, "/founder"):
		return true
	}
	return false
}

func Register(mux *http.ServeMux, auth *Auth, client *Backend.Client, monitor *Backend.Heartbeat, audit *AuditLog) {
	requireSession := func(owner bool, next func(http.ResponseWriter, *http.Request, session)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			s, ok := auth.Session(r)
			if !ok {
				writeError(w, http.StatusUnauthorized, "not logged in")
				return
			}
			if r.Method != http.MethodGet && !sameOrigin(r) {
				writeError(w, http.StatusForbidden, "cross-origin request")
				return
			}
			if owner && s.Role != "owner" {
				writeError(w, http.StatusForbidden, "owners only")
				return
			}
			next(w, r, s)
		}
	}

	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		if !sameOrigin(r) {
			writeError(w, http.StatusForbidden, "cross-origin request")
			return
		}
		var body struct{ User, Password string }
		if json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body) != nil {
			writeError(w, http.StatusBadRequest, "bad request")
			return
		}
		s, ok, limited := auth.Login(w, r, body.User, body.Password)
		if limited {
			writeError(w, http.StatusTooManyRequests, "too many failed logins, try again later")
			return
		}
		if !ok {
			writeError(w, http.StatusUnauthorized, "wrong user or password")
			return
		}
		audit.Add(AuditEntry{At: time.Now(), User: s.User, IP: clientIP(r), Action: "login", Status: 200})
		writeJSON(w, 200, map[string]string{"user": s.User, "role": s.Role})
	})

	mux.HandleFunc("POST /api/logout", func(w http.ResponseWriter, r *http.Request) {
		auth.Logout(w, r)
		writeJSON(w, 200, map[string]bool{"ok": true})
	})

	mux.HandleFunc("GET /api/me", requireSession(false, func(w http.ResponseWriter, r *http.Request, s session) {
		writeJSON(w, 200, map[string]string{"user": s.User, "role": s.Role, "backend": client.Name()})
	}))

	mux.HandleFunc("POST /api/me/password", requireSession(false, func(w http.ResponseWriter, r *http.Request, s session) {
		var body struct{ Password string }
		if json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body) != nil || len(body.Password) < 8 {
			writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
			return
		}
		if err := auth.SetAdmin(s.User, body.Password, s.Role); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		audit.Add(AuditEntry{At: time.Now(), User: s.User, IP: clientIP(r), Action: "change own password", Status: 200})
		writeJSON(w, 200, map[string]bool{"ok": true})
	}))

	mux.HandleFunc("GET /api/overview", requireSession(false, func(w http.ResponseWriter, r *http.Request, s session) {
		latest, ok := monitor.Latest()
		writeJSON(w, 200, map[string]any{"backend": client.Name(), "polled": ok, "latest": latest})
	}))

	mux.HandleFunc("GET /api/activity", requireSession(false, func(w http.ResponseWriter, r *http.Request, s session) {
		hours, _ := strconv.Atoi(r.URL.Query().Get("hours"))
		if hours <= 0 || hours > 24 {
			hours = 6
		}
		writeJSON(w, 200, monitor.Since(time.Now().Add(-time.Duration(hours)*time.Hour), 360))
	}))

	mux.HandleFunc("GET /api/audit", requireSession(false, func(w http.ResponseWriter, r *http.Request, s session) {
		writeJSON(w, 200, audit.Recent(300))
	}))

	mux.HandleFunc("GET /api/admins", requireSession(true, func(w http.ResponseWriter, r *http.Request, s session) {
		writeJSON(w, 200, auth.Admins())
	}))

	mux.HandleFunc("POST /api/admins", requireSession(true, func(w http.ResponseWriter, r *http.Request, s session) {
		var body struct{ User, Password, Role string }
		if json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body) != nil || strings.TrimSpace(body.User) == "" {
			writeError(w, http.StatusBadRequest, "user is required")
			return
		}
		if body.Password != "" && len(body.Password) < 8 {
			writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
			return
		}
		if err := auth.SetAdmin(strings.TrimSpace(body.User), body.Password, body.Role); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		audit.Add(AuditEntry{At: time.Now(), User: s.User, IP: clientIP(r), Action: "set admin " + body.User + " (" + body.Role + ")", Status: 200})
		writeJSON(w, 200, auth.Admins())
	}))

	mux.HandleFunc("DELETE /api/admins/{user}", requireSession(true, func(w http.ResponseWriter, r *http.Request, s session) {
		user := r.PathValue("user")
		removed, err := auth.RemoveAdmin(user)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		if !removed {
			writeError(w, http.StatusBadRequest, "no such admin, or it is the last owner")
			return
		}
		audit.Add(AuditEntry{At: time.Now(), User: s.User, IP: clientIP(r), Action: "remove admin " + user, Status: 200})
		writeJSON(w, 200, auth.Admins())
	}))

	// Everything under /api/b/ goes to the backend's /silicon/v1/ as is. The backend decides what it
	// supports; Silicon adds login, roles and the audit trail.
	mux.HandleFunc("/api/b/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/b")
		requireSession(ownerOnly(r.Method, path), func(w http.ResponseWriter, r *http.Request, s session) {
			var body any
			var raw []byte
			if r.Method != http.MethodGet && r.Method != http.MethodDelete {
				raw, _ = io.ReadAll(io.LimitReader(r.Body, 8<<20))
				if len(raw) > 0 && json.Unmarshal(raw, &body) != nil {
					writeError(w, http.StatusBadRequest, "body must be JSON")
					return
				}
			}
			target := path
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			status, data, err := client.DoAs(r.Context(), Backend.Actor{User: s.User, Role: s.Role}, r.Method, target, body)
			if r.Method != http.MethodGet {
				logged := string(raw)
				if len(logged) > 2000 {
					logged = logged[:2000] + "…"
				}
				audit.Add(AuditEntry{At: time.Now(), User: s.User, IP: clientIP(r), Action: r.Method, Path: path, Body: logged, Status: status})
			}
			if err != nil {
				writeError(w, http.StatusBadGateway, "backend unreachable: "+err.Error())
				return
			}
			if status == http.StatusUnauthorized || (status == http.StatusNotFound && strings.Contains(string(data), "SILICON_API_KEY")) {
				writeError(w, http.StatusBadGateway, Backend.Explain(status, data))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			w.Write(data)
		})(w, r)
	})
}
