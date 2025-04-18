/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package jobs

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

type GetCmd struct {
	Id          string `json:"id"`
	Data        chunkify.Job
	interactive bool
}

func (r *GetCmd) Execute() error {
	job, err := cmd.Config.Client.Job(r.Id)
	if err != nil {
		return err
	}

	r.Data = job
	return nil
}

func (r *GetCmd) View() {
	jobList := &ListCmd{Params: chunkify.JobListParams{Id: r.Data.Id}, Data: []chunkify.Job{r.Data}, interactive: r.interactive}
	if !cmd.Config.JSON && r.interactive {
		StartPolling(jobList)
	} else {
		jobList.View()
	}
}

func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get job-id",
		Short: "get info about a job",
		Long:  `get info about a job`,
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
