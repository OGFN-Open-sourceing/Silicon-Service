// Silicon-Service: a standalone admin and moderation panel for OGFN backends.
//
// Run it next to (or far from) a backend that implements the Silicon API (see API.md). On first
// start it writes silicon.json with a generated admin password and backend API key, then serves
// the dashboard over HTTP or HTTPS.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/OGFN-Open-sourceing/Silicon-Service/Source/API"
	"github.com/OGFN-Open-sourceing/Silicon-Service/Source/Backend"
	"github.com/OGFN-Open-sourceing/Silicon-Service/Source/Console"
	"github.com/OGFN-Open-sourceing/Silicon-Service/Source/Web"
)

const Version = "0.1.0"

func main() {
	configPath := flag.String("config", "silicon.json", "path to the config file (created on first run)")
	plain := flag.Bool("plain", false, "plain log output instead of the console screen (for services and log files)")
	flag.Parse()

	cfg, created, err := Backend.LoadOrCreateConfig(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	client := Backend.NewClient(cfg.Backend)
	monitor := Backend.NewHeartbeat(client, time.Duration(cfg.PollSeconds)*time.Second)
	go monitor.Run()

	auth := API.NewAuth(cfg, *configPath)
	audit := API.NewAuditLog(cfg.AuditLog)

	mux := http.NewServeMux()
	API.Register(mux, auth, client, monitor, audit)
	Web.Register(mux)

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           API.SecurityHeaders(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	scheme := "http"
	if cfg.HTTPS.Enabled {
		scheme = "https"
	}
	url := scheme + "://" + displayAddr(cfg.Listen)

	screen := &Console.Console{Version: Version, URL: url, Config: cfg, Client: client, Monitor: monitor, Audit: audit, Auth: auth, Logs: &Console.LogBuffer{}}
	if !*plain && screen.Start() {
		log.SetOutput(screen.Logs)
		log.SetFlags(log.Ltime)
	}
	fail := func(format string, args ...any) {
		screen.Stop()
		log.SetOutput(os.Stderr)
		log.Fatalf(format, args...)
	}

	if created {
		first := fmt.Sprintf("First run: log in as %q with password %q, and give the backend the API key (press 9).", cfg.Admins[0].User, cfg.FirstRunPassword)
		screen.Notice(first, 24*time.Hour)
		log.Printf("Created %s. First-run login: user %q, password %q (change it in the panel or the file).", *configPath, cfg.Admins[0].User, cfg.FirstRunPassword)
		log.Printf("Give your backend this API key (Vindertech: SILICON_API_KEY in .env): %s", cfg.Backend.APIKey)
	}

	log.Printf("Silicon-Service %s for %q (%s)", Version, cfg.Backend.Name, cfg.Backend.URL)
	log.Printf("Dashboard on %s", url)
	if cfg.HTTPS.Enabled {
		certFile, keyFile, err := Backend.EnsureCertificate(cfg.HTTPS)
		if err != nil {
			fail("https: %v", err)
		}
		fail("%v", srv.ListenAndServeTLS(certFile, keyFile))
	}
	fail("%v", srv.ListenAndServe())
}

func displayAddr(listen string) string {
	if len(listen) > 0 && listen[0] == ':' {
		return "localhost" + listen
	}
	return listen
}
