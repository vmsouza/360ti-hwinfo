package main

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config configures the report look: logo, names and colors.
type Config struct {
	AppName     string            `json:"app_name"`     // application name
	CompanyName string            `json:"company_name"` // company / vendor name
	Logo        string            `json:"logo"`         // path (relative to the exe folder)
	LogoBase64  string            `json:"logo_base64"`  // optional raw base64 or data URI
	Colors      map[string]string `json:"colors"`
}

// resolvedConfig holds the config with computed logo data URI and CSS.
type resolvedConfig struct {
	AppName     string
	CompanyName string
	LogoDataURI string
	ThemeCSS    string
	AccentColor [3]int
}

//go:embed logo360ti.png
var defaultLogo []byte

func defaultColors() map[string]string {
	return map[string]string{
		"accent":     "#2563eb",
		"accent2":    "#3b82f6",
		"accent3":    "#0ea5e9",
		"accent4":    "#60a5fa",
		"background": "#f3f5f9",
		"sidebar":    "#ffffff",
		"surface":    "#ffffff",
		"surface2":   "#f1f5f9",
		"border":     "#e5e9f0",
		"text":       "#0f172a",
		"muted":      "#64748b",
	}
}

// LoadConfig reads config.json next to the executable and resolves the logo
// (base64 data URI) and the theme CSS overrides.
func LoadConfig() *resolvedConfig {
	rc := &resolvedConfig{
		AppName:     "360ti HWiNFO",
		CompanyName: "360ti",
	}

	cfg := &Config{}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if b, err := os.ReadFile(filepath.Join(dir, "config.json")); err == nil {
			_ = json.Unmarshal(b, cfg)
		}
	}

	if cfg.AppName != "" {
		rc.AppName = cfg.AppName
	}
	if cfg.CompanyName != "" {
		rc.CompanyName = cfg.CompanyName
	}

	// resolve logo
	var logo []byte
	if cfg.LogoBase64 != "" {
		s := strings.TrimPrefix(cfg.LogoBase64, "data:image/png;base64,")
		if raw, err := base64.StdEncoding.DecodeString(s); err == nil {
			logo = raw
		}
	}
	if len(logo) == 0 && cfg.Logo != "" {
		if exe, err := os.Executable(); err == nil {
			if b, err := os.ReadFile(filepath.Join(filepath.Dir(exe), cfg.Logo)); err == nil {
				logo = b
			}
		}
	}
	if len(logo) == 0 {
		logo = defaultLogo
	}
	rc.LogoDataURI = "data:image/png;base64," + base64.StdEncoding.EncodeToString(logo)

	// merge colors over defaults
	colors := defaultColors()
	for k, v := range cfg.Colors {
		if strings.TrimSpace(v) != "" {
			colors[k] = v
		}
	}
	rc.ThemeCSS = buildThemeCSS(colors)
	rc.AccentColor = hexToRGB(colors["accent"])
	if rc.AccentColor == [3]int{} {
		rc.AccentColor = [3]int{37, 99, 235}
	}
	return rc
}

// hexToRGB converts "#rrggbb" into an [r,g,b] triplet.
func hexToRGB(hex string) [3]int {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) != 6 {
		return [3]int{}
	}
	var c [3]int
	for i := 0; i < 3; i++ {
		v, err := strconv.ParseUint(hex[i*2:i*2+2], 16, 8)
		if err != nil {
			return [3]int{}
		}
		c[i] = int(v)
	}
	return c
}

func buildThemeCSS(c map[string]string) string {
	vars := []struct{ cfg, css string }{
		{"accent", "--accent1"},
		{"accent2", "--accent2"},
		{"accent3", "--accent3"},
		{"accent4", "--accent4"},
		{"background", "--bg"},
		{"sidebar", "--sidebar"},
		{"surface", "--surface"},
		{"surface2", "--surface-2"},
		{"border", "--border"},
		{"text", "--text"},
		{"muted", "--muted"},
	}
	var b strings.Builder
	b.WriteString(":root{\n")
	for _, v := range vars {
		fmt.Fprintf(&b, "%s:%s;\n", v.css, c[v.cfg])
	}
	fmt.Fprintf(&b, "--grad:linear-gradient(135deg,%s,%s 45%%,%s);\n",
		c["accent"], c["accent2"], c["accent3"])
	b.WriteString("}")
	return b.String()
}