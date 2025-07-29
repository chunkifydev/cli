package projects

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/tokens"
	"github.com/spf13/cobra"
)

// newSelectCmd creates and configures a new cobra command for selecting a default project
func newSelectCmd(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select [project-id]",
		Short: "Select a project by default for all future commands",
		Long:  `Select a project by default for all future commands`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			// Get list of all projects
			list := ListCmd{}
			if err := list.Execute(); err != nil {
				fmt.Println("Couldn't not retrieve your projects", err)
				os.Exit(1)
			}
			projects := list.Data

			// If no project ID provided, show interactive prompt
			if len(args) == 0 {
				SelectProjectPrompt(projects)
				return
			}

			// Checking the project id is valid
			for _, project := range projects {
				if args[0] == project.Id {
					selectProject(args[0])
					fmt.Printf("Project %s (%s) selected", project.Name, project.Id)
					return
				}
			}
			printError(fmt.Errorf("project id '%s' is not valid", args[0]))
		},
	}

	return cmd
}

// SelectProjectPrompt displays an interactive prompt for selecting a project
func SelectProjectPrompt(projects []chunkify.Project) {
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

	selectProject(projects[projectIndex].Id)
}

// selectProject sets the given project as default and ensures a valid token exists
func selectProject(projectId string) {
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
	tokenCreate := tokens.CreateCmd{Params: chunkify.TokenCreateParams{Name: "chunkify-cli", Scope: "project", ProjectId: &projectId}}
	if err := tokenCreate.Execute(); err != nil {
		printError(fmt.Errorf("couldn't create a token for this project: %s", err))
		os.Exit(1)
	}

	// Save the new token
	if err := config.SetToken(tokenCreate.Data.Id, projectId, tokenCreate.Data.Token); err != nil {
		printError(fmt.Errorf("couldn't save the token: %s", err))
		os.Exit(1)
	}

	// Set as default project
	if err := config.Set("DefaultProject", projectId); err != nil {
		printError(fmt.Errorf("couldn't set the project for all future commands: %s", err))
		os.Exit(1)
	}
}
