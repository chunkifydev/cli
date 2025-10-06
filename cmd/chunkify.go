package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	_ "embed"

	"github.com/chunkifydev/chunkify-go"
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
var cfg = &config.Config{Endpoint: ChunkifyApiEndpoint}

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
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	// Initialize client with available tokens
	client := chunkify.NewClientWithConfig(chunkify.Config{
		AccessTokens: chunkify.AccessTokens{
			ProjectToken: cfg.Token,
		},
		BaseURL: cfg.Endpoint,
	})

	cfg.Client = &client
}

// init initializes the CLI by setting up configuration and registering all available commands
func init() {
	if !slices.Contains(os.Args, "--json") {
		chunkifyBanner = strings.Replace(chunkifyBanner, "{version}", version.Version, 1)
		fmt.Println("\n" + chunkifyBanner + "\n")
	}

	if os.Getenv("CHUNKIFY_ENDPOINT") != "" {
		cfg.Endpoint = os.Getenv("CHUNKIFY_ENDPOINT")
	} else {
		apiEndpoint, err := config.Get("config.endpoint")
		if err == nil && apiEndpoint != "" {
			cfg.Endpoint = apiEndpoint
		}
	}

	rootCmd = chunkifyCmd.NewCommand(cfg).Command
	rootCmd.AddCommand(webhook.NewCommand(cfg).Command)
	rootCmd.AddCommand(VersionCmd)
	rootCmd.AddCommand(config.NewCommand())
}
