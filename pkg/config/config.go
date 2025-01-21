package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

type Config struct {
	ApiEndpoint      string
	ProjectToken     string
	AccountToken     string
	DefaultProjectId string
	JSON             bool
	Debug            bool
}

var KeyringServiceKey = "chunkify-cli"

// we check for env variable first, then keyring
func (cfg *Config) SetDefaultTokens() error {
	var err error
	cfg.AccountToken = os.Getenv("LEVEL63_ACCOUNT_TOKEN")
	if cfg.AccountToken == "" {
		_, cfg.AccountToken, err = GetToken("AccountToken")
		if err != nil {
			return fmt.Errorf("couldn't get account token, please run `chunkify auth login`")
		}
	}

	cfg.ProjectToken = os.Getenv("LEVEL63_PROJECT_TOKEN")
	if cfg.ProjectToken != "" {
		return nil
	}

	if cfg.DefaultProjectId == "" {
		cfg.DefaultProjectId, err = Get("DefaultProject")
		if err != nil {
			return fmt.Errorf("select a project by running `chunkify select`")
		}
	}

	_, cfg.ProjectToken, err = GetToken(cfg.DefaultProjectId)
	if err != nil {
		return fmt.Errorf("couldn't get project token, please run `chunkify select`")
	}

	return nil
}

// return the tokenId, Token, error if any
func GetToken(pkey string) (string, string, error) {
	key := "AccountToken"
	if pkey != "AccountToken" {
		key = "project:" + pkey
	}
	tok, err := Get(key)
	if err != nil {
		return "", "", err
	}

	token := strings.Split(tok, ":")
	return token[0], token[1], nil
}

func SetToken(pkeyId, pkey, value string) error {
	key := "AccountToken"
	if pkey != "AccountToken" {
		key = "project:" + pkey
	}
	value = pkeyId + ":" + value
	return Set(key, value)
}

func Get(key string) (string, error) {
	//fmt.Println("config get: ", key)
	return keyring.Get(KeyringServiceKey, key)
}

func Set(key, value string) error {
	//fmt.Println("config set: ", key, value)
	return keyring.Set(KeyringServiceKey, key, value)
}

func DeleteAll() error {
	return keyring.DeleteAll(KeyringServiceKey)
}
