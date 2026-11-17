package chunkify

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	chunkify "github.com/chunkifydev/chunkify-go"
)

// apiTranscoding decodes a transcoding as the SDK does, so omitted fields are
// marked as such.
func apiTranscoding(t *testing.T, body string) *chunkify.Transcoding {
	t.Helper()
	var transcoding chunkify.Transcoding
	if err := json.Unmarshal([]byte(body), &transcoding); err != nil {
		t.Fatal(err)
	}
	return &transcoding
}

func jsonView(t *testing.T, app *App) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(app.JSONView()), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestJSONViewProgress(t *testing.T) {
	for _, tt := range []struct {
		name        string
		status      int
		transcoding string
		want        any
	}{
		{name: "encoding", status: Transcoding, transcoding: `{"id":"tcg_test","status":"preparing","processing_stage":"encoding","progress":42}`, want: 42.0},
		{name: "audio-only output omits progress", status: Transcoding, transcoding: `{"id":"tcg_test","status":"preparing","processing_stage":"encoding"}`, want: 0.0},
		{name: "not polled yet", status: Transcoding, want: 0.0},
		{name: "ready audio-only output", status: Transcoding, transcoding: `{"id":"tcg_test","status":"ready"}`, want: 100.0},
		{name: "completed", status: Completed, transcoding: `{"id":"tcg_test","status":"ready"}`, want: 100.0},
		{name: "completed upload only", status: Completed, want: 100.0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			app.Status = tt.status
			if tt.transcoding != "" {
				app.Transcoding = apiTranscoding(t, tt.transcoding)
			}
			out := jsonView(t, app)
			if got, ok := out["progress"]; !ok || got != tt.want {
				t.Fatalf("progress = %v (present %v), want %v", got, ok, tt.want)
			}
			for _, key := range []string{"status", "progress", "fps", "speed", "out_time", "eta"} {
				if _, ok := out[key]; !ok {
					t.Errorf("missing key %q in %v", key, out)
				}
			}
		})
	}
}

func TestTranscodingViewReadyAudioOnly(t *testing.T) {
	app := NewApp()
	app.Command = &ChunkifyCommand{Format: "m4a_aac"}
	app.Status = Transcoding
	app.Transcoding = apiTranscoding(t, `{"id":"tcg_test","status":"ready"}`)
	view, statusInfo := app.transcodingView()
	if strings.Contains(view, "▯") || !strings.Contains(statusInfo, "100%") {
		t.Fatalf("a ready audio-only transcoding must show a full bar and 100%%:\n%s\n%s", view, statusInfo)
	}
}

func TestAppErr(t *testing.T) {
	sourceErr := errors.New("source not found")
	for _, tt := range []struct {
		name   string
		status int
		err    error
		want   error
	}{
		{name: "completed", status: Completed},
		{name: "failed", status: Failed, err: sourceErr, want: sourceErr},
		{name: "cancelled", status: Cancelled, want: ErrCancelled},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			app.Status = tt.status
			app.Error = tt.err
			if got := app.Err(); got != tt.want {
				t.Fatalf("Err() = %v, want %v", got, tt.want)
			}
		})
	}
	app := NewApp()
	app.Status = Transcoding
	if app.Err() == nil {
		t.Fatal("a run that stopped before completing must report an error")
	}
}
