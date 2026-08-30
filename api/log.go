package api

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultLogLimit = 200
	maxLogLimit     = 500
	maxLogReadBytes = 1 << 20
)

func (server *Server) listAppLogs(ctx *gin.Context) {
	limit := defaultLogLimit
	if value, err := strconv.Atoi(ctx.Query("limit")); err == nil && value > 0 {
		limit = min(value, maxLogLimit)
	}

	path := server.config.AppLogPath
	if path == "" {
		path = filepath.Join("logs", "kashi.log")
	}
	lines, err := readLastLines(path, limit)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			server.writeJSON(ctx, http.StatusOK, envelope{"logs": []string{}, "file": filepath.Base(path)})
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"logs": lines, "file": filepath.Base(path)})
}

// readLastLines reads at most 1 MB from the end of the current log. The API
// therefore has a fixed memory ceiling even if a manually supplied log file
// is unexpectedly large.
func readLastLines(path string, limit int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	start := max(int64(0), info.Size()-maxLogReadBytes)
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 256*1024)
	lines := make([]string, 0, limit)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if len(lines) == limit {
			copy(lines, lines[1:])
			lines[len(lines)-1] = line
		} else {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}
