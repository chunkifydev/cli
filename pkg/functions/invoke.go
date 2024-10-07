package functions

import (
	"fmt"

	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type InvokeCmd struct {
	Id   string `json:"-"`
	Body string `json:"-"`

	Data api.FunctionInvoked `json:"-"`
}

func (r *InvokeCmd) Execute() error {
	resp, err := api.ApiRequest[api.FunctionInvoked](
		api.Request{
			Config: cmd.Config,
			Path:   fmt.Sprintf("/api/functions/%s/invoke", r.Id),
			Method: "POST",
			Body:   r.Body,
		})
	if err != nil {
		return err
	}

	r.Data = resp

	return nil
}

func (r *InvokeCmd) View() {
	if cmd.Config.Debug {
		return
	}

	fmt.Println(r.Data.Body)
}

func newInvokeCmd() *cobra.Command {
	req := InvokeCmd{}

	cmd := &cobra.Command{
		Use:   "invoke function-id",
		Short: "invoke a function",
		Long:  `invoke a function`,
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

	cmd.Flags().StringVar(&req.Body, "body", "", "Invoke the function with the given body")

	return cmd
}
