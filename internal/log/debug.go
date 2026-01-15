package log

import (
	"log"
	"os"
	"sync"
)

// DebugLogger handles debug logging to file and/or buffering.
// It implements io.Writer to be compatible with standard log.Logger.
type DebugLogger struct {
	mu      sync.Mutex
	file    *os.File
	buffer  []byte
	discard bool
}

var (
	globalDebugLogger = &DebugLogger{}
	// stdLogger wraps our custom writer to provide standard log formatting
	stdLogger = log.New(globalDebugLogger, "", log.LstdFlags|log.Lmicroseconds)
)

// Write implements io.Writer.
// It writes to the file if set, otherwise appends to the buffer.
func (l *DebugLogger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.discard {
		return len(p), nil
	}

	if l.file != nil {
		n, err = l.file.Write(p)
		// Sync to disk to ensure messages are written immediately
		// ignoring sync errors as they are not critical for logging
		_ = l.file.Sync()
		return n, err
	}

	// Buffer the output
	// Need to copy because p might be reused by the caller
	b := make([]byte, len(p))
	copy(b, p)
	l.buffer = append(l.buffer, b...)
	return len(p), nil
}

// SetFile sets the debug log file path. Creates the file if it doesn't exist.
// If path is empty, discards all buffered logs and future logs.
func SetFile(path string) error {
	globalDebugLogger.mu.Lock()
	defer globalDebugLogger.mu.Unlock()

	// Close any previously opened file.
	if globalDebugLogger.file != nil {
		_ = globalDebugLogger.file.Close()
		globalDebugLogger.file = nil
	}

	if path == "" {
		globalDebugLogger.discard = true
		globalDebugLogger.buffer = nil
		return nil
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) //nolint:gosec
	if err != nil {
		globalDebugLogger.discard = true
		globalDebugLogger.buffer = nil
		return err
	}

	globalDebugLogger.file = f
	globalDebugLogger.discard = false

	// Flush buffer to file.
	if len(globalDebugLogger.buffer) > 0 {
		_, _ = f.Write(globalDebugLogger.buffer)
		_ = f.Sync()
		globalDebugLogger.buffer = nil
	}

	return nil
}

// Printf writes a formatted debug message via the standard logger.
func Printf(format string, args ...any) {
	stdLogger.Printf(format, args...)
}

// Println writes a debug message via the standard logger.
func Println(v ...any) {
	stdLogger.Println(v...)
}

// Close closes the debug log file if open.
func Close() error {
	globalDebugLogger.mu.Lock()
	defer globalDebugLogger.mu.Unlock()

	if globalDebugLogger.file == nil {
		return nil
	}

	err := globalDebugLogger.file.Close()
	globalDebugLogger.file = nil
	return err
}
