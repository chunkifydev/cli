package jobs

import (
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type RestartCmd struct {
	Id          string
	interactive bool
	Data        api.Job
}

func (r *RestartCmd) Execute() error {
	job, err := api.ApiRequest[api.Job](
		api.Request{
			Config: cmd.Config,
			Path:   fmt.Sprintf("/api/jobs/%s/restart", r.Id),
			Method: "POST",
			Body:   map[string]string{},
		})
	if err != nil {
		return err
	}

	r.Data = job

	return nil
}

func (r *RestartCmd) View() {
	jobList := &ListCmd{CreatedSort: "asc", SourceId: r.Data.SourceId, interactive: r.interactive}
	jobList.Execute()
	if !cmd.Config.JSON && r.interactive {
		StartPolling(jobList)
	} else {
		jobList.View()
	}
}

func newRestartCmd() *cobra.Command {
	req := RestartCmd{}

	cmd := &cobra.Command{
		Use:   "restart job-id",
		Short: "Restart a job when status is error",
		Long:  `Restart a job when status is error`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	cmd.Flags().BoolVarP(&req.interactive, "interactive", "i", false, "Refresh the list in real time")
	return cmd
}
