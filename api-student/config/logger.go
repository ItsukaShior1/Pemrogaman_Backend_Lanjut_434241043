// Package config berisi inisialisasi logger aplikasi.
//
// Logger menulis ke dua tujuan sekaligus:
//   1. STDOUT — supaya log langsung terlihat saat develop dengan "go run"
//   2. File log — logs/app.log dengan rotasi otomatis berbasis ukuran
//
// Setiap baris log berformat JSON dan memuat field-field wajib:
//   time, request_id, method, path, status, duration_ms, ip
//
// Folder logs/ tidak di-commit ke git (lihat .gitignore).
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

// LogEntry adalah struktur satu baris log permintaan HTTP.
// Dibuat agar format JSON konsisten untuk parsing otomatis.
type LogEntry struct {
	Time       string `json:"time"`
	RequestID  string `json:"request_id"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	DurationMS int64  `json:"duration_ms"`
	IP         string `json:"ip"`
}

// AppLogger adalah logger aplikasi yang thread-safe (digunakan oleh
// beberapa goroutine secara bersamaan).
type AppLogger struct {
	mu      sync.Mutex
	writer  io.Writer
	console io.Writer
}

// NewAppLogger membuka file logs/app.log dan menggabungkannya dengan
// STDOUT (atau hanya file bila APP_MODE=release). Folder logs/ dibuat
// bila belum ada.
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

// WriteLog menulis satu entri log ke file (dengan rotasi) dan ke console.
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

// rotateIfNeeded memutar file logs/app.log bila ukurannya > MaxLogSize.
// Setelah rotasi, hanya MaxLogBackups file backup terbaru yang disimpan.
func (l *AppLogger) rotateIfNeeded() error {
	const MaxLogSize int64 = 10 * 1024 * 1024 // 10 MB
	const MaxLogBackups = 5

	info, err := os.Stat("logs/app.log")
	if err != nil || info.Size() <= MaxLogSize {
		return nil
	}

	// Tutup file writer lama lalu rename.
	if closer, ok := l.writer.(io.Closer); ok {
		_ = closer.Close()
	}

	rotated := fmt.Sprintf("logs/app.log.%d", time.Now().Unix())
	if err := os.Rename("logs/app.log", rotated); err != nil {
		// Buka ulang writer walau rotasi gagal agar log berikutnya masih jalan.
		l.openFile()
		return err
	}

	l.pruneBackups(MaxLogBackups)
	l.openFile()
	return nil
}

// openFile (re)open logs/app.log untuk writer baru.
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

// pruneBackups menghapus backup paling lama hingga tersisa maxBackups file.
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

	// Sort ascending (terlama dulu).
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

// RequestLogger adalah Fiber middleware yang mencatat setiap request
// dalam format JSON. Dipasang oleh route.Setup().
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
