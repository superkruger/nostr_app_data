package skmongo

import (
	"fmt"
	"net/url"

	"github.com/superkruger/nostr_app_data/app/utils/aws/secrets"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config is the Mongo DB configuration
type Config interface {
	IsComplete() bool
	DatabaseName() string
	toClientOptions() *options.ClientOptions
	WithTimeouts(timeouts Timeouts) Config
	WithClientOptions(clientOptions *options.ClientOptions) Config
}

// UserPasswordConfig is the Mongo DB configuration using username and password
type UserPasswordConfig struct {
	username      string
	password      string
	Database      string
	Host          string
	timeouts      Timeouts
	clientOptions *options.ClientOptions
}

// IsComplete checks if the config is complete
func (c UserPasswordConfig) IsComplete() bool {
	if c.Database == "" || c.Host == "" || c.username == "" || c.password == "" {
		return false
	}
	return true
}

// DatabaseName returns the name of the database in the configuration
func (c UserPasswordConfig) DatabaseName() string {
	return c.Database
}

// WithTimeouts sets the timeouts for the config
func (c UserPasswordConfig) WithTimeouts(timeouts Timeouts) Config {
	c.timeouts = timeouts
	return c
}

// WithClientOptions sets the additional clientOptions for the config
func (c UserPasswordConfig) WithClientOptions(clientOptions *options.ClientOptions) Config {
	c.clientOptions = clientOptions
	return c
}

// toClientOptions constructs the mongo ClientOptions corresponding with the config
func (c UserPasswordConfig) toClientOptions() *options.ClientOptions {
	u := &url.URL{
		Scheme: "mongodb+srv",
		User:   url.UserPassword(c.username, c.password),
		Host:   c.Host,
	}
	opts := options.Client().
		ApplyURI(u.String()).
		SetRetryWrites(true).
		SetMaxConnecting(10)
	// For read, we can use
	// SetReadPreference(readpref.SecondaryPreferred(readpref.WithMaxStaleness(90 * time.Second)))

	c.timeouts.setValuesIn(opts)
	return options.MergeClientOptions(opts, c.clientOptions)
}

// ConfigFromSecret creates the config from the secret
func ConfigFromSecret(secretName string) (UserPasswordConfig, error) {
	var c struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		Host     string `json:"host"`
	}
	err := secrets.NewService().GetAndUnmarshal(secretName, &c)
	if err != nil {
		return UserPasswordConfig{}, err
	}
	cfg, err := NewConfig(c.Username, c.Password, c.Database, c.Host)
	if err != nil {
		return cfg, ErrInvalidConfig{secretName: secretName}
	}
	return cfg, nil
}

// NewConfig creates a config based on the username and password
func NewConfig(username string, password string, database string, host string) (UserPasswordConfig, error) {
	cfg := UserPasswordConfig{
		username: username,
		password: password,
		Database: database,
		Host:     host,
		timeouts: timeoutsForLambda(),
	}
	if !cfg.IsComplete() {
		return cfg, ErrInvalidConfig{}
	}
	return cfg, nil
}

// ErrInvalidConfig is the error thrown when the config is invalid or incomplete
type ErrInvalidConfig struct {
	secretName string
}

// Error implements the error interface, and returns a string representation
func (c ErrInvalidConfig) Error() string {
	if c.secretName == "" {
		return "the mongo db config is invalid or incomplete"
	}
	return fmt.Sprintf("the mongo db config behind secret '%v' is invalid or incomplete", c.secretName)
}
