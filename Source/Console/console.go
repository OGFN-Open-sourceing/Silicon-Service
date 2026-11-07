// Package Console draws Silicon's terminal screen: what it is serving, whether the backend
// answers, who is online, and a numbered menu of admin actions.
package Console

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/OGFN-Open-sourceing/Silicon-Service/Source/API"
	"github.com/OGFN-Open-sourceing/Silicon-Service/Source/Backend"
)

type Player struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	State       string `json:"state"`
	Playlist    string `json:"playlist"`
}

type Server struct {
	ID         string `json:"id"`
	Address    string `json:"address"`
	Playlist   string `json:"playlist"`
	State      string `json:"state"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"maxPlayers"`
}

type Console struct {
	Version string
	URL     string
	Config  *Backend.Config
	Client  *Backend.Client
	Monitor *Backend.Heartbeat
	Audit   *API.AuditLog
	Auth    *API.Auth
	Logs    *LogBuffer

	mu        sync.Mutex
	drawMu    sync.Mutex
	restore   func()
	showKey   bool
	notice    string
	noticeBad bool
	noticeEnd time.Time
	players   []Player
	servers   []Server
	flow      *flow
	answers   []string
	input     string
	confirm   bool
	busy      string
	lastW     int
	lastH     int
}

// Start takes over the terminal. It returns false when stdin/stdout is not an interactive
// console, in which case the caller keeps plain log output.
func (c *Console) Start() bool {
	restore, ok := enterRawMode()
	if !ok {
		return false
	}
	c.restore = restore
	os.Stdout.WriteString("\x1b[?1049h\x1b[?25l\x1b[2J")
	go c.readKeys()
	go c.pollLists()
	go func() {
		for {
			c.draw()
			time.Sleep(time.Second)
		}
	}()
	return true
}

func (c *Console) Stop() {
	c.drawMu.Lock()
	defer c.drawMu.Unlock()
	if c.restore != nil {
		os.Stdout.WriteString(reset + "\x1b[?25h\x1b[?1049l")
		c.restore()
		c.restore = nil
	}
}

// Notice shows msg above the menu for d.
func (c *Console) Notice(msg string, d time.Duration) { c.setNotice(msg, false, d) }

func (c *Console) setNotice(msg string, bad bool, d time.Duration) {
	c.mu.Lock()
	c.notice, c.noticeBad, c.noticeEnd = msg, bad, time.Now().Add(d)
	c.mu.Unlock()
	c.draw()
}

func (c *Console) say(msg string)  { c.setNotice(msg, false, 20*time.Second) }
func (c *Console) fail(msg string) { c.setNotice(msg, true, 30*time.Second) }

func (c *Console) pollLists() {
	for {
		if sample, ok := c.Monitor.Latest(); ok && sample.Up {
			c.refreshLists()
		}
		time.Sleep(3 * time.Second)
	}
}

func (c *Console) refreshLists() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var p struct {
		Players []Player `json:"players"`
	}
	var s struct {
		Servers []Server `json:"servers"`
	}
	errP := c.Client.GetJSON(ctx, "/players", &p)
	errS := c.Client.GetJSON(ctx, "/servers", &s)
	c.mu.Lock()
	if errP == nil {
		c.players = p.Players
	}
	if errS == nil {
		c.servers = s.Servers
	}
	c.mu.Unlock()
}

func (c *Console) readKeys() {
	buf := make([]byte, 64)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}
		c.handle(buf[:n])
		c.draw()
	}
}

func (c *Console) handle(data []byte) {
	if len(data) > 1 && data[0] == 0x1b {
		return
	}
	c.mu.Lock()
	inFlow := c.flow != nil
	c.mu.Unlock()
	if inFlow {
		c.handleFlow(data)
		return
	}
	for _, key := range data {
		switch {
		case key == 'q' || key == 'Q' || key == 3:
			c.Stop()
			os.Exit(0)
		case key >= '0' && key <= '9':
			c.choose(int(key - '0'))
		}
	}
}

func (c *Console) handleFlow(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		data = data[size:]
		switch {
		case r == 3 || r == 0x1b:
			c.flow, c.answers, c.input, c.confirm = nil, nil, "", false
			return
		case c.confirm:
			if r == 'y' || r == 'Y' {
				c.launch()
			} else if r == 'n' || r == 'N' {
				c.flow, c.answers, c.confirm = nil, nil, false
			}
			return
		case r == '\r' || r == '\n':
			c.submit()
			if c.flow == nil || c.confirm {
				return
			}
		case r == 8 || r == 127:
			if c.input != "" {
				_, last := utf8.DecodeLastRuneInString(c.input)
				c.input = c.input[:len(c.input)-last]
			}
		case r >= 32 && r != utf8.RuneError && utf8.RuneCountInString(c.input) < 200:
			c.input += string(r)
		}
	}
}

func (c *Console) submit() {
	step := c.flow.steps[len(c.answers)]
	value := trimSpace(c.input)
	if value == "" && !step.optional {
		return
	}
	c.answers = append(c.answers, value)
	c.input = ""
	if len(c.answers) < len(c.flow.steps) {
		return
	}
	if c.flow.confirm != nil {
		c.confirm = true
		return
	}
	c.launch()
}

func (c *Console) launch() {
	f, answers := c.flow, c.answers
	c.flow, c.answers, c.confirm = nil, nil, false
	c.busy = f.title + "…"
	go func() {
		msg, err := f.run(c, answers)
		c.mu.Lock()
		c.busy = ""
		c.mu.Unlock()
		if err != nil {
			c.fail(f.title + " failed: " + err.Error())
			return
		}
		c.say(msg)
		go c.refreshLists()
	}()
}

func (c *Console) choose(n int) {
	for _, a := range menu {
		if a.key != n {
			continue
		}
		if a.flow == nil {
			a.now(c)
			return
		}
		c.mu.Lock()
		if c.busy == "" {
			c.flow, c.answers, c.input, c.confirm = a.flow, nil, "", false
			if len(a.flow.steps) == 0 {
				if a.flow.confirm != nil {
					c.confirm = true
				} else {
					c.launch()
				}
			}
		}
		c.mu.Unlock()
		return
	}
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
