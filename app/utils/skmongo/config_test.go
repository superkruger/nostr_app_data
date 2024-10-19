package skmongo

import (
	"testing"
	"time"

	"github.com/pascaldekloe/goe/verify"
)

func TestIsComplete(t *testing.T) {
	cases := map[string]struct {
		name     string
		pwd      string
		database string
		host     string
		expected bool
	}{
		"missing name":     {"", "pwd", "db", "host", false},
		"missing password": {"name", "", "db", "host", false},
		"missing database": {"name", "pwd", "", "host", false},
		"missing host":     {"name", "pwd", "db", "", false},
		"all filled in":    {"name", "pwd", "db", "host", true},
	}

	for name, testCase := range cases {
		cfg := UserPasswordConfig{
			username: testCase.name,
			password: testCase.pwd,
			Database: testCase.database,
			Host:     testCase.host}
		got := cfg.IsComplete()

		if got != testCase.expected {
			t.Errorf("%v: got %v, expected %v", name, got, testCase.expected)
		}
	}
}

func TestNewConfig(t *testing.T) {
	tests := map[string]struct {
		username   string
		password   string
		Host       string
		Database   string
		wantConfig UserPasswordConfig
	}{
		"happy flow": {
			wantConfig: UserPasswordConfig{
				username: "name",
				password: "pwd",
				Host:     "host",
				Database: "db",
				timeouts: Timeouts{
					ServerSelection: 3 * time.Second,
					Connect:         1 * time.Second,
					Socket:          10 * time.Second,
					MaxIdleTimeMS:   15 * time.Minute,
				},
			},
			username: "name",
			password: "pwd",
			Host:     "host",
			Database: "db",
		}}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewConfig(testCase.username, testCase.password, testCase.Database, testCase.Host)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			verify.Values(t, "config", got.timeouts, testCase.wantConfig.timeouts)
		})
	}
}
