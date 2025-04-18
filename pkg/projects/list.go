package projects

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

type ListCmd struct {
	Data []chunkify.Project
}

func (r *ListCmd) Execute() error {
	projects, err := cmd.Config.Client.ProjectList()
	if err != nil {
		return err
	}

	r.Data = projects

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
		fmt.Println(styles.DefaultText.Render("No project found."))
		return
	}

	fmt.Println(r.projectsTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

func (r *ListCmd) projectsTable() *table.Table {
	rightCols := []int{}
	centerCols := []int{4}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "Project Id", "Name", "Storage").
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
		Rows(projectsListToRows(r.Data)...)

	return table
}

func projectsListToRows(projects []chunkify.Project) [][]string {
	rows := make([][]string, len(projects))
	for i, project := range projects {
		rows[i] = []string{
			project.CreatedAt.Format(time.RFC822),
			styles.Id.Render(project.Id),
			project.Name,
			project.Storage,
		}
	}
	return rows
}

func newListCmd() *cobra.Command {
	req := ListCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all projects",
		Long:  `list all projects`,
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
