package chunkify

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/chunkify-go/option"
)

func TestStartTranscodingProgress(t *testing.T) {
	for _, tt := range []struct {
		name      string
		responses []int // status code per poll; the last one repeats
		wantError string
	}{
		{name: "ready", responses: []int{200}},
		{name: "transient error then ready", responses: []int{503, 200}},
		{name: "deleted transcoding stops polling", responses: []int{404}, wantError: "error retrieving transcoding tcg_test"},
		{name: "revoked token stops polling", responses: []int{401}, wantError: "401"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var polls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/transcodings/tcg_test" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				n := int(polls.Add(1)) - 1
				code := tt.responses[min(n, len(tt.responses)-1)]
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(code)
				if code == http.StatusOK {
					fmt.Fprint(w, `{"data":{"id":"tcg_test","status":"ready","progress":100}}`)
					return
				}
				fmt.Fprint(w, `{"status":"error","error":{"message":"nope"}}`)
			}))
			defer server.Close()

			client := chunkify.NewClient(option.WithBaseURL(server.URL), option.WithProjectAccessToken("project-secret"), option.WithMaxRetries(0))
			app := NewApp()
			app.Client = &client
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			go app.StartTranscodingProgress(ctx, "tcg_test")

			select {
			case err := <-app.Progress.TranscodingCompleted:
				if tt.wantError == "" {
					if err != nil {
						t.Fatalf("want normal completion, got %v", err)
					}
					if app.Transcoding == nil || app.Transcoding.Status != chunkify.TranscodingStatusReady {
						t.Fatalf("want ready transcoding, got %+v", app.Transcoding)
					}
				} else if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("want error containing %q, got %v", tt.wantError, err)
				}
			case <-ctx.Done():
				t.Fatal("polling did not finish")
			}
		})
	}
}

func TestIsPermanentAPIError(t *testing.T) {
	for code, want := range map[int]bool{400: true, 401: true, 403: true, 404: true, 408: false, 429: false, 500: false, 503: false} {
		err := fmt.Errorf("wrapped: %w", &chunkify.Error{StatusCode: code})
		if got := isPermanentAPIError(err); got != want {
			t.Errorf("isPermanentAPIError(%d) = %v, want %v", code, got, want)
		}
	}
	if isPermanentAPIError(context.DeadlineExceeded) {
		t.Error("network and context errors must be retried")
	}
}
