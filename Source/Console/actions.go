package Console

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/OGFN-Open-sourceing/Silicon-Service/Source/API"
	"github.com/OGFN-Open-sourceing/Silicon-Service/Source/Backend"
)

type step struct {
	prompt   string
	optional bool
}

type flow struct {
	title   string
	steps   []step
	confirm func(answers []string) string
	run     func(c *Console, answers []string) (string, error)
}

type action struct {
	key   int
	label string
	flow  *flow
	now   func(c *Console)
}

var menu []action

func init() {
	menu = []action{
		{key: 1, label: "Announce", flow: &flow{
			title: "Announce",
			steps: []step{{prompt: "Message to everyone online"}},
			run: func(c *Console, a []string) (string, error) {
				if err := c.call(http.MethodPost, "/announce", map[string]string{"message": a[0]}); err != nil {
					return "", err
				}
				return "Announcement sent.", nil
			},
		}},
		{key: 2, label: "Kick player", flow: &flow{
			title:   "Kick player",
			steps:   []step{{prompt: "Player (list number or name)"}, {prompt: "Reason (optional)", optional: true}},
			confirm: func(a []string) string { return "Kick " + a[0] + withReason(a[1]) + "?" },
			run: func(c *Console, a []string) (string, error) {
				p, err := c.findPlayer(a[0])
				if err != nil {
					return "", err
				}
				if err := c.call(http.MethodPost, "/players/"+url.PathEscape(p.AccountID)+"/kick", map[string]string{"reason": a[1]}); err != nil {
					return "", err
				}
				return "Kicked " + p.DisplayName + ".", nil
			},
		}},
		{key: 3, label: "Ban account", flow: &flow{
			title:   "Ban account",
			steps:   []step{{prompt: "Account (list number, name or account id)"}, {prompt: "Reason (optional)", optional: true}},
			confirm: func(a []string) string { return "Ban " + a[0] + withReason(a[1]) + "?" },
			run: func(c *Console, a []string) (string, error) {
				id, name, err := c.findAccount(a[0])
				if err != nil {
					return "", err
				}
				if err := c.call(http.MethodPost, "/accounts/"+url.PathEscape(id)+"/ban", map[string]string{"reason": a[1]}); err != nil {
					return "", err
				}
				return "Banned " + name + ".", nil
			},
		}},
		{key: 4, label: "Unban account", flow: &flow{
			title: "Unban account",
			steps: []step{{prompt: "Account (name or account id)"}},
			run: func(c *Console, a []string) (string, error) {
				id, name, err := c.findAccount(a[0])
				if err != nil {
					return "", err
				}
				if err := c.call(http.MethodPost, "/accounts/"+url.PathEscape(id)+"/unban", map[string]string{}); err != nil {
					return "", err
				}
				return "Unbanned " + name + ".", nil
			},
		}},
		{key: 5, label: "Give V-Bucks", flow: &flow{
			title: "Give V-Bucks",
			steps: []step{{prompt: "Account (list number, name or account id)"}, {prompt: "Amount (negative removes)"}},
			run: func(c *Console, a []string) (string, error) {
				amount, err := strconv.Atoi(strings.ReplaceAll(a[1], ",", ""))
				if err != nil {
					return "", fmt.Errorf("%q is not a whole number", a[1])
				}
				id, name, err := c.findAccount(a[0])
				if err != nil {
					return "", err
				}
				if err := c.call(http.MethodPost, "/accounts/"+url.PathEscape(id)+"/vbucks", map[string]int{"amount": amount}); err != nil {
					return "", err
				}
				return fmt.Sprintf("Gave %d V-Bucks to %s.", amount, name), nil
			},
		}},
		{key: 6, label: "End a match", flow: &flow{
			title:   "End a match",
			steps:   []step{{prompt: "Server (list number or id)"}},
			confirm: func(a []string) string { return "End match on server " + a[0] + "? Its players go back to the lobby." },
			run: func(c *Console, a []string) (string, error) {
				s, err := c.findServer(a[0])
				if err != nil {
					return "", err
				}
				if err := c.call(http.MethodPost, "/servers/"+url.PathEscape(s.ID)+"/end", map[string]string{}); err != nil {
					return "", err
				}
				return "Ended the match on " + s.Address + ".", nil
			},
		}},
		{key: 7, label: "Restart backend", flow: &flow{
			title: "Restart backend",
			steps: []step{{prompt: "1 = graceful (75 s warning), 2 = now"}},
			confirm: func(a []string) string {
				if a[0] == "2" {
					return "Restart the backend right now? Everyone online is disconnected."
				}
				return "Restart the backend after a 75 second warning?"
			},
			run: func(c *Console, a []string) (string, error) {
				mode := "graceful"
				if a[0] == "2" {
					mode = "now"
				}
				if err := c.call(http.MethodPost, "/restart", map[string]string{"mode": mode}); err != nil {
					return "", err
				}
				return "Backend restart requested (" + mode + ").", nil
			},
		}},
		{key: 8, label: "Open dashboard", now: func(c *Console) {
			if err := openBrowser(c.URL); err != nil {
				c.fail("Could not open a browser: " + err.Error())
				return
			}
			c.say("Opened " + c.URL + " in your browser.")
		}},
		{key: 9, label: "Show API key", now: func(c *Console) {
			c.mu.Lock()
			c.showKey = !c.showKey
			c.mu.Unlock()
		}},
		{key: 0, label: "Reset owner password", flow: &flow{
			title:   "Reset owner password",
			steps:   nil,
			confirm: func([]string) string { return "Give the owner account a new random password?" },
			run:     (*Console).resetPassword,
		}},
	}
}

func (c *Console) choiceLabel(a action) string {
	if a.key == 9 {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.showKey {
			return "Hide API key"
		}
	}
	return a.label
}

func withReason(r string) string {
	if r == "" {
		return ""
	}
	return " (" + r + ")"
}

func (c *Console) call(method, path string, body any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	status, data, err := c.Client.DoAs(ctx, Backend.Actor{User: "Console", Role: "owner"}, method, path, body)
	logged, _ := json.Marshal(body)
	c.Audit.Add(API.AuditEntry{At: time.Now(), User: "console", IP: "local", Action: method, Path: "/api/b" + path, Body: string(logged), Status: status})
	if err != nil {
		return err
	}
	if status < 200 || status > 299 {
		return fmt.Errorf("%s", Backend.Explain(status, data))
	}
	return nil
}

func (c *Console) findPlayer(q string) (Player, error) {
	c.mu.Lock()
	players := append([]Player(nil), c.players...)
	c.mu.Unlock()
	if n, err := strconv.Atoi(q); err == nil {
		if n >= 1 && n <= len(players) {
			return players[n-1], nil
		}
		return Player{}, fmt.Errorf("there is no player #%d in the list", n)
	}
	lower := strings.ToLower(q)
	var partial []Player
	for _, p := range players {
		if strings.ToLower(p.DisplayName) == lower || p.AccountID == q {
			return p, nil
		}
		if strings.Contains(strings.ToLower(p.DisplayName), lower) {
			partial = append(partial, p)
		}
	}
	if len(partial) == 1 {
		return partial[0], nil
	}
	if len(partial) > 1 {
		return Player{}, fmt.Errorf("%q matches %d online players, be more exact", q, len(partial))
	}
	return Player{}, fmt.Errorf("nobody called %q is online", q)
}

func (c *Console) findAccount(q string) (string, string, error) {
	if _, err := strconv.Atoi(q); err == nil {
		p, err := c.findPlayer(q)
		return p.AccountID, p.DisplayName, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var res struct {
		Accounts []struct {
			AccountID   string `json:"accountId"`
			DisplayName string `json:"displayName"`
		} `json:"accounts"`
	}
	if err := c.Client.GetJSON(ctx, "/accounts?q="+url.QueryEscape(q), &res); err != nil {
		return "", "", err
	}
	lower := strings.ToLower(q)
	for _, a := range res.Accounts {
		if strings.ToLower(a.DisplayName) == lower || a.AccountID == q {
			return a.AccountID, a.DisplayName, nil
		}
	}
	switch len(res.Accounts) {
	case 0:
		return "", "", fmt.Errorf("no account matches %q", q)
	case 1:
		return res.Accounts[0].AccountID, res.Accounts[0].DisplayName, nil
	}
	return "", "", fmt.Errorf("%q matches %d accounts, type the full name or account id", q, len(res.Accounts))
}

func (c *Console) findServer(q string) (Server, error) {
	c.mu.Lock()
	servers := append([]Server(nil), c.servers...)
	c.mu.Unlock()
	if n, err := strconv.Atoi(q); err == nil {
		if n >= 1 && n <= len(servers) {
			return servers[n-1], nil
		}
		return Server{}, fmt.Errorf("there is no server #%d in the list", n)
	}
	for _, s := range servers {
		if s.ID == q || strings.HasPrefix(s.ID, q) || s.Address == q {
			return s, nil
		}
	}
	return Server{}, fmt.Errorf("no game server matches %q", q)
}

func (c *Console) resetPassword(_ []string) (string, error) {
	var owner *Backend.AdminUser
	for i := range c.Config.Admins {
		if c.Config.Admins[i].Role == "owner" {
			owner = &c.Config.Admins[i]
			break
		}
	}
	if owner == nil {
		return "", fmt.Errorf("there is no owner account")
	}
	raw := make([]byte, 12)
	rand.Read(raw)
	password := base64.RawURLEncoding.EncodeToString(raw)
	if err := c.Auth.SetAdmin(owner.User, password, owner.Role); err != nil {
		return "", err
	}
	c.Audit.Add(API.AuditEntry{At: time.Now(), User: "console", IP: "local", Action: "reset owner password", Status: 200})
	return fmt.Sprintf("New password for %q: %s%s%s", owner.User, bold, password, reset), nil
}
