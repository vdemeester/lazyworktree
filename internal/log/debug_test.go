package log

import (
	"os"
	"path/filepath"
	"testing"
)

func resetDebugLogger(t *testing.T) func() {
	t.Helper()

	globalDebugLogger.mu.Lock()
	prevFile := globalDebugLogger.file
	prevBuffer := append([]byte(nil), globalDebugLogger.buffer...)
	prevDiscard := globalDebugLogger.discard
	globalDebugLogger.file = nil
	globalDebugLogger.buffer = nil
	globalDebugLogger.discard = false
	globalDebugLogger.mu.Unlock()

	return func() {
		globalDebugLogger.mu.Lock()
		if globalDebugLogger.file != nil {
			_ = globalDebugLogger.file.Close()
		}
		globalDebugLogger.file = prevFile
		globalDebugLogger.buffer = prevBuffer
		globalDebugLogger.discard = prevDiscard
		globalDebugLogger.mu.Unlock()
	}
}

func TestSetFileFailureDiscardsLogs(t *testing.T) {
	restore := resetDebugLogger(t)
	t.Cleanup(restore)

	unwritableDir := t.TempDir()
	if err := os.Chmod(unwritableDir, 0o500); err != nil { //nolint:gosec
		t.Fatalf("set directory permissions: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(unwritableDir, 0o700) //nolint:gosec
	})

	logPath := filepath.Join(unwritableDir, "debug.log")
	if err := SetFile(logPath); err == nil {
		t.Fatalf("expected SetFile to fail for %q", logPath)
	}

	globalDebugLogger.mu.Lock()
	discard := globalDebugLogger.discard
	bufferLen := len(globalDebugLogger.buffer)
	globalDebugLogger.mu.Unlock()

	if !discard {
		t.Fatalf("expected discard to be enabled after SetFile failure")
	}
	if bufferLen != 0 {
		t.Fatalf("expected buffer to be cleared after SetFile failure")
	}

	Printf("should be discarded")

	globalDebugLogger.mu.Lock()
	bufferLen = len(globalDebugLogger.buffer)
	globalDebugLogger.mu.Unlock()

	if bufferLen != 0 {
		t.Fatalf("expected buffer to remain empty after logging")
	}
}
