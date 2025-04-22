package jobs

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"

	chunkify "github.com/chunkifydev/chunkify-go"
)

// FilesListCmd represents the command for listing files associated with a job
type FilesListCmd struct {
	Id      string          `json:"id"` // ID of the job to list files for
	presign bool            // Whether to return presigned URLs for the files
	Data    []chunkify.File // List of files returned from the API
}

// Execute retrieves the list of files for the specified job ID
func (r *FilesListCmd) Execute() error {
	files, err := cmd.Config.Client.JobListFiles(r.Id)
	if err != nil {
		return err
	}

	r.Data = files

	return nil
}

// View displays the list of files in either JSON or table format
func (r *FilesListCmd) View() {
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
		fmt.Println(styles.DefaultText.Render("No file found."))
		return
	}

	if r.presign {
		for _, file := range r.Data {
			fmt.Printf("%s (%s):\n%s\n\n",
				styles.Important.Render(file.Path),
				formatter.Size(file.Size),
				styles.Debug.Render(file.Url),
			)
			if file.Url == "" {
				fmt.Println(styles.Error.Render("Presigned URL not found."))
			}
		}
		return
	}

	fmt.Println(r.filesTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

// filesTable creates and returns a formatted table of files
func (r *FilesListCmd) filesTable() *table.Table {
	rightCols := []int{4}
	centerCols := []int{1}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "Storage Id", "Path", "Mime-Type", "Size").
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
		Rows(filesListToRows(r.Data)...)

	return table
}

// filesListToRows converts a slice of files into rows for table display
func filesListToRows(files []chunkify.File) [][]string {
	rows := make([][]string, len(files))
	for i, file := range files {
		rows[i] = []string{
			file.CreatedAt.Format(time.RFC822),
			file.StorageId,
			styles.Important.Render(file.Path),
			file.MimeType,
			formatter.Size(file.Size),
		}
	}
	return rows
}

// newFilesListCmd creates a new command for listing files of a job
func newFilesListCmd() *cobra.Command {
	req := FilesListCmd{}

	cmd := &cobra.Command{
		Use:   "files job-id",
		Short: "list all files of a job",
		Long:  `list all files of a job`,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			req.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}

			req.View()
		},
	}

	cmd.Flags().BoolVarP(&req.presign, "presign", "p", false, "Return the presigned URL of the files")

	return cmd
}
