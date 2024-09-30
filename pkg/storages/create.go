package storages

import (
	"github.com/level63/cli/pkg/api"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint"`
	AccessKeyId     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Name            string `json:"name"`
	Provider        string `json:"provider"`
	Region          string `json:"region"`
	Location        string `json:"location"`
	Public          bool   `json:"public"`

	Data api.Storage
}

func (r *CreateCmd) Execute() error {
	project, err := api.ApiRequest[api.Storage](api.Request{Config: cmd.Config, Path: "/api/storages", Method: "POST", Body: r})
	if err != nil {
		return err
	}

	r.Data = project

	return nil
}

func (r *CreateCmd) View() {
	projectList := ListCmd{Data: []api.Storage{r.Data}}
	projectList.View()
}

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

	cmd.Flags().StringVar(&req.Name, "name", "", "The name of the storage (required)")
	cmd.Flags().StringVar(&req.Endpoint, "endpoint", "", "The S3 compatible endpoint of the storage")
	cmd.Flags().StringVar(&req.AccessKeyId, "access-key-id", "", "The S3 access key id of the storage (required)")
	cmd.Flags().StringVar(&req.SecretAccessKey, "secret-access-key", "", "The S3 secret access key of the storage (required)")
	cmd.Flags().StringVar(&req.Bucket, "bucket", "", "The S3 bucket name of the storage (required)")
	cmd.Flags().StringVar(&req.Provider, "provider", "", "The storage provider: level63, aws, cloudflare (required)")
	cmd.Flags().StringVar(&req.Region, "region", "", "The region of the storage (required)")
	cmd.Flags().StringVar(&req.Location, "location", "", "The location of the storage: US, EU, ASIA (required)")
	cmd.Flags().BoolVar(&req.Public, "public", false, "The uploaded files will be publicly available or not")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("access-key-id")
	cmd.MarkFlagRequired("secret-access-key")
	cmd.MarkFlagRequired("bucket")
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("location")

	return cmd
}
