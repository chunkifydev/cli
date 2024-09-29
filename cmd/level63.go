package cmd

import (
	"os"

	"github.com/level63/cli/pkg/config"
	"github.com/level63/cli/pkg/jobs"
	"github.com/level63/cli/pkg/sources"
	"github.com/spf13/cobra"
)

var cfg = config.Config{
	AccountApiKey: os.Getenv("LEVEL63_ACCOUNT_API_KEY"),
	ProjectApiKey: os.Getenv("LEVEL63_API_KEY"),
}

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
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&cfg.JSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "Print debug info")
	rootCmd.Flags().StringVar(&cfg.ApiEndpoint, "endpoint", "https://api-pr34.level63-staging.dev/pr34", "The API endpoint")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(sources.NewCommand(&cfg).Command)
	rootCmd.AddCommand(jobs.NewCommand(&cfg).Command)
}
