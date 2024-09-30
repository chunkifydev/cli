package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/config"
	"github.com/level63/cli/pkg/styles"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

func checkAccountSetup(cmd *cobra.Command, args []string) {
	if cfg.AccountApiKey == "" && cmd.Name() != "setup" {
		accountApiKey, err := keyring.Get(serviceKey, "AccountApiKey")
		if err != nil {
			fmt.Printf("You must setup your account first.\nRun `level63 setup` to do so.\n")
			os.Exit(1)
		}

		projectApiKey, err := keyring.Get(serviceKey, "ProjectApiKey")
		if err != nil {
			fmt.Printf("You must setup your project first.\nRun `level63 setup` to do so.\n")
			os.Exit(1)
		}

		cfg.AccountApiKey = accountApiKey
		cfg.ProjectApiKey = projectApiKey
	}
}

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "setup your account and project",
		Long:  `setup your account and project`,
		Run: func(c *cobra.Command, args []string) {
			if c.Flag("delete").Changed {
				destroyAllKeys()
				os.Exit(0)
			}

			setupAccount()
		},
	}

	cmd.Flags().Bool("delete", false, "Delete all api keys on this system")

	return cmd
}

func setupAccount() {
	fmt.Printf("Please enter your account API key:\n")
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	// check the key is correct
	newCfg := &config.Config{AccountApiKey: string(bytePassword), ApiEndpoint: cfg.ApiEndpoint}
	projects, err := getAllProjects(newCfg)
	if err != nil {
		printError(fmt.Errorf("account API key is invalid"))
		os.Exit(1)
	}

	if err := keyring.Set(serviceKey, "AccountApiKey", string(bytePassword)); err != nil {
		printError(err)
		os.Exit(1)
	}

	setupProject(projects)
}

func setupProject(projects []api.Project) {
	if len(projects) == 0 {
		fmt.Println("You don't have any projects.\nPlease create a project first by running " + styles.DefaultText.Render("`level63 projects create`"))
		os.Exit(0)
	}

	fmt.Printf("Please enter the API key for the project you want to use:\n")

	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := keyring.Set(serviceKey, "ProjectApiKey", string(bytePassword)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("\n\nYou can now run " +
		styles.DefaultText.Render("`level63 jobs list`") + " to see your jobs.\n\n" +
		styles.Important.Render("Setup is complete"))
}

func getAllProjects(config *config.Config) ([]api.Project, error) {
	projects, err := api.ApiRequest[[]api.Project](api.Request{
		Config: config,
		Path:   "/api/projects",
		Method: "GET",
	})
	if err != nil {
		fmt.Println(err)
		return []api.Project{}, err
	}

	return projects, nil
}

func destroyAllKeys() {
	keyring.Delete(serviceKey, "AccountApiKey")
	keyring.Delete(serviceKey, "ProjectApiKey")
	fmt.Println("All keys deleted")
}

func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
