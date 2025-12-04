package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func exit(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

type Encoder interface {
	Encode([]byte) string
}

type Normalizer interface {
	Normalize(string) string
}

func mustReadSecret(path string) []byte {
	info, err := os.Stat(path)
	if err != nil {
		exit("failed to stat secret file: %v", err)
		return nil
	}

	if info.Mode().Perm() != 0o600 {
		exit("secret file mode %04o, expected 0600", info.Mode().Perm())
		return nil
	}

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		if stat.Uid != uint32(os.Getuid()) {
			exit("secret file not owned by current user (uid=%d, file uid=%d)", os.Getuid(), stat.Uid)
			return nil
		}
		if stat.Mode&0o777 != 0o600 {
			exit("secret file actual mode %04o, expected 0600", stat.Mode&0o777)
			return nil
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		exit("failed to read secret file: %v", err)
		return nil
	}

	s := strings.TrimRight(string(data), "\n")
	return []byte(s)
}

func mustGetDefaultSecretPath() string {
	home := mustUserHome()
	return filepath.Join(home, ".config", "pw", "secret")
}

func pw(site string) string {
	secretPath := mustGetDefaultSecretPath()
	secret := mustReadSecret(secretPath)

	// 讀 site policy（找不到就會是 nil）
	policy := loadSitePolicy(site)
	slog.Debug("loaded site policy", "site", site, "policy", policy)
	var normalizer Normalizer = SegNormalizer{Policy: policy}

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(site))
	digest := h.Sum(nil)

	var encoder Encoder = Base62Encoder{}
	encoded := encoder.Encode(digest)

	length := 12
	if policy != nil && policy.MaxLen > 0 {
		length = policy.MaxLen
	}
	for len(encoded) < length {
		h2 := hmac.New(sha256.New, secret)
		h2.Write([]byte(encoded))
		digest2 := h2.Sum(nil)
		encoded += encoder.Encode(digest2)
	}

	password := normalizer.Normalize(encoded)

	return password
}

func main() {
	legacyMode := flag.Bool("legacy", false, "use legacy mode")
	debug := flag.Bool("debug", false, "enable debug logging")
	help := flag.Bool("help", false, "show detail help message")

	flag.Parse()

	if *help {
		fmt.Print(cmdUsage())
		return
	}

	if *debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	reader := bufio.NewReader(os.Stdin)
	siteName, err := reader.ReadString('\n')
	if err != nil {
		exit("Error reading stdin: %v", err)
	}
	site := strings.TrimSpace(siteName)
	if site == "" {
		exit("site name is empty")
	}
	site = sanitizeSiteKey(site)

	var password string
	if *legacyMode {
		password = legacy(site)
	} else {
		password = pw(site)
	}

	fmt.Println(password)
}

func cmdUsage() string {
	return `pw - deterministic HMAC password generator

Usage:
  pw [-legacy] [-debug]
  echo <site> | pw

Secret:
  ~/.config/pw/secret   (mode 0600, high-entropy key required)

Site policy:
  ~/.config/pw/sites/<site>.conf
  Keys:
    len=N              密碼長度 (default 12)
    upper=true/false   需大寫
    lower=true/false   需小寫
    digit=true/false   需數字
    special=true/false 需特殊符號
    special_chars=...  特殊符號集合 (default "!@#$%^&*")

Example policy:
  len=20
  upper=true
  lower=true
  digit=true
  special=true
  special_chars=!@#$%^&*

Flags:
  -legacy    使用 legacy 密碼模式
  -debug     顯示 debug (包含 site 名稱、組成分析)
  -help      顯示本訊息
`
}
