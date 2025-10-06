package chunkify

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	chunkify "github.com/chunkifydev/chunkify-go"
)

func TestProgressWriter_Write(t *testing.T) {
	// Create a buffer to write to
	var buf bytes.Buffer

	// Create a progress channel
	progressChan := make(chan DownloadProgress, 10)

	// Create a test file
	file := chunkify.File{
		Id:   "file_123",
		Path: "test.txt",
		Url:  "http://example.com/test.txt",
	}

	// Create progress writer
	pw := &progressWriter{
		w:            &buf,
		total:        100,
		written:      0,
		start:        time.Now(),
		updateEvery:  1 * time.Millisecond, // Very frequent updates for testing
		progressChan: progressChan,
		file:         file,
	}

	// Write some data
	testData := []byte("Hello, World!")
	n, err := pw.Write(testData)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
	}

	// Check that data was written to buffer
	if buf.String() != string(testData) {
		t.Errorf("Expected buffer content %q, got %q", string(testData), buf.String())
	}

	// Check that progress was sent
	select {
	case progress := <-progressChan:
		if progress.WrittenBytes != int64(len(testData)) {
			t.Errorf("Expected WrittenBytes %d, got %d", len(testData), progress.WrittenBytes)
		}
		if progress.TotalBytes != 100 {
			t.Errorf("Expected TotalBytes 100, got %d", progress.TotalBytes)
		}
		if progress.File.Id != file.Id {
			t.Errorf("Expected File.Id %s, got %s", file.Id, progress.File.Id)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected progress update but none received")
	}
}

func TestProgressWriter_Write_NoTotal(t *testing.T) {
	var buf bytes.Buffer
	progressChan := make(chan DownloadProgress, 10)

	file := chunkify.File{
		Id:   "file_456",
		Path: "test.txt",
		Url:  "http://example.com/test.txt",
	}

	pw := &progressWriter{
		w:            &buf,
		total:        0, // Unknown total
		written:      0,
		start:        time.Now(),
		updateEvery:  1 * time.Millisecond,
		progressChan: progressChan,
		file:         file,
	}

	testData := []byte("Test data")
	_, err := pw.Write(testData)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check progress for unknown total
	select {
	case progress := <-progressChan:
		if progress.TotalBytes != 0 {
			t.Errorf("Expected TotalBytes 0, got %d", progress.TotalBytes)
		}
		if progress.Progress != 0 {
			t.Errorf("Expected Progress 0, got %f", progress.Progress)
		}
		if progress.Eta != 0 {
			t.Errorf("Expected Eta 0, got %v", progress.Eta)
		}
		if progress.WrittenBytes != int64(len(testData)) {
			t.Errorf("Expected WrittenBytes %d, got %d", len(testData), progress.WrittenBytes)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected progress update but none received")
	}
}

func TestProgressWriter_Write_ProgressCalculation(t *testing.T) {
	var buf bytes.Buffer
	progressChan := make(chan DownloadProgress, 10)

	file := chunkify.File{
		Id:   "file_789",
		Path: "test.txt",
		Url:  "http://example.com/test.txt",
	}

	pw := &progressWriter{
		w:            &buf,
		total:        200,
		written:      0,
		start:        time.Now(),
		updateEvery:  1 * time.Millisecond,
		progressChan: progressChan,
		file:         file,
	}

	// Write 50 bytes (25% of 200)
	testData := make([]byte, 50)
	// Add a small delay to make ETA calculation more reliable
	time.Sleep(10 * time.Millisecond)
	_, err := pw.Write(testData)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	select {
	case progress := <-progressChan:
		expectedProgress := 25.0 // 50/200 * 100
		if progress.Progress != expectedProgress {
			t.Errorf("Expected Progress %f, got %f", expectedProgress, progress.Progress)
		}
		if progress.WrittenBytes != 50 {
			t.Errorf("Expected WrittenBytes 50, got %d", progress.WrittenBytes)
		}
		if progress.TotalBytes != 200 {
			t.Errorf("Expected TotalBytes 200, got %d", progress.TotalBytes)
		}
		// ETA should be calculated (remaining 150 bytes at current speed)
		// Note: ETA might be 0 if speed is very high, so we just check it's not negative
		if progress.Eta < 0 {
			t.Error("Expected ETA to not be negative")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected progress update but none received")
	}
}

func TestDownloadFile_Success(t *testing.T) {
	// Create a test server
	testData := "Hello, World! This is test data for download."
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testData))
	}))
	defer server.Close()

	// Create temporary directory for output
	tempDir, err := os.MkdirTemp("", "download_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "downloaded.txt")
	progressChan := make(chan DownloadProgress, 10)

	file := chunkify.File{
		Id:   "file_test",
		Path: "test.txt",
		Url:  server.URL,
	}

	// Download the file
	ctx := context.Background()
	err = DownloadFile(ctx, file, outputFile, progressChan)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Expected output file to be created")
	}

	// Check file content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("Expected file content %q, got %q", testData, string(content))
	}

	// Check that no .part file remains
	partFile := outputFile + ".part"
	if _, err := os.Stat(partFile); !os.IsNotExist(err) {
		t.Error("Expected .part file to be removed after successful download")
	}
}

func TestDownloadFile_BadStatus(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "download_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "downloaded.txt")
	progressChan := make(chan DownloadProgress, 10)

	file := chunkify.File{
		Id:   "file_test",
		Path: "test.txt",
		Url:  server.URL,
	}

	ctx := context.Background()
	err = DownloadFile(ctx, file, outputFile, progressChan)

	if err == nil {
		t.Error("Expected error for bad status code")
	}

	// Check that output file was not created
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Error("Expected output file to not be created on error")
	}
}

func TestDownloadFile_ContextCancellation(t *testing.T) {
	// Create a test server that takes a long time to respond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "download_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "downloaded.txt")
	progressChan := make(chan DownloadProgress, 10)

	file := chunkify.File{
		Id:   "file_test",
		Path: "test.txt",
		Url:  server.URL,
	}

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = DownloadFile(ctx, file, outputFile, progressChan)

	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	// Check that output file was not created
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Error("Expected output file to not be created on cancellation")
	}
}

func TestDownloadFile_ProgressUpdates(t *testing.T) {
	// Create a test server with larger data
	testData := make([]byte, 1000) // 1KB of data
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "download_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "downloaded.txt")
	progressChan := make(chan DownloadProgress, 100)

	file := chunkify.File{
		Id:   "file_test",
		Path: "test.txt",
		Url:  server.URL,
	}

	ctx := context.Background()

	// Start download in goroutine
	go func() {
		DownloadFile(ctx, file, outputFile, progressChan)
		close(progressChan)
	}()

	// Collect progress updates
	var progressUpdates []DownloadProgress
	for progress := range progressChan {
		progressUpdates = append(progressUpdates, progress)
	}

	// Should have received multiple progress updates
	if len(progressUpdates) == 0 {
		t.Error("Expected to receive progress updates")
	}

	// Check final progress
	if len(progressUpdates) > 0 {
		finalProgress := progressUpdates[len(progressUpdates)-1]
		if finalProgress.WrittenBytes != int64(len(testData)) {
			t.Errorf("Expected final WrittenBytes %d, got %d", len(testData), finalProgress.WrittenBytes)
		}
		// Progress might not be exactly 100.0 due to floating point precision
		if finalProgress.Progress < 99.0 {
			t.Errorf("Expected final Progress >= 99.0, got %f", finalProgress.Progress)
		}
	}
}
