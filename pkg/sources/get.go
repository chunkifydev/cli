package sources

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// GetCmd represents the command for retrieving source information
type GetCmd struct {
	Id   string          `json:"id"` // ID of the source to retrieve
	Data chunkify.Source // The retrieved source data
}

// Execute retrieves the source information from the API
func (r *GetCmd) Execute() error {
	source, err := cmd.Config.Client.Source(r.Id)
	if err != nil {
		return err
	}

	r.Data = source

	return nil
}

// View displays the retrieved source information
func (r *GetCmd) View() {
	sourceList := ListCmd{Data: []chunkify.Source{r.Data}}
	sourceList.View()
}

// newGetCmd creates and configures a new cobra command for retrieving source information
func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get source-id",
		Short: "get info about a source",
		Long:  "get info about a source",
		Args:  cobra.ExactArgs(1), // Requires exactly one argument (source ID)
		Run: func(cmd *cobra.Command, args []string) {
			req.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	return cmd
}
