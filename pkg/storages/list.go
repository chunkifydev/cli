package storages

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/formatter"
	"github.com/level63/cli/pkg/styles"
	"github.com/spf13/cobra"
)

type ListCmd struct {
	Data []api.Storage
}

func (r *ListCmd) Execute() error {
	apiReq := api.Request{
		Config: cmd.Config,
		Path:   "/api/storages",
		Method: "GET",
	}

	storages, err := api.ApiRequest[[]api.Storage](apiReq)
	if err != nil {
		return err
	}

	r.Data = storages

	return nil
}

func (r *ListCmd) View() {
	if cmd.Config.JSON {
		dataBytes, err := json.MarshalIndent(r.Data, "", "  ")
		if err != nil {
			printError(err)
			return
		}
		fmt.Println(string(dataBytes))
		return
	}

	if len(r.Data) == 0 {
		fmt.Println(styles.DefaultText.Render("No storage found."))
		return
	}

	fmt.Println(r.storagesTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

func (r *ListCmd) storagesTable() *table.Table {
	rightCols := []int{}
	centerCols := []int{3, 4}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Storage Id", "Provider", "Bucket", "Region", "Private", "Test").
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				if slices.Contains(rightCols, col) {
					return styles.Right.Padding(0, 1).Foreground(styles.GrayColor)
				}
				if slices.Contains(centerCols, col) {
					return styles.Center.Padding(0, 1).Foreground(styles.GrayColor)
				}

				return styles.Header.Padding(0, 1)
			case slices.Contains(rightCols, col):
				return styles.Right.Padding(0, 1)
			case slices.Contains(centerCols, col):
				return styles.Center.Padding(0, 1)
			default:
				return styles.TableSpacing
			}
		}).
		Rows(storagesListToRows(r.Data)...)

	return table
}

func storagesListToRows(storages []api.Storage) [][]string {
	rows := make([][]string, len(storages))
	for i, storage := range storages {
		test := ""
		if storage.Reserved {
			test = "Ephemeral (24h)"
		}

		rows[i] = []string{
			styles.Important.Render(storage.Id),
			storage.Provider,
			storage.Bucket,
			storage.Region,
			formatter.BoolDefaultColor(!storage.Public),
			styles.DefaultText.Render(test),
		}
	}
	return rows
}

func newListCmd() *cobra.Command {
	req := ListCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all storages",
		Long:  `list all storages`,
		Run: func(_ *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	return cmd
}
