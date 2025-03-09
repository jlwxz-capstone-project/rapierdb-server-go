package log

import (
	go_log "log"
	"os"
	"testing"
)

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	Level       int
	debugLogger *go_log.Logger
	infoLogger  *go_log.Logger
	warnLogger  *go_log.Logger
	errorLogger *go_log.Logger
}

var Log *Logger

func init() {
	Log = &Logger{
		Level:       LevelDebug,
		debugLogger: go_log.New(os.Stdout, "\033[92m[DEBUG]\033[0m ", go_log.LstdFlags),
		infoLogger:  go_log.New(os.Stdout, "\033[34m[INFO]\033[0m  ", go_log.LstdFlags),
		warnLogger:  go_log.New(os.Stderr, "\033[33m[WARN]\033[0m  ", go_log.LstdFlags),
		errorLogger: go_log.New(os.Stderr, "\033[1;31m[ERROR]\033[0m ", go_log.LstdFlags),
	}
}

func SetLevel(level int) {
	Log.Level = level
}

func Debug(v ...any) {
	if Log.Level > LevelDebug {
		return
	}
	Log.debugLogger.Println(v...)
}

func Info(v ...any) {
	if Log.Level > LevelInfo {
		return
	}
	Log.infoLogger.Println(v...)
}

func Warn(v ...any) {
	if Log.Level > LevelWarn {
		return
	}
	Log.warnLogger.Println(v...)
}

func Error(v ...any) {
	if Log.Level > LevelError {
		return
	}
	Log.errorLogger.Println(v...)
}

func Debugf(format string, v ...any) {
	if Log.Level > LevelDebug {
		return
	}
	Log.debugLogger.Printf(format, v...)
}

func Infof(format string, v ...any) {
	if Log.Level > LevelInfo {
		return
	}
	Log.infoLogger.Printf(format, v...)
}

func Warnf(format string, v ...any) {
	if Log.Level > LevelWarn {
		return
	}
	Log.warnLogger.Printf(format, v...)
}

func Errorf(format string, v ...any) {
	if Log.Level > LevelError {
		return
	}
	Log.errorLogger.Printf(format, v...)
}

func Terrorf(t *testing.T, format string, v ...any) {
	Errorf(format, v...)
	t.Fail()
}
