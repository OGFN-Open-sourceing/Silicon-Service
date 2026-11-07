// Package Backend holds Silicon's config and everything that talks to the backend it manages.
package Backend

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
)

type HTTPSConfig struct {
	Enabled bool `json:"enabled"`
	// Leave both empty to have Silicon make (and reuse) a self-signed certificate.
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	// Extra names/IPs for the self-signed certificate besides localhost and 127.0.0.1.
	Hosts []string `json:"hosts"`
}

type BackendConfig struct {
	Name string `json:"name"`
	// Base URL of the backend, e.g. http://127.0.0.1:3551 (Silicon calls <url>/silicon/v1/...).
	URL string `json:"url"`
	// Shared secret sent as "Authorization: Bearer <key>". The backend must be given the same key.
	APIKey         string `json:"apiKey"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
}

type AdminUser struct {
	User string `json:"user"`
	// pbkdf2-sha256$<iterations>$<salt hex>$<hash hex>
	PasswordHash string `json:"passwordHash"`
	// "owner" can do everything; "moderator" can't edit files, grant V-Bucks or restart.
	Role string `json:"role"`
}

type Config struct {
	Listen       string        `json:"listen"`
	HTTPS        HTTPSConfig   `json:"https"`
	Backend      BackendConfig `json:"backend"`
	Admins       []AdminUser   `json:"admins"`
	SessionHours int           `json:"sessionHours"`
	PollSeconds  int           `json:"pollSeconds"`
	AuditLog     string        `json:"auditLog"`
	// Only known on the run that created the file.
	FirstRunPassword string `json:"-"`
}

func RandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func defaults() *Config {
	return &Config{
		Listen: "127.0.0.1:8420",
		Backend: BackendConfig{
			Name:           "Vindertech V2",
			URL:            "http://127.0.0.1:3551",
			APIKey:         RandomHex(24),
			TimeoutSeconds: 8,
		},
		SessionHours: 12,
		PollSeconds:  10,
		AuditLog:     "silicon-audit.jsonl",
	}
}

// LoadOrCreateConfig reads path, or writes a fresh one (with a generated owner login) if it doesn't exist.
func LoadOrCreateConfig(path string) (*Config, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg := defaults()
		password := RandomHex(8)
		cfg.Admins = []AdminUser{{User: "admin", PasswordHash: HashPassword(password), Role: "owner"}}
		cfg.FirstRunPassword = password
		return cfg, true, SaveConfig(path, cfg)
	}
	if err != nil {
		return nil, false, err
	}
	cfg := defaults()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, false, err
	}
	if len(cfg.Admins) == 0 {
		return nil, false, errors.New("no admins in config: delete the file to regenerate it, or add one")
	}
	if cfg.SessionHours <= 0 {
		cfg.SessionHours = 12
	}
	if cfg.PollSeconds <= 0 {
		cfg.PollSeconds = 10
	}
	return cfg, false, nil
}

func SaveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
