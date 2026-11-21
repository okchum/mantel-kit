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

// ExpandEnv turns "$NAME" or "${NAME}" into the environment value; other
// strings pass through, with "$NAME" inside them expanded ("Bearer $TOKEN").
func ExpandEnv(s string) string {
	if strings.HasPrefix(s, "$") {
		return os.Getenv(envName(s))
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
	if strings.HasPrefix(s, "$") {
		check(envName(s))
		return names
	}
	os.Expand(s, check)
	return names
}

// envName is the variable a whole-value "$NAME" or "${NAME}" names.
func envName(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(s, "$"), "{"), "}")
}

// ParseWindow parses "HH:MM-HH:MM" into minutes since midnight.
func ParseWindow(s string) (start, end int, err error) {
	a, b, ok := strings.Cut(s, "-")
	if !ok {
		return 0, 0, errors.New(`want "HH:MM-HH:MM"`)
	}
	if start, err = parseHM(a); err != nil {
		return
	}
	end, err = parseHM(b)
	return
}

func parseHM(s string) (int, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(s))
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

// InWindow reports whether t's local clock time falls inside the window (wrapping past midnight).
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
