package tokens

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
	Data []chunkify.Token
}

func (r *ListCmd) Execute() error {
	tokens, err := cmd.Config.Client.TokenList()
	if err != nil {
		return err
	}

	r.Data = tokens

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
		fmt.Println(styles.DefaultText.Render("No token found."))
		return
	}

	fmt.Println(r.tokensTable())
	if len(r.Data) > 1 {
		fmt.Println(styles.Debug.MarginTop(1).Render(fmt.Sprintf("Total: %d\n", len(r.Data))))
	}
}

func (r *ListCmd) tokensTable() *table.Table {
	rightCols := []int{}
	centerCols := []int{4}

	table := table.New().
		BorderRow(true).
		BorderColumn(false).
		BorderStyle(styles.Border).
		Headers("Date", "Id", "Name", "Scope", "Project Id").
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
		Rows(tokensListToRows(r.Data)...)

	return table
}

func tokensListToRows(tokens []chunkify.Token) [][]string {
	rows := make([][]string, len(tokens))
	for i, token := range tokens {
		rows[i] = []string{
			token.CreatedAt.Format(time.RFC822),
			styles.Id.Render(token.Id),
			token.Name,
			token.Scope,
			token.ProjectId,
		}
	}
	return rows
}

func newListCmd() *cobra.Command {
	req := ListCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all tokens",
		Long:  `list all tokens`,
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
