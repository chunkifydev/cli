package cmd

import (
	"os"

	"github.com/level63/cli/pkg/config"
	"github.com/level63/cli/pkg/functions"
	"github.com/level63/cli/pkg/jobs"
	"github.com/level63/cli/pkg/logs"
	"github.com/level63/cli/pkg/notifications"
	"github.com/level63/cli/pkg/projects"
	"github.com/level63/cli/pkg/sources"
	"github.com/level63/cli/pkg/storages"
	"github.com/level63/cli/pkg/tokens"
	"github.com/level63/cli/pkg/webhooks"
	"github.com/spf13/cobra"
)

var cfg = &config.Config{}

type Commander interface {
	execute() error
	view() string
}

var rootCmd = &cobra.Command{
	Use:   "level63",
	Short: "level63 is a command line interface for Level63 API",
	Long:  `level63 is a command line interface for Level63 API.`,
}

func Execute() {
	rootCmd.PersistentPreRun = checkAccountSetup
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func checkAccountSetup(cmd *cobra.Command, args []string) {
	if cfg.AccountToken == "" && cmd.Name() != "auth" && cmd.Name() != "login" {
		err := cfg.SetDefaultTokens()
		if err != nil {
			printError(err)
			os.Exit(1)
		}
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&cfg.JSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "Print debug info")
	rootCmd.PersistentFlags().StringVar(&cfg.ApiEndpoint, "endpoint", "https://api.level63-staging.dev/v1", "The API endpoint")
	rootCmd.PersistentFlags().StringVar(&cfg.DefaultProjectId, "env-project-id", "", "Select the project and run the command")

	rootCmd.AddCommand(storages.NewCommand(cfg).Command)
	rootCmd.AddCommand(projects.NewCommand(cfg).Command)
	rootCmd.AddCommand(sources.NewCommand(cfg).Command)
	rootCmd.AddCommand(jobs.NewCommand(cfg).Command)
	rootCmd.AddCommand(logs.NewCommand(cfg).Command)
	rootCmd.AddCommand(webhooks.NewCommand(cfg).Command)
	rootCmd.AddCommand(functions.NewCommand(cfg).Command)
	rootCmd.AddCommand(notifications.NewCommand(cfg).Command)
	rootCmd.AddCommand(tokens.NewCommand(cfg).Command)
	rootCmd.AddCommand(newAuthCmd(cfg))
}
