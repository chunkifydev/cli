package cmd

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ChunkifyAuthUrl is the base URL for the authentication endpoint
const ChunkifyAuthUrl = "https://chunkify.dev/auth/cli"

var (
	// authUrl stores the authentication URL, can be overridden via CHUNKIFY_AUTH_URL env var
	authUrl string
	// noBrowser flag determines if browser-based auth should be skipped
	noBrowser = false
	// teamToken stores the authentication token received from the server
	teamToken chunkify.Token
)

// newAuthCmd creates the root auth command that contains login/logout subcommands
func newAuthCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Connect to your Chunkify account",
		Long:  `Connect to your Chunkify account`,
	}

	cmd.AddCommand(newLoginCmd(cfg))
	cmd.AddCommand(newLogoutCmd(cfg))

	return cmd
}

// newLoginCmd creates the login subcommand that handles user authentication
func newLoginCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Chunkify",
		Long:  `Login to Chunkify`,
		Run: func(c *cobra.Command, args []string) {
			// client := chunkify.NewClientWithConfig(chunkify.Config{
			// 	BaseURL: cfg.ApiEndpoint,
			// })
			// cfg.Client = &client

			login(cfg)
		},
	}

	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't open the browser to authorize the cli")
	if os.Getenv("CHUNKIFY_AUTH_URL") != "" {
		authUrl = os.Getenv("CHUNKIFY_AUTH_URL")
	} else {
		authUrl = ChunkifyAuthUrl
	}
	return cmd
}

// newLogoutCmd creates the logout subcommand that handles user deauthentication
func newLogoutCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove all data and logout",
		Long:  `Remove all data and logout`,
		Run: func(c *cobra.Command, args []string) {
			logout(cfg)
		},
	}

	return cmd
}

// login handles the authentication flow, either via browser or manual token entry
func login(cfg *config.Config) {
	// Clear any existing login data
	logout(cfg)

	if noBrowser {
		// If no-browser flag is set, prompt for manual token entry
		fmt.Printf("Please enter your team token:\n")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			printError(err)
			os.Exit(1)
		}
		teamToken.Token = string(bytePassword)
	} else {
		// Otherwise, initiate browser-based authentication
		fmt.Println("Waiting...")
		startTime := time.Now()
		localUrl := startHttpServer()
		hostname, _ := os.Hostname()
		// Open browser with authentication URL
		openbrowser(fmt.Sprintf("%s?url=%s&name=%s", authUrl, localUrl, hostname))
		t := time.NewTicker(1 * time.Second)

		// Wait for authentication to complete or timeout after 5 minutes
		for range t.C {
			if teamToken.Token != "" {
				t.Stop()
				break
			}
			if time.Since(startTime) > 5*time.Minute {
				fmt.Println("Login expired, please restart the process")
				t.Stop()
				os.Exit(1)
			}
		}
	}

	cfg.TeamToken = teamToken.Token

	// Initialize client with the received team token
	client := chunkify.NewClientWithConfig(chunkify.Config{
		AccessTokens: chunkify.AccessTokens{
			TeamToken: cfg.TeamToken,
		},
		BaseURL: cfg.ApiEndpoint,
	})
	cfg.Client = &client

	// Verify the team token by fetching projects
	fmt.Println("Checking your account...")
	projects, err := client.ProjectList()
	if err != nil {
		printError(fmt.Errorf("team token is invalid"))
		os.Exit(1)
	}

	// Save the team token to configuration
	if err := config.SetToken(teamToken.Id, "TeamToken", teamToken.Token); err != nil {
		printError(err)
		os.Exit(1)
	}

	// Prompt user to select a project
	SelectProjectPrompt(cfg, projects)

	fmt.Println()
	fmt.Println("All good. Run `chunkify help` for help")
}

// getAllProjects fetches all projects accessible to the authenticated user
func getAllProjects(config *config.Config) ([]chunkify.Project, error) {
	projects, err := config.Client.ProjectList()
	if err != nil {
		return []chunkify.Project{}, err
	}

	return projects, nil
}

// SelectProjectPrompt displays an interactive prompt for selecting a project
func SelectProjectPrompt(cfg *config.Config, projects []chunkify.Project) {
	// Handle case when no projects exist
	if len(projects) == 0 {
		fmt.Println("You don't have any project. You can create a new project by running `chunkify projects create --name 'Project name'`")
		return
	}

	// Display numbered list of projects
	fmt.Println("What project do you want to use?")
	for i, project := range projects {
		fmt.Printf("%d. %s\n", i+1, project.Name)
	}

	// Get user input for project selection
	fmt.Printf("Please enter the project number: ")
	reader := bufio.NewReader(os.Stdin)
	projectNumberStr, _ := reader.ReadString('\n')
	projectNumber, err := strconv.Atoi(strings.TrimSpace(projectNumberStr))
	if err != nil {
		printError(fmt.Errorf("project number is not valid number %s", err))
		os.Exit(1)
	}

	// Validate project number is within range
	projectIndex := projectNumber - 1
	if projectIndex < 0 || projectIndex > len(projects) {
		printError(fmt.Errorf("project number is not valid"))
		os.Exit(1)
	}

	selectProject(cfg, projects[projectIndex].Id)
}

// selectProject sets the given project as default and ensures a valid token exists
func selectProject(cfg *config.Config, projectId string) {
	// Check if we already have a token for this project
	_, err := config.Get("project:" + projectId)
	if err == nil {
		// If token exists, just set as default project
		if err := config.Set("DefaultProject", projectId); err != nil {
			printError(fmt.Errorf("couldn't set the project for all future commands: %s", err))
			os.Exit(1)
		}
		return
	}

	// We don't have the token saved
	// So we generate a token to use this project
	tok, err := cfg.Client.TokenCreate(chunkify.TokenCreateParams{Name: "chunkify-cli", Scope: "project", ProjectId: &projectId})
	if err != nil {
		printError(fmt.Errorf("couldn't create a token for this project: %s", err))
		os.Exit(1)
	}

	// Save the new token
	if err := config.SetToken(tok.Id, projectId, tok.Token); err != nil {
		printError(fmt.Errorf("couldn't save the token: %s", err))
		os.Exit(1)
	}

	// Set as default project
	if err := config.Set("DefaultProject", projectId); err != nil {
		printError(fmt.Errorf("couldn't set the project for all future commands: %s", err))
		os.Exit(1)
	}
}

// revokeToken revokes a specific token from the server
func revokeToken(cfg *config.Config, tokenId string) error {
	err := cfg.Client.TokenRevoke(tokenId)
	if err != nil {
		return err
	}

	return nil
}

// logout handles the deauthentication process by revoking all tokens and clearing local config
func logout(cfg *config.Config) {
	// Revoke all project tokens
	projects, err := getAllProjects(cfg)
	if err == nil {
		for _, project := range projects {
			tokId, _, _ := config.GetToken(project.Id)
			if tokId != "" {
				fmt.Println("Revoke project token", project.Name)
				revokeToken(cfg, tokId)
			}
		}
	}

	// Revoke account token
	tokId, _, _ := config.GetToken("TeamToken")
	if tokId != "" {
		fmt.Println("Revoke team token")
		revokeToken(cfg, tokId)
	}

	if err := config.DeleteAll(); err != nil {
		printError(err)
		return
	}
}

// printError formats and prints error messages
func printError(err error) {
	fmt.Println(err.Error())
}

// authHandler processes the authentication callback from the server
func authHandler(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("token")
	token := strings.Split(tok, ":")
	teamToken.Id = token[0]
	teamToken.Token = token[1]

	http.Redirect(w, r, "/auth/ok", http.StatusFound)
}

// authOkHandler displays success message after successful authentication
func authOkHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Connection established. You can close this window.")
}

// startHttpServer starts a local HTTP server to handle authentication callback
func startHttpServer() string {
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/auth/ok", authOkHandler)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	go func() {
		if err := http.Serve(listener, nil); err != nil {
			panic(err)
		}
	}()

	return fmt.Sprintf("http://localhost:%d/auth", port)
}

// openbrowser opens the system browser based on the operating system
func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Println("Cannot open the browser, try with --no-browser")
		os.Exit(1)
	}
}
