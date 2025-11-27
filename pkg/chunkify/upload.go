package chunkify

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/chunkifydev/chunkify-go"
)

// UploadBlobWithContext uploads a file to the specified URL.
// UploadCreate must be called first to get the URL of the upload.
// The r parameter is the reader of the file to upload.
// Returns an error if the request fails.
func UploadBlob(ctx context.Context, r io.Reader, uploadResponse *chunkify.Upload) error {
	// http put request with reader
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadResponse.UploadURL, r)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file: %s", resp.Status)
	}

	return nil
}

// uploadProgressWriter tracks upload progress and emits fractional updates on a channel.
type uploadProgressWriter struct {
	ch        chan UploadProgress
	total     int64
	written   int64
	lastEmit  time.Time
	startTime time.Time
}

func (w *uploadProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.written += int64(n)

	// Compute metrics
	var progress float64 = -1
	if w.total > 0 {
		progress = (float64(w.written) / float64(w.total)) * 100
	}
	elapsed := time.Since(w.startTime).Seconds()
	var bytesPerSecond float64
	if elapsed > 0 {
		bytesPerSecond = float64(w.written) / elapsed
	}
	etaSeconds := -1.0
	if w.total > 0 && bytesPerSecond > 0 {
		remaining := float64(w.total - w.written)
		if remaining <= 0 {
			etaSeconds = 0
		} else {
			etaSeconds = remaining / bytesPerSecond
		}

		if etaSeconds < -1 {
			etaSeconds = -1.0
		}
	}

	// Throttle emissions to ~10Hz and always emit on completion
	now := time.Now()
	if (w.total > 0 && progress >= 100) || w.lastEmit.IsZero() || now.Sub(w.lastEmit) >= 100*time.Millisecond {
		select {
		case w.ch <- UploadProgress{Progress: progress, TotalBytes: w.total, WrittenBytes: w.written, Speed: bytesPerSecond, Eta: time.Duration(etaSeconds * float64(time.Second))}:
		default:
		}
		w.lastEmit = now
	}

	return n, nil
}

// UploadProgress represents the progress of an ongoing file upload.
type UploadProgress struct {
	// Progress is a float between 0 and 100 indicating the percentage of upload completion
	Progress float64
	// WrittenBytes is the number of bytes uploaded so far
	WrittenBytes int64
	// TotalBytes is the total size of the file in bytes (-1 if unknown)
	TotalBytes int64
	// Speed is the average upload throughput since start (bytes/sec)
	Speed float64
	// Eta is the estimated seconds remaining (-1 if unknown)
	Eta time.Duration
}

// UploadBlobWithProgress uploads a file to the specified URL and sends the progress to the progress channel.
// UploadCreate must be called first to get the URL of the upload.
// The r parameter is the reader of the file to upload.
// Returns an error if the request fails.
func UploadBlobWithProgress(ctx context.Context, r io.Reader, uploadResponse *chunkify.Upload, progress chan UploadProgress) error {
	// Determine total size if possible
	var size int64 = -1
	if seeker, ok := r.(io.Seeker); ok {
		if curr, err := seeker.Seek(0, io.SeekCurrent); err == nil {
			if end, err2 := seeker.Seek(0, io.SeekEnd); err2 == nil {
				size = end - curr
				_, _ = seeker.Seek(curr, io.SeekStart)
			}
		}
	}

	// Wrap reader to capture progress
	pw := &uploadProgressWriter{ch: progress, total: size, startTime: time.Now()}
	body := io.TeeReader(r, pw)

	// http put request with reader
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadResponse.UploadURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	if size >= 0 {
		req.ContentLength = size
	}

	// Send initial progress value without blocking
	select {
	case progress <- UploadProgress{Progress: func() float64 {
		if size > 0 {
			return 0
		}
		return -1
	}(), TotalBytes: size, WrittenBytes: 0, Speed: 0, Eta: -1}:
	default:
	}

	defer func() {
		// Ensure channel is closed by the producer when finished
		close(progress)
	}()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file: %s", resp.Status)
	}

	// Ensure final 100% progress if total was known
	if size > 0 {
		elapsed := time.Since(pw.startTime).Seconds()
		bps := 0.0
		if elapsed > 0 {
			bps = float64(size) / elapsed
		}
		select {
		case progress <- UploadProgress{Progress: 100, TotalBytes: size, WrittenBytes: size, Speed: bps, Eta: 0 * time.Second}:
		default:
		}
	}

	return nil
}
