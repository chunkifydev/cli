package transcoders

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

// ListCmd represents the command for retrieving transcoder statuses information
type ListCmd struct {
	Id          string                      `json:"id"` // ID of the job to retrieve
	Data        []chunkify.TranscoderStatus // Response data containing job details
	interactive bool                        // Whether to run in interactive mode
}

// Execute retrieves the job information from the API
func (r *ListCmd) Execute() error {
	transcoders, err := cmd.Config.Client.JobListTranscoders(r.Id)
	if err != nil {
		return err
	}

	r.Data = transcoders
	return nil
}

func (r *ListCmd) View() {
	if !cmd.Config.JSON && r.interactive {
		StartPolling(r)
	} else {
		r.displayView()
	}
}

func (r *ListCmd) displayView() {
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
		fmt.Println(styles.DefaultText.Render("No Transcoders found."))
		return
	}

	fmt.Println(r.transcodersTable())

	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

func (r *ListCmd) transcodersTable() *table.Table {
	rightCols := []int{3, 6, 7, 8, 9}
	centerCols := []int{5}

	rows := transcoderStatusesListToRows(r.Data)

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "ID", "Instance ID", "Chunk#", "Status", "Progress", "Speed", "FPS", "Out_Time", "Frame").
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
		Rows(rows...)

	return table
}

func transcoderStatusesListToRows(transcoders []chunkify.TranscoderStatus) [][]string {
	rows := make([][]string, len(transcoders))
	for i, transcoder := range transcoders {
		rows[i] = []string{
			transcoder.CreatedAt.Format(time.RFC822),
			transcoder.Id,
			transcoder.TranscoderInstanceId,
			fmt.Sprintf("%d", transcoder.ChunkNumber),
			formatter.TranscoderStatus(transcoder.Status),
			fmt.Sprintf("%.f%%", transcoder.Progress),
			fmt.Sprintf("%.2fx", transcoder.Speed),
			fmt.Sprintf("%.f", transcoder.Fps),
			fmt.Sprintf("%d", transcoder.OutTime),
			fmt.Sprintf("%d", transcoder.Frame),
		}
	}
	return rows
}

// newGetCmd creates a new command for retrieving trans information
func newListCmd() *cobra.Command {
	req := ListCmd{}

	cmd := &cobra.Command{
		Use:   "list job-id",
		Short: "list transcoders statuses for a job",
		Long:  `list transcoders statuses for a job`,
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
