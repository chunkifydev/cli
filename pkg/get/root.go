package get

import (
	"fmt"
	"strings"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/files"
	"github.com/chunkifydev/cli/pkg/jobs"
	"github.com/chunkifydev/cli/pkg/notifications"
	"github.com/chunkifydev/cli/pkg/projects"
	"github.com/chunkifydev/cli/pkg/sources"
	"github.com/chunkifydev/cli/pkg/storages"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/chunkifydev/cli/pkg/uploads"
	"github.com/chunkifydev/cli/pkg/webhooks"

	"github.com/spf13/cobra"
)

const (
	jobPrefixId          = "job"  // Prefix for job-related identifiers
	filePrefixId         = "file" // Prefix for file-related identifiers
	projectPrefixId      = "proj" // Prefix for project-related identifiers
	storagePrefixId      = "stor" // Prefix for storage-related identifiers
	sourcePrefixId       = "src"  // Prefix for source-related identifiers
	webhookPrefixId      = "wh"   // Prefix for webhook-related identifiers
	notificationPrefixId = "notf" // Prefix for notification-related identifiers
	uploadPrefixId       = "upl"  // Prefix for upload-related identifiers
)

type Getter interface {
	Execute() error
	View()
}

// NewCommand creates and configures a new root command for project management
func NewCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get object-id",
		Short: "get info about an object",
		Long:  "get info about an object",
		Args:  cobra.ExactArgs(1), // Requires exactly one argument (project ID)
		Run: func(cmd *cobra.Command, args []string) {
			object, err := objectCommand(args[0])
			if err != nil {
				printError(err)
				return
			}

			if err := object.Execute(); err != nil {
				printError(err)
				return
			}
			object.View()
		},
	}

	return cmd
}

func objectCommand(objectId string) (Getter, error) {
	prefixParts := strings.Split(objectId, "_")
	if len(prefixParts) < 2 {
		return nil, fmt.Errorf("invalid object id: %s", objectId)
	}

	prefix := prefixParts[0]

	switch prefix {
	case jobPrefixId:
		return &jobs.GetCmd{Id: objectId}, nil
	case filePrefixId:
		return &files.GetCmd{Id: objectId}, nil
	case projectPrefixId:
		return &projects.GetCmd{Id: objectId}, nil
	case sourcePrefixId:
		return &sources.GetCmd{Id: objectId}, nil
	case storagePrefixId:
		return &storages.GetCmd{Id: objectId}, nil
	case uploadPrefixId:
		return &uploads.GetCmd{Id: objectId}, nil
	case webhookPrefixId:
		return &webhooks.GetCmd{Id: objectId}, nil
	case notificationPrefixId:
		return &notifications.GetCmd{Id: objectId}, nil
	}

	return nil, fmt.Errorf("invalid object id: %s", objectId)
}

// printError formats and prints an error message using the defined style
func printError(err error) {
	fmt.Println(styles.Error.Render(err.Error()))
}
