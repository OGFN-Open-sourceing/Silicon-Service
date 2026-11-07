// Package API is the dashboard's own HTTP API: login, admins, the audit log, and the proxy to the
// managed backend's Silicon API.
package API

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/OGFN-Open-sourceing/Silicon-Service/Source/Backend"
)

const sessionCookie = "silicon_session"

type session struct {
	User    string
	Role    string
	Expires time.Time
}

// Auth keeps logins in memory: restarting Silicon logs everyone out, which is fine for a panel.
type Auth struct {
	mu         sync.Mutex
	cfg        *Backend.Config
	configPath string
	sessions   map[string]session
	failures   map[string][]time.Time
}

func NewAuth(cfg *Backend.Config, configPath string) *Auth {
	return &Auth{cfg: cfg, configPath: configPath, sessions: map[string]session{}, failures: map[string][]time.Time{}}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// tooManyFailures allows 8 failed logins per IP per 15 minutes.
func (a *Auth) tooManyFailures(ip string) bool {
	cutoff := time.Now().Add(-15 * time.Minute)
	recent := a.failures[ip][:0]
	for _, t := range a.failures[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	a.failures[ip] = recent
	return len(recent) >= 8
}

func (a *Auth) Login(w http.ResponseWriter, r *http.Request, user, password string) (session, bool, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	ip := clientIP(r)
	if a.tooManyFailures(ip) {
		return session{}, false, true
	}
	for _, admin := range a.cfg.Admins {
		if strings.EqualFold(admin.User, user) && Backend.CheckPassword(admin.PasswordHash, password) {
			token := Backend.RandomHex(32)
			s := session{User: admin.User, Role: roleOf(admin), Expires: time.Now().Add(time.Duration(a.cfg.SessionHours) * time.Hour)}
			a.sessions[token] = s
			http.SetCookie(w, &http.Cookie{
				Name: sessionCookie, Value: token, Path: "/", HttpOnly: true,
				SameSite: http.SameSiteStrictMode, Secure: a.cfg.HTTPS.Enabled, Expires: s.Expires,
			})
			return s, true, false
		}
	}
	a.failures[ip] = append(a.failures[ip], time.Now())
	return session{}, false, false
}

func roleOf(admin Backend.AdminUser) string {
	if admin.Role == "moderator" {
		return "moderator"
	}
	return "owner"
}

func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		a.mu.Lock()
		delete(a.sessions, c.Value)
		a.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1})
}

func (a *Auth) Session(r *http.Request) (session, bool) {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return session{}, false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	s, ok := a.sessions[c.Value]
	if !ok || time.Now().After(s.Expires) {
		delete(a.sessions, c.Value)
		return session{}, false
	}
	return s, true
}

// Admins returns the admin list without hashes.
func (a *Auth) Admins() []map[string]string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := []map[string]string{}
	for _, admin := range a.cfg.Admins {
		out = append(out, map[string]string{"user": admin.User, "role": roleOf(admin)})
	}
	return out
}

// SetAdmin adds or updates an admin (password may be empty to keep the old one) and saves the config.
func (a *Auth) SetAdmin(user, password, role string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if role != "moderator" {
		role = "owner"
	}
	for i, admin := range a.cfg.Admins {
		if strings.EqualFold(admin.User, user) {
			a.cfg.Admins[i].Role = role
			if password != "" {
				a.cfg.Admins[i].PasswordHash = Backend.HashPassword(password)
			}
			return Backend.SaveConfig(a.configPath, a.cfg)
		}
	}
	a.cfg.Admins = append(a.cfg.Admins, Backend.AdminUser{User: user, PasswordHash: Backend.HashPassword(password), Role: role})
	return Backend.SaveConfig(a.configPath, a.cfg)
}

// RemoveAdmin refuses to remove the last owner, and ends the removed admin's sessions.
func (a *Auth) RemoveAdmin(user string) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	owners, index := 0, -1
	for i, admin := range a.cfg.Admins {
		if roleOf(admin) == "owner" {
			owners++
		}
		if strings.EqualFold(admin.User, user) {
			index = i
		}
	}
	if index < 0 || (roleOf(a.cfg.Admins[index]) == "owner" && owners <= 1) {
		return false, nil
	}
	a.cfg.Admins = append(a.cfg.Admins[:index], a.cfg.Admins[index+1:]...)
	for token, s := range a.sessions {
		if strings.EqualFold(s.User, user) {
			delete(a.sessions, token)
		}
	}
	return true, Backend.SaveConfig(a.configPath, a.cfg)
}
