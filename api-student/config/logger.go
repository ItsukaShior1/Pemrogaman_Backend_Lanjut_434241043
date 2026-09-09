package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LogEntry struct {
	Time       string `json:"time"`
	RequestID  string `json:"request_id"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	DurationMS int64  `json:"duration_ms"`
	IP         string `json:"ip"`
}

type AppLogger struct {
	mu      sync.Mutex
	writer  io.Writer
	console io.Writer
}

func NewAppLogger() (*AppLogger, error) {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		return nil, fmt.Errorf("gagal membuat folder logs: %w", err)
	}

	file, err := os.OpenFile(
		"logs/app.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o644,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka logs/app.log: %w", err)
	}

	console := io.Writer(os.Stdout)
	if os.Getenv("APP_MODE") == "release" {
		console = io.Discard
	}

	return &AppLogger{
		writer:  file,
		console: console,
	}, nil
}

func (l *AppLogger) WriteLog(e LogEntry) error {
	line, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("gagal encode log: %w", err)
	}
	line = append(line, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.console.Write(line); err != nil {
		log.Printf("gagal tulis log ke STDOUT: %v", err)
	}

	if err := l.rotateIfNeeded(); err != nil {
		log.Printf("rotasi log gagal: %v", err)
	}

	if _, err := l.writer.Write(line); err != nil {
		return fmt.Errorf("gagal tulis log ke file: %w", err)
	}
	return nil
}

func (l *AppLogger) rotateIfNeeded() error {
	const MaxLogSize int64 = 10 * 1024 * 1024 // 10 MB
	const MaxLogBackups = 5

	info, err := os.Stat("logs/app.log")
	if err != nil || info.Size() <= MaxLogSize {
		return nil
	}

	if closer, ok := l.writer.(io.Closer); ok {
		_ = closer.Close()
	}

	rotated := fmt.Sprintf("logs/app.log.%d", time.Now().Unix())
	if err := os.Rename("logs/app.log", rotated); err != nil {
		l.openFile()
		return err
	}

	l.pruneBackups(MaxLogBackups)
	l.openFile()
	return nil
}

func (l *AppLogger) openFile() {
	f, err := os.OpenFile(
		"logs/app.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o644,
	)
	if err != nil {
		log.Printf("gagal membuka logs/app.log: %v", err)
		return
	}
	l.writer = f
}

func (l *AppLogger) pruneBackups(maxBackups int) {
	entries, err := os.ReadDir("logs")
	if err != nil {
		return
	}

	type bk struct {
		name string
		unix int64
	}
	var backups []bk
	for _, e := range entries {
		if e.IsDir() || e.Name() == "app.log" {
			continue
		}
		var unix int64
		_, _ = fmt.Sscanf(e.Name(), "app.log.%d", &unix)
		backups = append(backups, bk{e.Name(), unix})
	}

	for i := 0; i < len(backups); i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[i].unix > backups[j].unix {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	for i := 0; i < len(backups)-maxBackups; i++ {
		_ = os.Remove("logs/" + backups[i].name)
	}
}
func (l *AppLogger) RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		rid, _ := c.Locals("requestid").(string)
		entry := LogEntry{
			Time:       time.Now().Format("2006-01-02T15:04:05Z07:00"),
			RequestID:  rid,
			Method:     c.Method(),
			Path:       c.Path(),
			Status:     c.Response().StatusCode(),
			DurationMS: time.Since(start).Milliseconds(),
			IP:         c.IP(),
		}

		if werr := l.WriteLog(entry); werr != nil {
			log.Printf("gagal menulis log: %v", werr)
		}
		return err
	}
}
