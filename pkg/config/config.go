package config

import (
	"fmt"
	"os"
	"strings"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

// Config holds configuration settings for the CLI including API endpoint,
// authentication tokens, client instance to use the library and output format preferences
type Config struct {
	Endpoint string
	Token    string
	Client   *chunkify.Client
}

// KeyringServiceKey is the service name used for storing secrets in the system keyring
const (
	KeyringServiceKey = "chunkify-cli"
	ConfigEndpointKey = "config.endpoint"
	ConfigTokenKey    = "config.token"
)

// SetToken attempts to set the project token from environment variables first,
// falling back to the keyring if not found in environment.
func (cfg *Config) SetToken() error {
	var err error
	cfg.Token = os.Getenv("CHUNKIFY_TOKEN")
	if cfg.Token != "" {
		return nil
	}

	tok, err := Get(ConfigTokenKey)
	if err == nil && tok != "" {
		cfg.Token = tok
		return nil
	}

	return fmt.Errorf("you need to authenticate first by setting your project token.\nRun `chunkify config token <sk_project_token>`")
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
	cmd := &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Manage configuration settings",
		Long: `Manage configuration settings for chunkify.

Available configuration keys:
  token     - Chunkify project token
  endpoint  - Chunkify API endpoint URL
  delete    - Delete config

Examples:
  chunkify config token                    # Get project token
  chunkify config token sk_project_token   # Set token to sk_project_token
  chunkify config delete                   # Delete config
`,
		Args: cobra.RangeArgs(1, 2),

		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			switch key {
			case "delete":
				if err := DeleteAll(); err != nil {
					return err
				}
				fmt.Println("Config deleted successfully")
				return nil
			case "token":
				if len(args) == 1 {
					// Get token
					tok, err := Get(ConfigTokenKey)
					if err != nil {
						return fmt.Errorf("%s not found", ConfigTokenKey)
					}
					fmt.Println(tok)
					return nil
				}
				// Set token
				value := args[1]
				fmt.Println("setting", ConfigTokenKey, "to", value)
				if err := Set(ConfigTokenKey, value); err != nil {
					return err
				}
				fmt.Println(ConfigTokenKey, "set to", value)
				return nil
			case "endpoint":
				if len(args) == 1 {
					// Get endpoint
					endpoint, err := Get(ConfigEndpointKey)
					if err != nil {
						return fmt.Errorf("%s not found", ConfigEndpointKey)
					}
					fmt.Println(endpoint)
					return nil
				}
				// Set endpoint
				value := args[1]
				if err := Set(ConfigEndpointKey, value); err != nil {
					return err
				}
				fmt.Println(ConfigEndpointKey, "set to", value)
				return nil
			default:
				return fmt.Errorf("invalid configuration key '%s'. Available keys: token, endpoint", key)
			}
		},
	}

	return cmd
}
