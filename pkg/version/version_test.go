package version

import (
	"strings"
	"testing"
)

func TestUpdateInstructions(t *testing.T) {
	original := InstallMethod
	t.Cleanup(func() { InstallMethod = original })
	InstallMethod = "npm"
	if instructions := UpdateInstructions(); !strings.Contains(instructions, "npm install -g @chunkify/cli@latest") || !strings.Contains(instructions, "npx @chunkify/cli@latest") {
		t.Fatalf("missing npm update instructions: %s", instructions)
	}
	InstallMethod = ""
	if instructions := UpdateInstructions(); !strings.Contains(instructions, "chunkify update") {
		t.Fatalf("missing native update instructions: %s", instructions)
	}
}
