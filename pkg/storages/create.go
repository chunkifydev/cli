package storages

import (
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command for creating a new storage configuration
type CreateCmd struct {
	Params chunkify.StorageCreateParams // Parameters for creating the storage
	Data   chunkify.Storage             `json:"-"` // The created storage data
}

// Execute creates a new storage with the specified parameters
func (r *CreateCmd) Execute() error {
	storage, err := cmd.Config.Client.StorageCreate(r.Params)
	if err != nil {
		return err
	}

	r.Data = storage

	return nil
}

// View displays the newly created storage information
func (r *CreateCmd) View() {
	projectList := ListCmd{Data: []chunkify.Storage{r.Data}}
	projectList.View()
}

// newCreateCmd creates and configures a new cobra command for creating storage configurations
func newCreateCmd() *cobra.Command {
	req := CreateCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new storage",
		Long:  `Create a new storage`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}
			req.View()
		},
	}

	// Configure command flags for storage parameters
	cmd.Flags().StringVar(&req.Params.Endpoint, "endpoint", "", "The S3 compatible endpoint of the storage")
	cmd.Flags().StringVar(&req.Params.AccessKeyId, "access-key-id", "", "The S3 access key id of the storage")
	cmd.Flags().StringVar(&req.Params.SecretAccessKey, "secret-access-key", "", "The S3 secret access key of the storage")
	cmd.Flags().StringVar(&req.Params.Bucket, "bucket", "", "The S3 bucket name of the storage")
	cmd.Flags().StringVar(&req.Params.Provider, "provider", "", "The storage provider: chunkify, aws, cloudflare (required)")
	cmd.Flags().StringVar(&req.Params.Region, "region", "", "The region of the storage (required)")
	cmd.Flags().StringVar(&req.Params.Location, "location", "", "The location of the storage: US, EU, ASIA")
	cmd.Flags().BoolVar(&req.Params.Public, "public", false, "The uploaded files will be publicly available or not")

	// Mark provider flag as required for all storage configurations
	cmd.MarkFlagRequired("provider")

	// Add validation for required flags based on provider
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")

		// Validate chunkify provider requirements
		if provider == "chunkify" {
			if err := cmd.MarkFlagRequired("region"); err != nil {
				return err
			}
			// Chunkify provider does not require any other flags
			return nil
		}

		// Validate AWS provider requirements
		if provider == "aws" {
			if err := cmd.MarkFlagRequired("region"); err != nil {
				return err
			}
		}

		// Validate Cloudflare provider requirements
		if provider == "cloudflare" {
			if err := cmd.MarkFlagRequired("endpoint"); err != nil {
				return err
			}
			if err := cmd.MarkFlagRequired("location"); err != nil {
				return err
			}
		}

		// All providers except chunkify require bucket, access-key-id and secret-access-key
		if err := cmd.MarkFlagRequired("access-key-id"); err != nil {
			return err
		}
		if err := cmd.MarkFlagRequired("secret-access-key"); err != nil {
			return err
		}
		if err := cmd.MarkFlagRequired("bucket"); err != nil {
			return err
		}

		return nil
	}

	return cmd
}
