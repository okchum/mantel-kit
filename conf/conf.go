// Package conf holds the configuration value types and formats Mantel and
// Mantel Cast share: "30s" durations, "$NAME" values and "HH:MM-HH:MM"
// daily windows.
package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Duration is a time.Duration that unmarshals from strings like "30s" or "5m".
type Duration struct{ time.Duration }

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("duration must be a string like \"30s\": %w", err)
	}
	if s == "" {
		d.Duration = 0
		return nil
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = v
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) { return json.Marshal(d.String()) }

// Or returns d, or def when d is zero.
func (d Duration) Or(def time.Duration) time.Duration {
	if d.Duration <= 0 {
		return def
	}
	return d.Duration
}

// ExpandEnv expands "$NAME" and "${NAME}" references to environment values
// wherever they appear: a whole value ("$TOKEN"), a prefix ("$HOME/ca.pem")
// or inside text ("Bearer $TOKEN"). Unset variables expand to "".
func ExpandEnv(s string) string {
	if name, ok := wholeName(s); ok {
		return os.Getenv(name)
	}
	return os.Expand(s, os.Getenv)
}

// EmptyEnv lists the environment variables s refers to (as ExpandEnv reads
// them) that are unset or empty, so a token like "$MANTEL_TOKEN" that
// silently became "" can be warned about.
func EmptyEnv(s string) []string {
	var names []string
	check := func(name string) string {
		if os.Getenv(name) == "" {
			names = append(names, name)
		}
		return ""
	}
	if name, ok := wholeName(s); ok {
		check(name)
		return names
	}
	os.Expand(s, check)
	return names
}

// wholeName reports the variable when all of s is one reference, "$NAME" or
// "${NAME}" with NAME a shell identifier. Anything else ("$HOME/ca.pem",
// "${A}-${B}") is left to os.Expand, which expands each reference in place.
func wholeName(s string) (string, bool) {
	name, ok := strings.CutPrefix(s, "$")
	if !ok {
		return "", false
	}
	if inner, ok := strings.CutPrefix(name, "{"); ok {
		if name, ok = strings.CutSuffix(inner, "}"); !ok {
			return "", false
		}
	}
	if name == "" {
		return "", false
	}
	for i, r := range name {
		if r != '_' && (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && (i == 0 || r < '0' || r > '9') {
			return "", false
		}
	}
	return name, true
}

// ParseWindow parses "HH:MM-HH:MM" into minutes since midnight. The end is
// exclusive and may be "24:00" (end of day), so "00:00-24:00" is the whole
// day and "22:00-24:00" runs up to midnight.
func ParseWindow(s string) (start, end int, err error) {
	a, b, ok := strings.Cut(s, "-")
	if !ok {
		return 0, 0, errors.New(`want "HH:MM-HH:MM"`)
	}
	if start, err = parseHM(a); err != nil {
		return
	}
	if strings.TrimSpace(b) == "24:00" {
		return start, 24 * 60, nil
	}
	end, err = parseHM(b)
	return
}

// FormatWindow is the canonical "HH:MM-HH:MM" form of a window ParseWindow
// returned, for passing to code with a stricter parser (the web page).
func FormatWindow(start, end int) string {
	return fmt.Sprintf("%02d:%02d-%02d:%02d", start/60, start%60, end/60, end%60)
}

func parseHM(s string) (int, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(s))
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

// InWindow reports whether t's local clock time falls inside the window
// (start inclusive, end exclusive, wrapping past midnight).
func InWindow(window string, t time.Time) bool {
	start, end, err := ParseWindow(window)
	if err != nil || window == "" {
		return false
	}
	m := t.Hour()*60 + t.Minute()
	if start <= end {
		return m >= start && m < end
	}
	return m >= start || m < end
}
