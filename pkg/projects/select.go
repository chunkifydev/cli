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

func newSelectCmd(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select [project-id]",
		Short: "Select a project by default for all future commands",
		Long:  `Select a project by default for all future commands`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			list := ListCmd{}
			if err := list.Execute(); err != nil {
				fmt.Println("Couldn't not retrieve your projects", err)
				os.Exit(1)
			}
			projects := list.Data

			if len(args) == 0 {
				SelectProjectPrompt(projects)
				return
			}

			// checking the project id is valid
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

func SelectProjectPrompt(projects []chunkify.Project) {
	if len(projects) == 0 {
		fmt.Println("You don't have any project. You can create a new project by running `chunkify projects create --name 'Project name'`")
		return
	}

	fmt.Println("What project do you want to use?")
	for i, project := range projects {
		fmt.Printf("%d. %s\n", i+1, project.Name)
	}

	fmt.Printf("Please enter the project number: ")
	reader := bufio.NewReader(os.Stdin)
	projectNumberStr, _ := reader.ReadString('\n')
	projectNumber, err := strconv.Atoi(strings.TrimSpace(projectNumberStr))
	if err != nil {
		printError(fmt.Errorf("project number is not valid number %s", err))
		os.Exit(1)
	}

	projectIndex := projectNumber - 1
	if projectIndex < 0 || projectIndex > len(projects) {
		printError(fmt.Errorf("project number is not valid"))
		os.Exit(1)
	}

	selectProject(projects[projectIndex].Id)
}

func selectProject(projectId string) {
	_, err := config.Get("project:" + projectId)
	if err == nil {
		if err := config.Set("DefaultProject", projectId); err != nil {
			printError(fmt.Errorf("couldn't set the project for all future commands: %s", err))
			os.Exit(1)
		}
		return
	}

	// we don't have the token saved
	// so we generate a token to use this project
	tokenCreate := tokens.CreateCmd{Params: chunkify.TokenCreateParams{Name: "chunkify-cli", Scope: "project", ProjectId: projectId}}
	if err := tokenCreate.Execute(); err != nil {
		printError(fmt.Errorf("couldn't create a token for this project: %s", err))
		os.Exit(1)
	}

	if err := config.SetToken(tokenCreate.Data.Id, projectId, tokenCreate.Data.Token); err != nil {
		printError(fmt.Errorf("couldn't save the token: %s", err))
		os.Exit(1)
	}

	if err := config.Set("DefaultProject", projectId); err != nil {
		printError(fmt.Errorf("couldn't set the project for all future commands: %s", err))
		os.Exit(1)
	}
}
