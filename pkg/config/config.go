package config

import (
	"fmt"
	"os"
	"strings"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/zalando/go-keyring"
)

// Config holds configuration settings for the CLI including API endpoint,
// authentication tokens, client instance to use the library and output format preferences
type Config struct {
	ApiEndpoint  string
	ProjectToken string
	TeamToken    string
	Client       *chunkify.Client
}

// KeyringServiceKey is the service name used for storing secrets in the system keyring
var KeyringServiceKey = "chunkify-cli"

// SetDefaultTeamToken attempts to set the team token from environment variables first,
// falling back to the keyring if not found in environment
func (cfg *Config) SetDefaultTeamToken() error {
	var err error
	cfg.TeamToken = os.Getenv("CHUNKIFY_TEAM_TOKEN")
	if cfg.TeamToken == "" {
		_, cfg.TeamToken, err = GetToken("TeamToken")
		if err != nil {
			return fmt.Errorf("couldn't get team token, please run `chunkify auth login`")
		}
	}
	return nil
}

// SetDefaultProjectToken attempts to set the project token from environment variables first,
// falling back to the keyring if not found in environment. Requires a default project to be selected.
func (cfg *Config) SetDefaultProjectToken() error {
	var err error
	cfg.ProjectToken = os.Getenv("CHUNKIFY_PROJECT_TOKEN")
	if cfg.ProjectToken != "" {
		return nil
	}

	defaultProject, err := Get("DefaultProject")
	if err != nil {
		return fmt.Errorf("select a project by running `chunkify select`")
	}

	_, cfg.ProjectToken, err = GetToken(defaultProject)
	if err != nil {
		return fmt.Errorf("couldn't get project token, please run `chunkify select`")
	}

	return nil
}

// GetToken retrieves a token from the keyring and returns its ID and value.
// For team tokens, pkey should be "TeamToken". For project tokens, pkey should be the project name.
func GetToken(pkey string) (string, string, error) {
	key := "TeamToken"
	if pkey != "TeamToken" {
		key = "project:" + pkey
	}
	tok, err := Get(key)
	if err != nil {
		return "", "", err
	}

	token := strings.Split(tok, ":")
	return token[0], token[1], nil
}

// SetToken stores a token in the keyring with its ID.
// For team tokens, pkey should be "TeamToken". For project tokens, pkey should be the project name.
func SetToken(pkeyId, pkey, value string) error {
	key := "TeamToken"
	if pkey != "TeamToken" {
		key = "project:" + pkey
	}
	value = pkeyId + ":" + value
	return Set(key, value)
}

// Get retrieves a value from the system keyring using the KeyringServiceKey
func Get(key string) (string, error) {
	//fmt.Println("config get: ", key)
	return keyring.Get(KeyringServiceKey, key)
}

// Set stores a value in the system keyring using the KeyringServiceKey
func Set(key, value string) error {
	//fmt.Println("config set: ", key, value)
	return keyring.Set(KeyringServiceKey, key, value)
}

// DeleteAll removes all stored values for the KeyringServiceKey from the system keyring
func DeleteAll() error {
	return keyring.DeleteAll(KeyringServiceKey)
}
