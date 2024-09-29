package sources

import (
	"fmt"

	"github.com/TheZoraiz/ascii-image-converter/aic_package"
	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/styles"
	"github.com/spf13/cobra"
)

type GetCmd struct {
	Id   string `json:"id"`
	Data api.Source
}

func (r *GetCmd) Execute() error {
	apiReq := api.Request{
		Config: cmd.Config,
		Path:   fmt.Sprintf("/api/sources/%s", r.Id),
		Method: "GET",
	}

	if cmd.Config.Debug {
		fmt.Println(styles.Debug.Render(apiReq.String()))
	}

	source, err := api.ApiRequest[api.Source](apiReq)
	if err != nil {
		return err
	}

	r.Data = source

	return nil
}

func (r *GetCmd) View() {
	sourceList := ListCmd{Data: []api.Source{r.Data}}
	sourceList.View()

	if !cmd.Config.JSON && len(r.Data.Images) == 1 {
		asciiImage(r.Data)
	}
}

func newGetCmd() *cobra.Command {
	req := GetCmd{}

	cmd := &cobra.Command{
		Use:   "get source-id",
		Short: "get info about a source",
		Long:  "get info about a source",
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

	return cmd
}

func asciiImage(source api.Source) {
	flags := aic_package.DefaultFlags()
	flags.Width = 80
	flags.Colored = true
	flags.CustomMap = " .-=+#@"

	asciiArt, err := aic_package.Convert(source.Images[0], flags)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%v\n", asciiArt)
}
