package config

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestStorageConfigProfiles(t *testing.T) {
	keyring.MockInit()
	run := func(args ...string) string {
		t.Helper()
		cmd := NewCommand()
		var output bytes.Buffer
		cmd.SetOut(&output)
		cmd.SetErr(&output)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		return output.String()
	}
	run("storage-id", "stor_chunkify_default")
	run("storage-id", "stor_aws_staging", "--profile", "staging")
	if got := run("storage-id", "--profile", "staging"); !strings.Contains(got, "stor_aws_staging") {
		t.Fatalf("unexpected saved storage: %s", got)
	}
	cfg := &Config{Profile: "staging"}
	if err := cfg.LoadStorageID(); err != nil || cfg.StorageID != "stor_aws_staging" {
		t.Fatalf("profile storage: %q, %v", cfg.StorageID, err)
	}
	cfg.Profile = ""
	if err := cfg.LoadStorageID(); err != nil || cfg.StorageID != "stor_chunkify_default" {
		t.Fatalf("default storage: %q, %v", cfg.StorageID, err)
	}
	cfg.Profile = "missing"
	if err := cfg.LoadStorageID(); err != nil || cfg.StorageID != "" {
		t.Fatalf("must not inherit another profile's storage: %q, %v", cfg.StorageID, err)
	}
	run("storage-id", "", "--profile", "staging")
	cfg.Profile = "staging"
	if err := cfg.LoadStorageID(); err != nil || cfg.StorageID != "" {
		t.Fatalf("cleared storage: %q, %v", cfg.StorageID, err)
	}
	if value, err := Get(ConfigStorageIDKey); err != nil || value != "stor_chunkify_default" {
		t.Fatal("clearing staging changed the default profile")
	}
}

func TestStorageConfigRejectsInvalidID(t *testing.T) {
	keyring.MockInit()
	for _, value := range []string{"bucket", "stor_", "stor_aws/path", "stor_aws with spaces", "stor_" + strings.Repeat("a", 60)} {
		cmd := NewCommand()
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs([]string{"storage-id", value})
		if err := cmd.Execute(); err == nil {
			t.Errorf("accepted invalid ID %q", value)
		}
	}
}

func TestLoadStorageIDReportsKeyringFailure(t *testing.T) {
	want := errors.New("keyring unavailable")
	keyring.MockInitWithError(want)
	t.Cleanup(keyring.MockInit)
	if err := (&Config{}).LoadStorageID(); !errors.Is(err, want) {
		t.Fatalf("expected keyring error, got %v", err)
	}
}

func TestTeamTokenConfigAndProfiles(t *testing.T) {
	keyring.MockInit()
	t.Setenv("CHUNKIFY_TEAM_TOKEN", "")
	cmd := NewCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"team-token", "sk_team_secret1234", "--profile", "staging"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "sk_team_secret1234") {
		t.Fatal("config output exposed the team token")
	}
	cfg := &Config{Profile: "staging"}
	if err := cfg.SetTeamToken(); err != nil || cfg.TeamToken != "sk_team_secret1234" {
		t.Fatalf("team token for profile: %q, %v", cfg.TeamToken, err)
	}
	cfg.Profile = "other"
	if err := cfg.SetTeamToken(); !errors.Is(err, keyring.ErrNotFound) {
		t.Fatalf("another profile inherited the token: %v", err)
	}
	t.Setenv("CHUNKIFY_TEAM_TOKEN", "sk_team_environment")
	if err := cfg.SetTeamToken(); err != nil || cfg.TeamToken != "sk_team_environment" {
		t.Fatalf("environment token: %q, %v", cfg.TeamToken, err)
	}
}

func TestOpenAPIURLConfigAndProfiles(t *testing.T) {
	keyring.MockInit()
	t.Setenv("CHUNKIFY_OPENAPI_URL", "")
	run := func(args ...string) error {
		t.Helper()
		cmd := NewCommand()
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	const stagingURL = "https://staging.example.com/openapi.json"
	if err := run("openapi-url", stagingURL, "--profile", "staging"); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{Profile: "staging"}
	if got, err := cfg.OpenAPIURL(); err != nil || got != stagingURL {
		t.Fatalf("staging OpenAPI URL: %q, %v", got, err)
	}
	cfg.Profile = "other"
	if got, err := cfg.OpenAPIURL(); err != nil || got != "" {
		t.Fatalf("other profile inherited URL: %q, %v", got, err)
	}
	t.Setenv("CHUNKIFY_OPENAPI_URL", "https://override.example.com/schema.json")
	if got, err := cfg.OpenAPIURL(); err != nil || got != "https://override.example.com/schema.json" {
		t.Fatalf("environment override: %q, %v", got, err)
	}
	t.Setenv("CHUNKIFY_OPENAPI_URL", "")
	if err := run("openapi-url", "", "--profile", "staging"); err != nil {
		t.Fatal(err)
	}
	cfg.Profile = "staging"
	if got, err := cfg.OpenAPIURL(); err != nil || got != "" {
		t.Fatalf("cleared URL: %q, %v", got, err)
	}
}

func TestOpenAPIURLRejectsInvalidValues(t *testing.T) {
	keyring.MockInit()
	for _, value := range []string{"openapi.json", "ftp://example.com/openapi.json", "https://", "https://user:pass@example.com/openapi.json"} {
		cmd := NewCommand()
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs([]string{"openapi-url", value})
		if err := cmd.Execute(); err == nil {
			t.Errorf("accepted invalid URL %q", value)
		}
	}
}
