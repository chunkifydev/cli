package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

// Config holds configuration settings for the CLI including
// authentication tokens, client instance to use the library and profiles
type Config struct {
	Token     string
	Client    *chunkify.Client
	Profile   string
	StorageID string
}

func (cfg *Config) ConfigKey(key string) string {
	if cfg.Profile != "" {
		return cfg.Profile + "." + key
	}
	return key
}

// KeyringServiceKey is the service name used for storing secrets in the system keyring
const (
	KeyringServiceKey  = "chunkify-cli"
	ConfigEndpointKey  = "config.endpoint"
	ConfigTokenKey     = "config.token"
	ConfigStorageIDKey = "config.storage-id"
)

// LoadStorageID reads the selected profile's optional storage override.
func (cfg *Config) LoadStorageID() error {
	cfg.StorageID = ""
	value, err := Get(cfg.ConfigKey(ConfigStorageIDKey))
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading storage-id config: %w", err)
	}
	cfg.StorageID = value
	return nil
}

// SetToken attempts to set the project token from environment variables first,
// falling back to the keyring if not found in environment.
func (cfg *Config) SetToken() error {
	var err error
	cfg.Token = os.Getenv("CHUNKIFY_TOKEN")
	if cfg.Token != "" {
		return nil
	}

	tok, err := Get(cfg.ConfigKey(ConfigTokenKey))
	if err != nil {
		return err
	}

	cfg.Token = tok
	return nil
}

// Get retrieves a value from the system keyring using the KeyringServiceKey
func Get(key string) (string, error) {
	return keyring.Get(KeyringServiceKey, key)
}

// Set stores a value in the system keyring using the KeyringServiceKey
func Set(key, value string) error {
	return keyring.Set(KeyringServiceKey, key, value)
}

// DeleteAll removes all stored values for the KeyringServiceKey from the system keyring
func DeleteAll() error {
	return keyring.DeleteAll(KeyringServiceKey)
}

func NewCommand() *cobra.Command {
	var profile string

	cmd := &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Manage configuration settings",
		Long: `Manage configuration settings for chunkify.

Available configuration keys:
  token     - Chunkify project token
  endpoint  - Chunkify API endpoint URL
  storage-id - Storage for uploads and outputs; overrides both storage ID flags
  delete    - Delete config

Set a profile with --profile <profile> to save different project tokens

Examples:
  chunkify config token                    # Get project token
  chunkify config token sk_project_token   # Set token to sk_project_token
  chunkify config delete                   # Delete config
  chunkify config storage-id stor_aws_id   # Set CLI storage
  chunkify config storage-id ""            # Clear CLI storage

  Use a specific profile
  chunkify config token sk_project_token --profile your_profile
`,
		Args: cobra.RangeArgs(1, 2),

		RunE: func(cmd *cobra.Command, args []string) error {
			configKeyPrefix := ""

			if profile != "" {
				profile = strings.TrimSpace(profile)
				profile = strings.ToLower(profile)

				if len(profile) > 20 {
					return fmt.Errorf("profile name is too long. It should be less than 20 characters")
				}

				// check if the profile is a valid: a-z0-9_
				if !regexp.MustCompile(`^[a-z0-9_]+$`).MatchString(profile) {
					return fmt.Errorf("invalid profile: %s. It should only contain letters, numbers and underscores", profile)
				}
				configKeyPrefix = profile + "."
			}

			key := args[0]

			switch key {
			case "storage-id":
				configKey := configKeyPrefix + ConfigStorageIDKey
				if len(args) == 1 {
					value, err := Get(configKey)
					if err != nil {
						return fmt.Errorf("%s not found", configKey)
					}
					cmd.Println(configKey, "=", value)
					return nil
				}
				value := strings.TrimSpace(args[1])
				if value == "" {
					if err := keyring.Delete(KeyringServiceKey, configKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
						return err
					}
					cmd.Println("Cleared", configKey)
					return nil
				}
				if len(value) > 64 || !regexp.MustCompile(`^stor_[a-zA-Z0-9_-]+$`).MatchString(value) {
					return fmt.Errorf("invalid storage-id: expected a storage ID starting with 'stor_' and at most 64 characters")
				}
				if err := Set(configKey, value); err != nil {
					return err
				}
				cmd.Println("Set", configKey, "=", value)
				return nil
			case "delete":
				if err := DeleteAll(); err != nil {
					return err
				}
				fmt.Println("Config deleted successfully")
				return nil
			case "token":
				configKey := configKeyPrefix + ConfigTokenKey
				if len(args) == 1 {
					// Get token
					tok, err := Get(configKey)
					if err != nil {
						return fmt.Errorf("%s not found", configKey)
					}
					fmt.Println(configKey, "=", tok)
					return nil
				}
				// Set token
				value := strings.TrimSpace(args[1])
				if !strings.HasPrefix(value, "sk_project_") {
					return fmt.Errorf("invalid token: %s. It should start with 'sk_project_'", value)
				}
				if err := Set(configKey, value); err != nil {
					return err
				}
				fmt.Println("Set", configKey, "=", value)
				return nil
			case "endpoint":
				configKey := configKeyPrefix + ConfigEndpointKey
				if len(args) == 1 {
					// Get endpoint
					endpoint, err := Get(configKey)
					if err != nil {
						return fmt.Errorf("%s not found", configKey)
					}
					fmt.Println(configKey, "=", endpoint)
					return nil
				}
				// Set endpoint
				value := strings.TrimSpace(args[1])
				if err := Set(configKey, value); err != nil {
					return err
				}
				fmt.Println("Set", configKey, "=", value)
				return nil
			default:
				return fmt.Errorf("invalid configuration key '%s'. Available keys: token, endpoint, storage-id, delete", key)
			}
		},
	}

	cmd.Flags().StringVar(&profile, "profile", "", "Use a specific profile. When not set, the default profile is used.")

	return cmd
}
