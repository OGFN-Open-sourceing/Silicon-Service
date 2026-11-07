package Console

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	dim    = "\x1b[2m"
	red    = "\x1b[31m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	blue   = "\x1b[34m"
	cyan   = "\x1b[36m"
	invert = "\x1b[7m"
)

type screen struct {
	w     int
	lines []string
}

func (s *screen) add(format string, args ...any) {
	s.lines = append(s.lines, fmt.Sprintf(format, args...))
}

func (s *screen) blank() { s.lines = append(s.lines, "") }

func (s *screen) section(title, right string) {
	left := " " + bold + cyan + title + reset
	if right != "" {
		left += dim + "  " + right + reset
	}
	s.add("%s", left)
}

func (s *screen) boxTop(title, right string) {
	fill := s.w - 2 - 3 - vlen(title) - vlen(right) - 3
	s.add("%s╭─ %s%s%s %s%s %s%s ─╮%s", dim, reset+bold+cyan, title, reset+dim, strings.Repeat("─", max(1, fill)), reset+bold, right, reset+dim, reset)
}

func (s *screen) boxRow(label, value string) {
	content := fmt.Sprintf("  %s%-11s%s %s", dim, label, reset, value)
	s.add("%s│%s%s%s│%s", dim, reset, pad(content, s.w-2), dim, reset)
}

func (s *screen) boxBottom() {
	s.add("%s╰%s╯%s", dim, strings.Repeat("─", s.w-2), reset)
}

func (c *Console) draw() {
	c.drawMu.Lock()
	defer c.drawMu.Unlock()
	if c.restore == nil {
		return
	}
	w, h := termSize()
	w = clamp(w-1, 60, 140)
	h = max(h, 20)

	c.mu.Lock()
	showKey, notice, noticeBad, noticeEnd := c.showKey, c.notice, c.noticeBad, c.noticeEnd
	players := append([]Player(nil), c.players...)
	servers := append([]Server(nil), c.servers...)
	f, answers, input, confirming, busy := c.flow, append([]string(nil), c.answers...), c.input, c.confirm, c.busy
	resized := w != c.lastW || h != c.lastH
	c.lastW, c.lastH = w, h
	c.mu.Unlock()

	s := &screen{w: w}
	sample, polled := c.Monitor.Latest()
	up := polled && sample.Up

	s.boxTop("SILICON-SERVICE "+c.Version, c.Config.Backend.Name)
	s.boxRow("Dashboard", bold+c.URL+reset)
	state := yellow + "● connecting" + reset
	switch {
	case up:
		state = green + "● online" + reset
		if sample.Status != nil {
			state += dim + " · up " + uptime(sample.Status.UptimeSeconds) + reset
		}
	case polled:
		state = red + "● offline" + reset
	}
	s.boxRow("Backend", c.Config.Backend.URL+"  "+state)
	if polled && !up && sample.Error != "" {
		for _, part := range wrap(sample.Error, w-18) {
			s.boxRow("", red+part+reset)
		}
	}
	if up && sample.Status != nil {
		st := sample.Status
		s.boxRow("Players", fmt.Sprintf("%s%d%s online %s·%s %d queued %s·%s %d in a match %s·%s %d game server%s",
			bold+green, st.Online, reset, dim, reset, st.InQueue, dim, reset, st.InMatch, dim, reset, st.Servers, plural(st.Servers)))
	} else {
		s.boxRow("Players", dim+"-"+reset)
	}
	s.boxRow("Last hour", c.sparkline(min(48, w-34)))
	if showKey {
		s.boxRow("API key", yellow+c.Config.Backend.APIKey+reset)
		s.boxRow("", dim+"put this in the backend's .env as SILICON_API_KEY"+reset)
	} else {
		s.boxRow("API key", dim+"hidden · press 9 to show"+reset)
	}
	s.boxBottom()

	actionCols := 2
	if w >= 100 {
		actionCols = 3
	}
	actionRows := (len(menu) + actionCols - 1) / actionCols
	footer := 3
	if f != nil {
		footer = 4
	}
	fixed := len(s.lines) + 1 + 1 + 1 + actionRows + 1 + 1 + footer
	avail := h - fixed
	listRows := clamp(avail/2, 2, 10)
	logRows := max(2, avail-listRows-1)

	s.blank()
	c.drawLists(s, players, servers, listRows, up)

	s.blank()
	s.section("ACTIONS", "press a number")
	colW := (w - 2) / actionCols
	for r := 0; r < actionRows; r++ {
		row := " "
		for col := 0; col < actionCols; col++ {
			i := col*actionRows + r
			if i >= len(menu) {
				continue
			}
			a := menu[i]
			cell := fmt.Sprintf(" %s %d %s %s", invert+bold, a.key, reset, c.choiceLabel(a))
			row += pad(cell, colW)
		}
		s.add("%s", row)
	}
	s.add("  %s q %s %s", invert+bold, reset, "Quit")

	s.blank()
	s.section("ACTIVITY", "")
	c.drawActivity(s, logRows)

	s.add("%s%s%s", dim, strings.Repeat("─", w), reset)
	switch {
	case f != nil && confirming:
		s.add(" %s?%s %s %s[y/n]%s", bold+yellow, reset, f.confirm(answers), bold, reset)
		s.add("   %sEsc cancels%s", dim, reset)
		s.blank()
	case f != nil:
		st := f.steps[len(answers)]
		s.add(" %s▸ %s%s %s· step %d of %d%s", bold+cyan, f.title, reset, dim, len(answers)+1, len(f.steps), reset)
		shown := input
		if room := w - vlen(st.prompt) - 8; utf8.RuneCountInString(shown) > room {
			r := []rune(shown)
			shown = "…" + string(r[len(r)-room+1:])
		}
		s.add("   %s: %s%s%s▌%s", st.prompt, bold, shown, reset+cyan, reset)
		hint := "Enter to continue · Esc to cancel"
		if st.optional {
			hint = "Enter to skip or continue · Esc to cancel"
		}
		s.add("   %s%s%s", dim, hint, reset)
	case busy != "":
		s.add(" %s%s%s", yellow, busy, reset)
		s.blank()
	case notice != "" && time.Now().Before(noticeEnd):
		color := green
		if noticeBad {
			color = red
		}
		for i, part := range wrap(notice, w-3) {
			if i == 2 {
				break
			}
			s.add(" %s%s%s", color, part, reset)
		}
		if len(wrap(notice, w-3)) < 2 {
			s.blank()
		}
	default:
		s.add(" %sReady.%s", dim, reset)
		s.blank()
	}

	var b strings.Builder
	if resized {
		b.WriteString("\x1b[2J")
	}
	b.WriteString("\x1b[H")
	for i, l := range s.lines {
		if i >= h {
			break
		}
		b.WriteString(l)
		b.WriteString(reset + "\x1b[K")
		if i < len(s.lines)-1 && i < h-1 {
			b.WriteString("\r\n")
		}
	}
	b.WriteString("\x1b[J")
	os.Stdout.WriteString(b.String())
}

func (c *Console) drawLists(s *screen, players []Player, servers []Server, rows int, up bool) {
	playerLines := []string{}
	for i, p := range players {
		where := p.State
		if p.Playlist != "" {
			where += " · " + shortPlaylist(p.Playlist)
		}
		color := dim
		switch p.State {
		case "in match":
			color = green
		case "in queue", "queue":
			color = yellow
		}
		playerLines = append(playerLines, fmt.Sprintf("%s%3d%s  %-16s %s%s%s", dim, i+1, reset, clip(p.DisplayName, 16), color, clip(where, 28), reset))
	}
	serverLines := []string{}
	for i, sv := range servers {
		color := yellow
		if sv.State == "in match" {
			color = green
		}
		serverLines = append(serverLines, fmt.Sprintf("%s%3d%s  %-21s %s%-8s%s %d/%d %s%s%s",
			dim, i+1, reset, clip(sv.Address, 21), color, clip(sv.State, 8), reset, sv.Players, sv.MaxPlayers, dim, clip(shortPlaylist(sv.Playlist), 14), reset))
	}
	empty := func(what string) string {
		if !up {
			return dim + "     backend not reachable" + reset
		}
		return dim + "     no " + what + reset
	}
	fit := func(lines []string, what string) []string {
		if len(lines) == 0 {
			return []string{empty(what)}
		}
		if len(lines) > rows {
			more := len(lines) - rows + 1
			lines = append(lines[:rows-1:rows-1], fmt.Sprintf("%s     + %d more (dashboard has the full list)%s", dim, more, reset))
		}
		return lines
	}
	pl := fit(playerLines, "players online")
	sl := fit(serverLines, "game servers")
	pTitle := fmt.Sprintf("ONLINE PLAYERS %s(%d)%s", dim, len(players), reset)
	sTitle := fmt.Sprintf("GAME SERVERS %s(%d)%s", dim, len(servers), reset)

	if s.w >= 110 {
		half := s.w / 2
		s.add(" %s%s%s%s", bold+cyan, pad(pTitle, half-1), bold+cyan, sTitle+reset)
		for i := 0; i < rows; i++ {
			left, right := "", ""
			if i < len(pl) {
				left = pl[i]
			}
			if i < len(sl) {
				right = sl[i]
			}
			if left == "" && right == "" {
				break
			}
			s.add(" %s%s", pad(left, half-1), right)
		}
		return
	}
	s.section(pTitle, "")
	for _, l := range pl {
		s.add(" %s", l)
	}
	s.section(sTitle, "")
	for _, l := range sl {
		s.add(" %s", l)
	}
}

func (c *Console) drawActivity(s *screen, rows int) {
	type item struct {
		at   time.Time
		text string
	}
	var items []item
	for _, e := range c.Audit.Recent(rows) {
		what := strings.TrimSpace(e.Action + " " + e.Path)
		color := green
		if e.Status >= 400 || e.Status == 0 {
			color = red
		}
		status := ""
		if e.Status != 0 {
			status = fmt.Sprintf(" %s%d%s", color, e.Status, reset)
		}
		items = append(items, item{e.At, fmt.Sprintf("%s%s%s  %s%-10s%s %s%s", dim, e.At.Format("15:04:05"), reset, blue, clip(e.User, 10), reset, clip(what, s.w-34), status)})
	}
	for _, l := range c.Logs.Last(rows) {
		at, text := splitLogTime(l)
		items = append(items, item{at, fmt.Sprintf("%s%s%s  %s%-10s%s %s", dim, at.Format("15:04:05"), reset, dim, "log", reset, clip(text, s.w-24))})
	}
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].at.Before(items[j-1].at); j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
	if len(items) > rows {
		items = items[len(items)-rows:]
	}
	if len(items) == 0 {
		s.add("     %snothing yet%s", dim, reset)
		rows--
	}
	for _, it := range items {
		s.add(" %s", it.text)
	}
	for i := len(items); i < rows; i++ {
		s.blank()
	}
}

func splitLogTime(l string) (time.Time, string) {
	if len(l) >= 9 && l[2] == ':' && l[5] == ':' && l[8] == ' ' {
		if t, err := time.ParseInLocation("15:04:05", l[:8], time.Local); err == nil {
			now := time.Now()
			return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local), l[9:]
		}
	}
	return time.Now(), l
}

func (c *Console) sparkline(width int) string {
	samples := c.Monitor.Since(time.Now().Add(-time.Hour), width)
	if len(samples) == 0 {
		return dim + "no samples yet" + reset
	}
	peak := 0
	for _, x := range samples {
		if x.Up && x.Status != nil && x.Status.Online > peak {
			peak = x.Status.Online
		}
	}
	bars := []rune("▁▂▃▄▅▆▇█")
	var b strings.Builder
	b.WriteString(green)
	for _, x := range samples {
		if !x.Up || x.Status == nil {
			b.WriteString(red + "·" + green)
			continue
		}
		i := 0
		if peak > 0 {
			i = x.Status.Online * (len(bars) - 1) / peak
		}
		b.WriteRune(bars[i])
	}
	b.WriteString(reset + dim + fmt.Sprintf("  peak %d online", peak) + reset)
	return b.String()
}

func shortPlaylist(p string) string {
	p = strings.TrimPrefix(p, "Playlist_")
	p = strings.TrimPrefix(p, "playlist_")
	return p
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func vlen(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			for i < len(s) && !(s[i] >= 'A' && s[i] <= 'Z' || s[i] >= 'a' && s[i] <= 'z') {
				i++
			}
			i++
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		n++
	}
	return n
}

func pad(s string, w int) string {
	if n := vlen(s); n < w {
		return s + strings.Repeat(" ", w-n)
	}
	return s
}

func clamp(v, lo, hi int) int { return max(lo, min(hi, v)) }

func uptime(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	if d >= 24*time.Hour {
		return fmt.Sprintf("%dd %dh", int(d.Hours())/24, int(d.Hours())%24)
	}
	if d >= time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
}

func clip(s string, n int) string {
	if n < 1 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func wrap(s string, n int) []string {
	var out []string
	cur := ""
	for _, w := range strings.Fields(s) {
		if cur != "" && vlen(cur)+1+vlen(w) > n {
			out = append(out, cur)
			cur = w
			continue
		}
		if cur != "" {
			cur += " "
		}
		cur += w
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
