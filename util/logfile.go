package util

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const maxAppLogBytes int64 = 5 << 20

// RotatingLogWriter keeps the active log and one backup, bounding total disk
// usage to roughly 10 MB. Writes are serialized so concurrent HTTP handlers
// cannot rotate or write the file at the same time.
type RotatingLogWriter struct {
	mu   sync.Mutex
	file *os.File
	path string
	size int64
}

func newRotatingLogWriter(path string) (*RotatingLogWriter, error) {
	if path == "" {
		path = filepath.Join("logs", "kashi.log")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	return &RotatingLogWriter{file: file, path: path, size: info.Size()}, nil
}

func (writer *RotatingLogWriter) Write(data []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	if writer.size+int64(len(data)) > maxAppLogBytes {
		if err := writer.rotate(); err != nil {
			return 0, err
		}
	}
	written, err := writer.file.Write(data)
	writer.size += int64(written)
	return written, err
}

func (writer *RotatingLogWriter) rotate() error {
	if err := writer.file.Close(); err != nil {
		return err
	}
	_ = os.Remove(writer.path + ".1")
	if err := os.Rename(writer.path, writer.path+".1"); err != nil && !os.IsNotExist(err) {
		return err
	}
	file, err := os.OpenFile(writer.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	writer.file = file
	writer.size = 0
	return nil
}

func (writer *RotatingLogWriter) Close() error {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	return writer.file.Close()
}

func (writer *RotatingLogWriter) Path() string {
	return writer.path
}

// ConfigureAppLogging retains normal console output and mirrors all standard
// Go log messages to the bounded rotating file.
func ConfigureAppLogging(path string) (*RotatingLogWriter, error) {
	writer, err := newRotatingLogWriter(path)
	if err != nil {
		return nil, err
	}
	log.SetOutput(io.MultiWriter(os.Stderr, writer))
	log.Printf("logging to %s", writer.Path())
	return writer, nil
}
