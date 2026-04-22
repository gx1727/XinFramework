package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	levelNames = map[int]string{
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
	}
)

type dailyWriter struct {
	mu         sync.Mutex
	dir        string
	filePrefix string
	date       string
	file       *os.File
	writer     io.Writer
}

func newDailyWriter(dir string, filePrefix string) (*dailyWriter, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir failed: %w", err)
	}
	w := &dailyWriter{dir: dir, filePrefix: filePrefix}
	if err := w.rotate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *dailyWriter) rotate() error {
	today := time.Now().Format("2006-01-02")
	if w.file != nil && w.date == today {
		return nil
	}

	fileName := today + ".log"
	if w.filePrefix != "" {
		fileName = w.filePrefix + "-" + today + ".log"
	}
	path := filepath.Join(w.dir, fileName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file failed: %w", err)
	}

	if w.file != nil {
		w.file.Close()
	}

	w.file = f
	w.date = today
	w.writer = io.MultiWriter(os.Stdout, f)
	return nil
}

func (w *dailyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if w.date != today {
		if err := w.rotate(); err != nil {
			return os.Stderr.Write(p)
		}
	}

	return w.writer.Write(p)
}

func (w *dailyWriter) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		w.file.Close()
	}
}

type Logger struct {
	level  int
	logger *log.Logger
	writer *dailyWriter
}

var (
	std           *Logger
	stdDir        string
	stdLevel      int
	moduleLoggers = map[string]*Logger{}
	moduleMu      sync.Mutex
)

func Init(dir string, level string) {
	lvl := parseLevel(level)

	w, err := newDailyWriter(dir, "")
	if err != nil {
		log.Fatalf("logger init failed: %v", err)
	}

	std = &Logger{
		level:  lvl,
		logger: log.New(w, "[xin] ", log.LstdFlags),
		writer: w,
	}
	stdDir = dir
	stdLevel = lvl

	log.SetOutput(w)
}

func Module(filePrefix string) *Logger {
	if filePrefix == "" {
		return std
	}

	moduleMu.Lock()
	defer moduleMu.Unlock()

	if l, ok := moduleLoggers[filePrefix]; ok {
		return l
	}
	if std == nil {
		return nil
	}

	w, err := newDailyWriter(stdDir, filePrefix)
	if err != nil {
		log.Printf("module logger init failed(%s): %v", filePrefix, err)
		return std
	}

	l := &Logger{
		level:  stdLevel,
		logger: log.New(w, "[xin] ", log.LstdFlags),
		writer: w,
	}
	moduleLoggers[filePrefix] = l
	return l
}

func parseLevel(s string) int {
	switch s {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l *Logger) logf(level int, format string, args ...interface{}) {
	if l == nil {
		return
	}
	if level < l.level {
		return
	}
	msg := fmt.Sprintf("[%s] %s", levelNames[level], fmt.Sprintf(format, args...))
	l.logger.Output(3, msg)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf(LevelDebug, format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(LevelInfo, format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logf(LevelWarn, format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(LevelError, format, args...)
}

func Debugf(format string, args ...interface{}) {
	if std != nil {
		std.logf(LevelDebug, format, args...)
	}
}

func Infof(format string, args ...interface{}) {
	if std != nil {
		std.logf(LevelInfo, format, args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if std != nil {
		std.logf(LevelWarn, format, args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if std != nil {
		std.logf(LevelError, format, args...)
	}
}

func Debug(args ...interface{}) {
	if std != nil {
		std.logf(LevelDebug, "%s", fmt.Sprint(args...))
	}
}

func Info(args ...interface{}) {
	if std != nil {
		std.logf(LevelInfo, "%s", fmt.Sprint(args...))
	}
}

func Warn(args ...interface{}) {
	if std != nil {
		std.logf(LevelWarn, "%s", fmt.Sprint(args...))
	}
}

func Error(args ...interface{}) {
	if std != nil {
		std.logf(LevelError, "%s", fmt.Sprint(args...))
	}
}

func Close() {
	moduleMu.Lock()
	for k, l := range moduleLoggers {
		if l != nil && l.writer != nil {
			l.writer.Close()
		}
		delete(moduleLoggers, k)
	}
	moduleMu.Unlock()

	if std != nil && std.writer != nil {
		std.writer.Close()
	}
}
