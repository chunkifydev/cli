package version

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const (
	DevVersion   = "dev"
	GithubApiUrl = "https://api.github.com/repos/chunkifydev/cli/releases/latest"
)

var Version = DevVersion

func GetLatestVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GithubApiUrl, nil)
	if err != nil {
		return "", err
	}

	client := http.DefaultClient
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var release struct {
		Version string `json:"tag_name"`
	}
	err = json.Unmarshal(body, &release)
	if err != nil {
		return "", err
	}
	return release.Version, nil
}

// IsUpToDate checks if the current version is up to date with the latest version
// Version dev is always considered up to date and returns the latest version
func IsUpToDate() (bool, string) {
	if Version == DevVersion {
		return true, Version
	}

	latestVersion, err := GetLatestVersion()
	if err != nil {
		return false, ""
	}

	return Version == DevVersion || Version == latestVersion, latestVersion
}
