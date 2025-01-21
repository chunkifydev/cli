package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/chunkifydev/cli/pkg/api"
	"github.com/chunkifydev/cli/pkg/config"
	projectsCmd "github.com/chunkifydev/cli/pkg/projects"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authUrl string
var noBrowser = false
var accountToken api.Token

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

func newLoginCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Chunkify",
		Long:  `Login to Chunkify`,
		Run: func(c *cobra.Command, args []string) {
			login(cfg)
		},
	}

	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't open the browser to authorize the cli")
	cmd.Flags().StringVar(&authUrl, "auth-url", "https://app.chunkify.dev/auth/cli", "The auth URL endpoint")
	return cmd
}

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

func login(cfg *config.Config) {
	// Clear any existing login data
	logout(cfg)

	if noBrowser {
		// If no-browser flag is set, prompt for manual token entry
		fmt.Printf("Please enter your account token:\n")
		bytePassword, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			printError(err)
			os.Exit(1)
		}
		accountToken.Token = string(bytePassword)
	} else {
		// Otherwise, initiate browser-based authentication
		fmt.Println("Waiting...")
		startTime := time.Now()
		localUrl := startHttpServer()
		hostname, _ := os.Hostname()
		// Open browser with authentication URL
		openbrowser(fmt.Sprintf("%s?url=%s&name=%s", authUrl, localUrl, hostname))
		t := time.NewTicker(1 * time.Second)

		// Wait for authentication to complete or timeout
		for range t.C {
			if accountToken.Token != "" {
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

	cfg.AccountToken = accountToken.Token

	// Verify the account token by fetching projects
	fmt.Println("Checking your account...")
	list := projectsCmd.ListCmd{}
	if err := list.Execute(); err != nil {
		printError(fmt.Errorf("account token is invalid"))
		os.Exit(1)
	}
	projects := list.Data

	// Save the account token to configuration
	if err := config.SetToken(accountToken.Id, "AccountToken", accountToken.Token); err != nil {
		printError(err)
		os.Exit(1)
	}

	// Prompt user to select a project
	projectsCmd.SelectProjectPrompt(projects)

	fmt.Println()
	fmt.Println("All good. Run `chunkify help` for help")
}

func getAllProjects(config *config.Config) ([]api.Project, error) {
	projects, err := api.ApiRequest[[]api.Project](api.Request{
		Config: config,
		Path:   "/api/projects",
		Method: "GET",
	})
	if err != nil {
		return []api.Project{}, err
	}

	return projects, nil
}

func revokeToken(config *config.Config, tokenId string) error {
	_, err := api.ApiRequest[api.EmptyResponse](api.Request{
		Config: config,
		Path:   "/api/tokens/" + tokenId,
		Method: "DELETE",
	})
	if err != nil {
		return err
	}

	return nil
}

func logout(cfg *config.Config) {
	// revoke all project tokens
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

	// revoke account token
	tokId, _, _ := config.GetToken("AccountToken")
	if tokId != "" {
		fmt.Println("Revoke account token")
		revokeToken(cfg, tokId)
	}

	if err := config.DeleteAll(); err != nil {
		printError(err)
		return
	}
}

func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("token")
	token := strings.Split(tok, ":")
	fmt.Printf("token: %#+v\n", token)
	accountToken.Id = token[0]
	accountToken.Token = token[1]

	http.Redirect(w, r, "/auth/ok", http.StatusFound)
}

func authOkHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Connection established. You can close this window.")
}

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
