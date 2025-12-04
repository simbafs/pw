package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type SitePolicy struct {
	Site           string
	MaxLen         int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
	SpecialChars   string
}

func sanitizeSiteKey(site string) string {
	s := strings.TrimSpace(strings.ToLower(site))
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func mustUserHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		exit("failed to get user home directory: %v", err)
	}
	return home
}

func siteConfigPath(site string) string {
	home := mustUserHome()
	return filepath.Join(home, ".config", "pw", "sites", sanitizeSiteKey(site)+".conf")
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y", "on":
		return true
	}
	return false
}

func loadSitePolicy(site string) *SitePolicy {
	path := siteConfigPath(site)
	data, err := os.ReadFile(path)
	if err != nil {
		// 找不到 / 讀不到就當沒設定
		return nil
	}

	p := &SitePolicy{
		Site: site,
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		switch key {
		case "length", "len", "n":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				p.MaxLen = n
			}
		case "require_upper", "upper":
			p.RequireUpper = parseBool(val)
		case "require_lower", "lower":
			p.RequireLower = parseBool(val)
		case "require_digit", "digit":
			p.RequireDigit = parseBool(val)
		case "require_special", "special":
			p.RequireSpecial = parseBool(val)
		case "special_chars":
			p.SpecialChars = val
		}
	}

	if p.SpecialChars == "" {
		p.SpecialChars = "!@#$%^&*"
	}

	return p
}

func applyPolicy(s string, p *SitePolicy) string {
	b := []byte(s)

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, ch := range b {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	slog.Debug("current password composition", "hasUpper", hasUpper, "hasLower", hasLower, "hasDigit", hasDigit, "hasSpecial", hasSpecial)
	slog.Debug("policy requirements", "RequireUpper", p.RequireUpper, "RequireLower", p.RequireLower, "RequireDigit", p.RequireDigit, "RequireSpecial", p.RequireSpecial)

	// 決定性挑一個 index 來改（不隨機，同 site 每次一樣）
	pickIndex := func(label string) int {
		if len(b) == 0 {
			return 0
		}
		h := sha256.Sum256([]byte(p.Site + "|" + label))
		return int(h[0]) % len(b)
	}

	// 缺 upper
	if p.RequireUpper && !hasUpper && len(b) > 0 {
		i := pickIndex("upper")
		orig := b[i]
		idx := max(strings.IndexByte(base62Alphabet, orig), 0)
		b[i] = byte('A' + (idx % 26))
		slog.Debug("adding required upper case letter", "index", i, "original", string(orig), "new", string(b[i]))
		hasUpper = true
	}

	// 缺 lower
	if p.RequireLower && !hasLower && len(b) > 0 {
		i := pickIndex("lower")
		orig := b[i]
		idx := max(strings.IndexByte(base62Alphabet, orig), 0)
		b[i] = byte('a' + (idx % 26))
		slog.Debug("adding required lower case letter", "index", i, "original", string(orig), "new", string(b[i]))
		hasLower = true
	}

	// 缺 digit
	if p.RequireDigit && !hasDigit && len(b) > 0 {
		i := pickIndex("digit")
		orig := b[i]
		idx := max(strings.IndexByte(base62Alphabet, orig), 0)
		b[i] = byte('0' + (idx % 10))
		slog.Debug("adding required digit", "index", i, "original", string(orig), "new", string(b[i]))
		hasDigit = true
	}

	// 缺 special
	if p.RequireSpecial && !hasSpecial && len(b) > 0 && len(p.SpecialChars) > 0 {
		i := pickIndex("special")
		h := sha256.Sum256([]byte(p.Site + "|special_char"))
		j := int(h[1]) % len(p.SpecialChars)
		b[i] = p.SpecialChars[j]
		slog.Debug("adding required special character", "index", i, "new", string(b[i]))
		hasSpecial = true
	}

	return string(b)
}
