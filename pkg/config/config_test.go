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
