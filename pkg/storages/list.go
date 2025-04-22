package storages

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// ListCmd represents the command for listing storage configurations
type ListCmd struct {
	Data []chunkify.Storage // List of storages
}

// Execute retrieves the list of storage configurations
func (r *ListCmd) Execute() error {
	storages, err := cmd.Config.Client.StorageList()
	if err != nil {
		return err
	}

	r.Data = storages

	return nil
}

// View displays the list of storage configurations
// If JSON output is enabled, it prints the data in JSON format
// Otherwise, it displays the data in a formatted table
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

// storagesTable creates and configures a table for displaying storage information
func (r *ListCmd) storagesTable() *table.Table {
	rightCols := []int{}
	centerCols := []int{3, 4}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Storage ID", "Provider", "Bucket", "Region", "Private").
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

// storagesListToRows converts storage data into string rows for table display
func storagesListToRows(storages []chunkify.Storage) [][]string {
	rows := make([][]string, len(storages))
	for i, storage := range storages {
		rows[i] = []string{
			styles.Important.Render(storage.Id),
			storage.Provider,
			storage.Bucket,
			storage.Region,
			formatter.BoolDefaultColor(!storage.Public),
		}
	}
	return rows
}

// newListCmd creates and configures a new cobra command for listing storage configurations
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
