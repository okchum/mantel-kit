package conf

import (
	"encoding/json"
	"slices"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	var v struct{ D Duration }
	if err := json.Unmarshal([]byte(`{"D":"1m30s"}`), &v); err != nil || v.D.Duration != 90*time.Second {
		t.Fatalf("got %v, %v", v.D, err)
	}
	if err := json.Unmarshal([]byte(`{"D":""}`), &v); err != nil || v.D.Duration != 0 {
		t.Fatalf("empty: got %v, %v", v.D, err)
	}
	for _, bad := range []string{`{"D":30}`, `{"D":"soon"}`} {
		if err := json.Unmarshal([]byte(bad), &v); err == nil {
			t.Errorf("%s: no error", bad)
		}
	}
	b, err := json.Marshal(Duration{5 * time.Minute})
	if err != nil || string(b) != `"5m0s"` {
		t.Errorf("marshal: %s, %v", b, err)
	}
	if got := (Duration{}).Or(time.Hour); got != time.Hour {
		t.Errorf("zero Or: %v", got)
	}
	if got := (Duration{time.Second}).Or(time.Hour); got != time.Second {
		t.Errorf("set Or: %v", got)
	}
}

func TestExpandEnv(t *testing.T) {
	t.Setenv("KIT_A", "alpha")
	t.Setenv("KIT_EMPTY", "")
	for in, want := range map[string]string{
		"$KIT_A":           "alpha",
		"${KIT_A}":         "alpha",
		"$KIT_UNSET_XYZ":   "",
		"Bearer $KIT_A":    "Bearer alpha",
		"Bearer ${KIT_A}!": "Bearer alpha!",
		"plain":            "plain",
		"":                 "",
		"x-$KIT_EMPTY-y":   "x--y",
	} {
		if got := ExpandEnv(in); got != want {
			t.Errorf("ExpandEnv(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEmptyEnv(t *testing.T) {
	t.Setenv("KIT_A", "alpha")
	t.Setenv("KIT_EMPTY", "")
	for in, want := range map[string][]string{
		"$KIT_A":                          nil,
		"${KIT_EMPTY}":                    {"KIT_EMPTY"},
		"$KIT_UNSET_XYZ":                  {"KIT_UNSET_XYZ"},
		"Bearer $KIT_A":                   nil,
		"a $KIT_EMPTY b ${KIT_UNSET_XYZ}": {"KIT_EMPTY", "KIT_UNSET_XYZ"},
		"plain":                           nil,
	} {
		if got := EmptyEnv(in); !slices.Equal(got, want) {
			t.Errorf("EmptyEnv(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseWindow(t *testing.T) {
	s, e, err := ParseWindow("23:30-07:05")
	if err != nil || s != 23*60+30 || e != 7*60+5 {
		t.Fatalf("got %d, %d, %v", s, e, err)
	}
	if _, _, err := ParseWindow(" 04:55 - 05:20 "); err != nil {
		t.Errorf("spaces: %v", err)
	}
	for _, bad := range []string{"", "23:30", "25:00-01:00", "07:00-7pm"} {
		if _, _, err := ParseWindow(bad); err == nil {
			t.Errorf("%q: no error", bad)
		}
	}
}

func TestInWindow(t *testing.T) {
	at := func(h, m int) time.Time { return time.Date(2026, 1, 2, h, m, 0, 0, time.UTC) }
	for _, c := range []struct {
		w    string
		t    time.Time
		want bool
	}{
		{"09:00-17:00", at(9, 0), true},
		{"09:00-17:00", at(16, 59), true},
		{"09:00-17:00", at(17, 0), false},
		{"09:00-17:00", at(8, 59), false},
		{"23:30-07:00", at(23, 45), true},
		{"23:30-07:00", at(3, 0), true},
		{"23:30-07:00", at(7, 0), false},
		{"23:30-07:00", at(12, 0), false},
		{"", at(12, 0), false},
		{"bad", at(12, 0), false},
	} {
		if got := InWindow(c.w, c.t); got != c.want {
			t.Errorf("InWindow(%q, %s) = %v", c.w, c.t.Format("15:04"), got)
		}
	}
}
