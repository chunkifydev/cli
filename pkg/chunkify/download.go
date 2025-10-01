package chunkify

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	chunkify "github.com/chunkifydev/chunkify-go"
)

type progressWriter struct {
	w            io.Writer
	total        int64 // content-length (may be 0 when unknown)
	written      int64
	start        time.Time
	lastUpdate   time.Time
	updateEvery  time.Duration
	progressChan chan DownloadProgress
	file         chunkify.File
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	pw.written += int64(n)

	now := time.Now()
	if pw.lastUpdate.IsZero() || now.Sub(pw.lastUpdate) >= pw.updateEvery || err != nil {
		pw.print(now)
		pw.lastUpdate = now
	}

	return n, err
}

func (pw *progressWriter) print(now time.Time) {
	elapsed := now.Sub(pw.start).Seconds()
	if elapsed <= 0 {
		elapsed = 0.001
	}
	speed := float64(pw.written) / elapsed // bytes/sec

	pr := DownloadProgress{
		File:         pw.file,
		TotalBytes:   pw.total,
		WrittenBytes: pw.written,
		Speed:        speed,
	}

	if pw.total > 0 {
		remain := pw.total - pw.written
		eta := time.Duration(float64(remain)/speed) * time.Second
		percent := float64(pw.written) * 100 / float64(pw.total)
		pr.Progress = percent
		pr.Eta = eta.Truncate(time.Second)
		//fmt.Printf("\r%.1f%%  %s / %s  —  %s/s  ETA %s",
		//percent, humanBytes(pw.written), humanBytes(pw.total), humanBytes(int64(speed)), eta.Truncate(time.Second))
	} else {
		// unknown length
		//fmt.Printf("\r%s downloaded  —  %s/s",
		//humanBytes(pw.written), humanBytes(int64(speed)))
	}

	pw.progressChan <- pr
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// DownloadFile streams a URL to `output` with console progress.
func DownloadFile(ctx context.Context, file chunkify.File, output string, progressChan chan DownloadProgress) error {
	// HTTP client with sane timeouts
	client := &http.Client{
		Timeout: 0, // no overall timeout; we rely on ctx + transport timeouts below
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 2 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, file.Url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Prepare output file (atomic write via temp file)
	tmp := output + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() {
		out.Close()
		// best effort: if we exit with error, leave the .part file for troubleshooting/resume logic
	}()

	var total int64
	if resp.ContentLength > 0 {
		total = resp.ContentLength
	}

	pw := &progressWriter{
		w:            out,
		total:        total,
		start:        time.Now(),
		updateEvery:  200 * time.Millisecond,
		progressChan: progressChan,
		file:         file,
	}

	// Stream copy with progress
	if _, err = io.Copy(pw, resp.Body); err != nil {
		//fmt.Println() // end progress line
		return fmt.Errorf("download copy: %w", err)
	}

	// Final progress line
	pw.print(time.Now())

	//fmt.Println()

	// Close file before rename to ensure flush
	if err := out.Close(); err != nil {
		return fmt.Errorf("close file: %w", err)
	}

	// Atomically move into place
	if err := os.Rename(tmp, output); err != nil {
		return fmt.Errorf("rename %q -> %q: %w", tmp, output, err)
	}

	//fmt.Printf("Saved to %s\n", output)
	return nil
}
