package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "embed"

	"github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/chunkify-go/option"
	"github.com/chunkifydev/cli/pkg/api"
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
	// API commands and JSON output report errors as JSON on stderr.
	jsonErrors := isAPIInvocation(os.Args) || jsonRequested(os.Args)
	if jsonErrors {
		rootCmd.SilenceErrors = true
		rootCmd.SilenceUsage = true
	}
	rootCmd.PersistentPreRun = initChunkifyClient

	// Check for updates after each command
	// TODO: check updates less often
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "update" || cmd.Name() == "version" || isAPICommand(cmd) || jsonOutput(cmd) {
			return
		}
		upToDate, latestVersion := version.IsUpToDate()
		if !upToDate {
			fmt.Println("  ────────────────────────────────────────────────")
			fmt.Println("  A new version of Chunkify CLI is available:", latestVersion)
			fmt.Println(" ", version.UpdateInstructions())
		}
	}

	if err := rootCmd.Execute(); err != nil {
		if jsonErrors {
			writeJSONError(err)
		}
		os.Exit(exitCode(err))
	}
}

// initChunkifyClient verifies authentication tokens and initializes the Chunkify client.
func initChunkifyClient(cmd *cobra.Command, args []string) {
	if cmd.Name() == "version" || cmd.Name() == "update" {
		return
	}
	if isAPICommand(cmd) {
		for i := 0; i < len(args); i++ {
			if args[i] == "--profile" && i+1 < len(args) {
				cfg.Profile = args[i+1]
				i++
			} else if strings.HasPrefix(args[i], "--profile=") {
				cfg.Profile = strings.TrimPrefix(args[i], "--profile=")
			}
		}
	}

	// Regular commands require a project token. API actions load their own token.
	if cmd.Name() != "config" && !isAPICommand(cmd) {
		if cfg.Token == "" {
			if err := cfg.SetToken(); err != nil {
				help := "You haven't set your project token yet.\nRun `chunkify config token <sk_project_token>` or set CHUNKIFY_TOKEN."
				if cfg.Profile != "" {
					help = fmt.Sprintf("Profile '%s' doesn't exist.\nRun `chunkify config token <sk_project_token> --profile %s` to link a project token to it.", cfg.Profile, cfg.Profile)
				}
				// Keep stdout free for command output, such as JSON progress.
				if jsonRequested(os.Args) {
					writeJSONError(fmt.Errorf("authentication issue: %s", strings.ReplaceAll(help, "\n", " ")))
				} else {
					fmt.Fprintf(os.Stderr, "Authentication issue\n\n%s\n", help)
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
	if !jsonRequested(os.Args) && !isAPIInvocation(os.Args) {
		chunkifyBanner = strings.Replace(chunkifyBanner, "{version}", version.Version, 1)
		fmt.Println("\n" + chunkifyBanner + "\n")
	}

	rootCmd = chunkifyCmd.NewCommand(cfg).Command
	rootCmd.AddCommand(webhook.NewCommand(cfg).Command)
	rootCmd.AddCommand(VersionCmd)
	rootCmd.AddCommand(CliUpdateCmd)
	rootCmd.AddCommand(config.NewCommand())
	rootCmd.AddCommand(api.NewCommand(cfg))

	rootCmd.PersistentFlags().StringVar(&cfg.Profile, "profile", "", "Use a specific profile. When not set, the default profile is used. See config command for more details.")
}

// exitCode returns 130, like a shell after Ctrl+C, when the user cancelled,
// and 1 for other errors.
func exitCode(err error) int {
	if errors.Is(err, chunkifyCmd.ErrCancelled) {
		return 130
	}
	return 1
}

// jsonOutput reports whether the command prints JSON progress, so nothing else may go to stdout.
func jsonOutput(cmd *cobra.Command) bool {
	enabled, err := cmd.Flags().GetBool("json")
	return err == nil && enabled
}

// jsonRequested reports whether the arguments turn on --json. It reads them
// before cobra parses flags, for output that happens earlier, like the banner.
func jsonRequested(args []string) bool {
	enabled := false
	for _, arg := range args {
		if arg == "--json" {
			enabled = true
		} else if value, ok := strings.CutPrefix(arg, "--json="); ok {
			enabled, _ = strconv.ParseBool(value)
		}
	}
	return enabled
}

// writeJSONError reports err on stderr as {"error": "..."}.
func writeJSONError(err error) {
	_ = json.NewEncoder(os.Stderr).Encode(map[string]string{"error": err.Error()})
}

func isAPICommand(cmd *cobra.Command) bool {
	for cmd != nil {
		if cmd.Name() == "api" {
			return true
		}
		cmd = cmd.Parent()
	}
	return false
}

func isAPIInvocation(args []string) bool {
	for i := 1; i < len(args); i++ {
		if args[i] == "--profile" {
			i++
			continue
		}
		if strings.HasPrefix(args[i], "--profile=") {
			continue
		}
		return args[i] == "api"
	}
	return false
}
