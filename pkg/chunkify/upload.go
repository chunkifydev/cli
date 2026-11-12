package chunkify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/chunkifydev/chunkify-go"
	"golang.org/x/sync/errgroup"
)

const (
	// multipartConcurrency is the number of parts transferred at once.
	multipartConcurrency = 4
	// multipartAttempts is the number of tries for each part before giving up.
	multipartAttempts = 3
)

var multipartRetryDelay = time.Second

// UploadBlob uploads a file to the specified URL.
// UploadCreate must be called first to get the URL of the upload.
// The r parameter is the reader of the file to upload.
// The caller must complete the upload session after the transfer succeeds.
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
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
	w.advance(int64(len(p)))
	return len(p), nil
}

// advance records n transferred bytes. A negative n rolls back bytes from a
// failed attempt that will be sent again.
func (w *uploadProgressWriter) advance(n int64) {
	w.written += n

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
// The caller must complete the upload session after the transfer succeeds.
// Returns an error if the request fails.
func UploadBlobWithProgress(ctx context.Context, r io.Reader, uploadResponse *chunkify.Upload, progress chan UploadProgress) error {
	defer close(progress)

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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
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

// UploadFileWithProgress transfers r to the upload session using the transfer
// mode chosen by the API and returns the parameters for completing it. The
// progress channel is closed when the transfer ends.
func UploadFileWithProgress(ctx context.Context, r io.ReaderAt, size int64, upload *chunkify.Upload, progress chan UploadProgress) (chunkify.UploadCompleteParams, error) {
	if upload.UploadMode == chunkify.UploadUploadModeMultipart {
		parts, err := UploadMultipartWithProgress(ctx, r, size, upload.Multipart, progress)
		if err != nil {
			return chunkify.UploadCompleteParams{}, err
		}
		return chunkify.UploadCompleteParams{Parts: parts}, nil
	}
	return chunkify.UploadCompleteParams{}, UploadBlobWithProgress(ctx, io.NewSectionReader(r, 0, size), upload, progress)
}

// lockedProgress serializes progress updates from concurrent part transfers.
type lockedProgress struct {
	mu sync.Mutex
	w  *uploadProgressWriter
}

func (p *lockedProgress) advance(n int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.w.advance(n)
}

// countingReader reports bytes read from a part to the shared progress.
type countingReader struct {
	r        io.Reader
	progress *lockedProgress
	read     int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.read += int64(n)
	c.progress.advance(int64(n))
	return n, err
}

// UploadMultipartWithProgress uploads each part of r to its signed URL and
// returns the part numbers and storage ETags in ascending order. Parts are
// retried with their original URLs. The progress channel is closed on return.
func UploadMultipartWithProgress(ctx context.Context, r io.ReaderAt, size int64, multipart chunkify.UploadMultipart, progress chan UploadProgress) ([]chunkify.UploadCompleteParamsPart, error) {
	defer close(progress)

	partSize := multipart.PartSizeBytes
	if partSize <= 0 || len(multipart.Parts) == 0 {
		return nil, fmt.Errorf("missing multipart upload details")
	}
	if want := (size + partSize - 1) / partSize; int64(len(multipart.Parts)) != want {
		return nil, fmt.Errorf("multipart upload has %d parts, expected %d for %d bytes", len(multipart.Parts), want, size)
	}
	for i, part := range multipart.Parts {
		if part.PartNumber != int64(i+1) || part.UploadURL == "" {
			return nil, fmt.Errorf("invalid multipart part %d", part.PartNumber)
		}
	}
	if !multipart.ExpiresAt.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, multipart.ExpiresAt)
		defer cancel()
	}

	shared := &lockedProgress{w: &uploadProgressWriter{ch: progress, total: size, startTime: time.Now()}}
	shared.advance(0)

	results := make([]chunkify.UploadCompleteParamsPart, len(multipart.Parts))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(multipartConcurrency)
	for i, part := range multipart.Parts {
		offset := (part.PartNumber - 1) * partSize
		length := min(partSize, size-offset)
		g.Go(func() error {
			etag, err := uploadPart(gctx, part.UploadURL, io.NewSectionReader(r, offset, length), length, shared)
			if err != nil {
				return fmt.Errorf("part %d: %w", part.PartNumber, err)
			}
			results[i] = chunkify.UploadCompleteParamsPart{PartNumber: part.PartNumber, Etag: etag}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// errPermanent marks a storage response that retrying cannot fix.
var errPermanent = errors.New("permanent failure")

func uploadPart(ctx context.Context, url string, part *io.SectionReader, length int64, progress *lockedProgress) (string, error) {
	var lastErr error
	for attempt := range multipartAttempts {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(multipartRetryDelay * time.Duration(attempt)):
			}
		}
		if _, err := part.Seek(0, io.SeekStart); err != nil {
			return "", err
		}
		body := &countingReader{r: part, progress: progress}
		etag, err := putPart(ctx, url, body, length)
		if err == nil {
			return etag, nil
		}
		// These bytes are sent again on the next attempt.
		progress.advance(-body.read)
		lastErr = err
		if errors.Is(err, errPermanent) || ctx.Err() != nil {
			break
		}
	}
	return "", lastErr
}

func putPart(ctx context.Context, url string, body io.Reader, length int64) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errPermanent, err)
	}
	req.ContentLength = length

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("failed to upload part: %s", resp.Status)
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			err = fmt.Errorf("%w: %v", errPermanent, err)
		}
		return "", err
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		return "", fmt.Errorf("%w: storage returned no ETag", errPermanent)
	}
	return etag, nil
}
