package cmd

import (
	"fmt"
	"os"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/jobs"
	"github.com/chunkifydev/cli/pkg/logs"
	"github.com/chunkifydev/cli/pkg/notifications"
	"github.com/chunkifydev/cli/pkg/projects"
	"github.com/chunkifydev/cli/pkg/sources"
	"github.com/chunkifydev/cli/pkg/storages"
	"github.com/chunkifydev/cli/pkg/tokens"
	"github.com/chunkifydev/cli/pkg/webhooks"
	"github.com/spf13/cobra"

	chunkify "github.com/chunkifydev/chunkify-go"
)

var cfg = &config.Config{}

type Commander interface {
	execute() error
	view() string
}

var rootCmd = &cobra.Command{
	Use:   "chunkify",
	Short: "chunkify is a command line interface for Chunkify API",
	Long:  `chunkify is a command line interface for Chunkify API.`,
}

func Execute() {
	rootCmd.PersistentPreRun = checkAccountSetup
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

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

func init() {
	rootCmd.PersistentFlags().BoolVar(&cfg.JSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "Print debug info")
	rootCmd.PersistentFlags().StringVar(&cfg.ApiEndpoint, "endpoint", "https://api.chunkify.ing/v1", "The API endpoint")
	rootCmd.PersistentFlags().StringVar(&cfg.DefaultProjectId, "env-project-id", "", "Select the project and run the command")

	rootCmd.AddCommand(storages.NewCommand(cfg).Command)
	rootCmd.AddCommand(projects.NewCommand(cfg).Command)
	rootCmd.AddCommand(sources.NewCommand(cfg).Command)
	rootCmd.AddCommand(jobs.NewCommand(cfg).Command)
	rootCmd.AddCommand(logs.NewCommand(cfg).Command)
	rootCmd.AddCommand(webhooks.NewCommand(cfg).Command)
	rootCmd.AddCommand(notifications.NewCommand(cfg).Command)
	rootCmd.AddCommand(tokens.NewCommand(cfg).Command)
	rootCmd.AddCommand(newAuthCmd(cfg))
}
