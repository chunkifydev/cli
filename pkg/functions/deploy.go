package functions

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type DeployCmd struct {
	Script       string            `json:"script"`
	Name         string            `json:"name"`
	Enabled      bool              `json:"enabled"`
	Events       string            `json:"events,omitempty"`
	Environments map[string]string `json:"environments"`
	env          []string          `json:"-"`
	envFile      string            `json:"-"`
	Data         api.Function      `json:"-"`
}

func (r *DeployCmd) Execute() error {
	function, err := api.ApiRequest[api.Function](api.Request{Config: cmd.Config, Path: "/api/functions", Method: "POST", Body: r})
	if err != nil {
		return err
	}

	r.Data = function

	return nil
}

func (r *DeployCmd) View() {
	functionsList := ListCmd{Data: []api.Function{r.Data}}
	functionsList.View()
}

func fileToBase64(path string) (string, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(f), nil
}

func envFileToMap(path string) (map[string]string, error) {
	envVars := map[string]string{}

	f, err := os.ReadFile(path)
	if err != nil {
		return envVars, err
	}

	content := string(f)
	content = strings.TrimSpace(content)

	for _, line := range strings.Split(content, "\n") {
		v := strings.Split(line, "=")
		if len(v) == 2 {
			envVars[strings.TrimSpace(v[0])] = strings.TrimSpace(v[1])
		}
	}

	return envVars, nil
}

func newCreateCmd() *cobra.Command {
	req := DeployCmd{Enabled: true, Environments: map[string]string{}}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new function for your current project",
		Long:  `Deploy a new function for your current project`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			base64Script, err := fileToBase64(args[0])
			if err != nil {
				printError(fmt.Errorf("script file '%s' not valid: %s", args[0], err.Error()))
				return
			}
			req.Script = base64Script

			if req.Name == "" {
				req.Name = path.Base(args[0])
			}

			if req.envFile != "" {
				envVars, err := envFileToMap(req.envFile)
				if err != nil {
					printError(fmt.Errorf("env file '%s' not valid: %s", req.envFile, err.Error()))
					return
				}

				req.Environments = envVars
			}

			for _, envVar := range req.env {
				v := strings.Split(envVar, "=")
				req.Environments[strings.TrimSpace(v[0])] = strings.TrimSpace(v[1])
			}

			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	cmd.Flags().StringVar(&req.envFile, "env-file", "", "Add Environment variables from a file")
	cmd.Flags().StringSliceVar(&req.env, "env", []string{}, "Set environment variables: VAR=value")
	cmd.Flags().StringVar(&req.Name, "name", "", "The function name. If not set, the filename is the name by default")
	cmd.Flags().StringVar(&req.Events, "events", "*", "Create a function that will trigger for specific events. *, job.* or job.completed")
	cmd.MarkFlagRequired("url")

	return cmd
}
