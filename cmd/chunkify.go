package cmd

import (
	"fmt"
	"os"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/flags"

	"github.com/chunkifydev/cli/pkg/notifications"
	"github.com/spf13/cobra"

	"github.com/chunkifydev/chunkify-go"
)

// ChunkifyApiEndpoint is the default API endpoint URL for Chunkify
const (
	ChunkifyApiEndpoint = "https://api.chunkify.dev/v1"
)

// cfg holds the global configuration for the CLI defined in config pkg
var cfg = &config.Config{ApiEndpoint: ChunkifyApiEndpoint}

// someParam is an example flag bound to the root command
var someParam int64

// Commander defines the interface for command execution and view generation
type Commander interface {
	execute() error
	view() string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chunkify",
	Short: "chunkify is a command line interface for Chunkify API",
	Long:  `chunkify is a command line interface for Chunkify API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Place root command logic here. For demonstration, this simply
		// acknowledges the flag value when provided.
		if cmd.Flags().Changed("some-param") {
			fmt.Println("some-param:", someParam)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.PersistentPreRun = checkAccountSetup
	//if len(os.Args) == 1 {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	//}
}

// checkAccountSetup verifies and sets up authentication tokens based on the command being executed.
// It handles different authentication requirements for various command types:
// - auth commands may work without tokens
// - projects and tokens commands require team token
// - other commands require project token
func checkAccountSetup(cmd *cobra.Command, args []string) {
	// Early return if no parent command (shouldn't happen, but safer)
	if cmd.Parent() == nil {
		return
	}

	// Determine command type and handle accordingly
	switch cmd.Parent().Name() {
	case "auth":
		// For auth, try to set team token (i.e if it not the first login)
		if cfg.TeamToken == "" {
			if cfg.TeamToken == "" {
				// Try to set team token, but don't fail if it doesn't exist
				cfg.SetDefaultTeamToken() // Ignore error since it's optional for auth
			}
		}

	case "projects", "tokens":
		// Projects and tokens commands require team token
		if cfg.TeamToken == "" {
			if err := cfg.SetDefaultTeamToken(); err != nil {
				fmt.Println("error setting team token:", err)
				printError(err)
				os.Exit(1)
			}
		}

	default:
		// All other commands require project token
		if cfg.ProjectToken == "" {
			if err := cfg.SetDefaultProjectToken(); err != nil {
				printError(err)
				os.Exit(1)
			}
		}
	}

	// Initialize client with available tokens
	client := chunkify.NewClientWithConfig(chunkify.Config{
		AccessTokens: chunkify.AccessTokens{
			TeamToken:    cfg.TeamToken,
			ProjectToken: cfg.ProjectToken,
		},
		BaseURL: cfg.ApiEndpoint,
	})

	cfg.Client = &client
}

// init initializes the CLI by setting up configuration and registering all available commands
func init() {
	if os.Getenv("CHUNKIFY_API_ENDPOINT") != "" {
		cfg.ApiEndpoint = os.Getenv("CHUNKIFY_API_ENDPOINT")
	}
	hostname := ""
	flags.StringVar(rootCmd.Flags(), &hostname, "hostname", "", "Use the given hostname for the localdev webhook. If not provided, we use the hostname of the machine. It's purely visual, it will just appear on Chunkify")
	rootCmd.PersistentFlags().BoolVar(&cfg.JSON, "json", false, "Output in JSON format")

	// Root-level flag to support running without a subcommand
	flags.Int64Var(rootCmd.Flags(), &someParam, "some-param", 0, "Example parameter for root command execution")

	rootCmd.AddCommand(notifications.NewCommand(cfg).Command)
	rootCmd.AddCommand(newAuthCmd(cfg))
	rootCmd.AddCommand(VersionCmd)

}
