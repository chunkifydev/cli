package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	_ "embed"

	"github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/chunkify-go/option"
	chunkifyCmd "github.com/chunkifydev/cli/pkg/chunkify"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/version"
	"github.com/chunkifydev/cli/pkg/webhook"
	"github.com/spf13/cobra"
)

// ChunkifyApiEndpoint is the default API endpoint URL for Chunkify
const (
	ChunkifyApiEndpoint = "https://api.chunkify.dev/v1"
)

//go:embed chunkify.txt
var chunkifyBanner string

// cfg holds the global configuration for the CLI defined in config pkg
var cfg = &config.Config{}

// Commander defines the interface for command execution and view generation
type Commander interface {
	execute() error
	view() string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.PersistentPreRun = initChunkifyClient

	// Check for updates after each command
	// TODO: check updates less often
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "update" {
			return
		}
		upToDate, latestVersion := version.IsUpToDate()
		if !upToDate {
			fmt.Println("  ────────────────────────────────────────────────")
			fmt.Println("  A new version of Chunkify CLI is available:", latestVersion)
			fmt.Println("  Run `chunkify update` to update to the latest version.")
		}
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// initChunkifyClient verifies authentication tokens and initializes the Chunkify client.
func initChunkifyClient(cmd *cobra.Command, args []string) {
	// All commands require project token, except config
	if cmd.Name() != "config" {
		if cfg.Token == "" {
			if err := cfg.SetToken(); err != nil {
				fmt.Printf("Authentication issue\n\n")
				if cfg.Profile != "" {
					fmt.Printf("Profile '%s' doesn't exist.\nRun `chunkify config token <sk_project_token> --profile %s` to link a project token to it.\n", cfg.Profile, cfg.Profile)
				} else {
					fmt.Printf("You haven't set your project token yet.\nRun `chunkify config token <sk_project_token>`\n")
				}
				os.Exit(1)
			}
		}
	}

	// Setting endpoint
	endpoint := ChunkifyApiEndpoint

	if os.Getenv("CHUNKIFY_ENDPOINT") != "" {
		endpoint = os.Getenv("CHUNKIFY_ENDPOINT")
	} else {
		apiEndpoint, err := config.Get(cfg.ConfigKey("config.endpoint"))
		if err == nil && apiEndpoint != "" {
			endpoint = apiEndpoint
		}
	}

	// Initialize client with available tokens
	client := chunkify.NewClient(
		option.WithProjectAccessToken(cfg.Token),
		option.WithBaseURL(endpoint),
	)

	cfg.Client = &client
}

// init initializes the CLI by setting up configuration and registering all available commands
func init() {
	if !slices.Contains(os.Args, "--json") {
		chunkifyBanner = strings.Replace(chunkifyBanner, "{version}", version.Version, 1)
		fmt.Println("\n" + chunkifyBanner + "\n")
	}

	rootCmd = chunkifyCmd.NewCommand(cfg).Command
	rootCmd.AddCommand(webhook.NewCommand(cfg).Command)
	rootCmd.AddCommand(VersionCmd)
	rootCmd.AddCommand(CliUpdateCmd)
	rootCmd.AddCommand(config.NewCommand())

	rootCmd.PersistentFlags().StringVar(&cfg.Profile, "profile", "", "Use a specific profile. When not set, the default profile is used. See config command for more details.")
}
