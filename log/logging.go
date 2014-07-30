package log

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
)

const (
	LevelFatal = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)

var Level = func() int {
	switch os.Getenv("LOG_LEVEL") {
	case "FATAL":
		return LevelFatal
	case "ERROR":
		return LevelError
	case "WARN":
		return LevelWarn
	case "DEBUG":
		return LevelDebug
	default:
		return LevelInfo
	}
}()

var (
	DefaultLogger = New()

	defaultOutput io.Writer = os.Stdout
)

func New() *Logger {
	return NewWithID("")
}

func NewWithID(id string) *Logger {
	return &Logger{
		ID:    id,
		Level: Level,                         // grab default
		l:     log.New(defaultOutput, "", 0), // don't touch the default logger on 'log' package
	}
}

func Fatal(id, description string, keysAndValues ...interface{}) {
	if Level < LevelFatal {
		return
	}
	logMessage(DefaultLogger.l, id, "FATAL", description, keysAndValues...)
}

// Error outputs an error message with an optional list of key/value pairs.
func Error(id, description string, keysAndValues ...interface{}) {
	if Level < LevelError {
		return
	}
	logMessage(DefaultLogger.l, id, "ERROR", description, keysAndValues...)
}

// Warn outputs a warning message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelWarn, calling this method will yield no
// side effects.
func Warn(id, description string, keysAndValues ...interface{}) {
	if Level < LevelWarn {
		return
	}
	logMessage(DefaultLogger.l, id, "WARN ", description, keysAndValues...)
}

// Info outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelInfo, calling this method will yield no
// side effects.
func Info(id, description string, keysAndValues ...interface{}) {
	if Level < LevelInfo {
		return
	}
	logMessage(DefaultLogger.l, id, "INFO ", description, keysAndValues...)
}

// Debug outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelDebug, calling this method will yield no
// side effects.
func Debug(id, description string, keysAndValues ...interface{}) {
	if Level < LevelDebug {
		return
	}
	logMessage(DefaultLogger.l, id, "DEBUG", description, keysAndValues...)
}

// SetOutput sets the output destination for the default logger.
//
// All new logger instances created after this call will use the provided
// io.Writer as destination for their output.
//
// If you specifically want to change the output of DefaultLogger and not
// affect new Logger instance creation, use log.DefaultLogger.SetOutput()
func SetOutput(w io.Writer) {
	defaultOutput = w
	DefaultLogger.SetOutput(w)
}

type Logger struct {
	ID    string
	Level int

	l *log.Logger
}

// Fatal outputs an error message with an optional list of key/value pairs and exits
func (s *Logger) Fatal(description string, keysAndValues ...interface{}) {
	if s.Level < LevelFatal {
		return
	}
	logMessage(s.l, s.ID, "FATAL", description, keysAndValues...)
	os.Exit(1)
}

// Error outputs an error message with an optional list of key/value pairs.
func (s *Logger) Error(description string, keysAndValues ...interface{}) {
	if s.Level < LevelError {
		return
	}
	logMessage(s.l, s.ID, "ERROR", description, keysAndValues...)
}

// Warn outputs a warning message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelWarn, calling this method will yield no
// side effects.
func (s *Logger) Warn(description string, keysAndValues ...interface{}) {
	if s.Level < LevelWarn {
		return
	}
	logMessage(s.l, s.ID, "WARN ", description, keysAndValues...)
}

// Info outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelInfo, calling this method will yield no
// side effects.
func (s *Logger) Info(description string, keysAndValues ...interface{}) {
	if s.Level < LevelInfo {
		return
	}
	logMessage(s.l, s.ID, "INFO ", description, keysAndValues...)
}

// Debug outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelDebug, calling this method will yield no
// side effects.
func (s *Logger) Debug(description string, keysAndValues ...interface{}) {
	if s.Level < LevelDebug {
		return
	}
	logMessage(s.l, s.ID, "DEBUG", description, keysAndValues...)
}

// SetOutput sets the output destination for the logger.
//
// Useful to change where the log stream ends up being written to.
func (s *Logger) SetOutput(w io.Writer) {
	s.l = log.New(w, "", 0)
}

// logMessage writes a formatted message to the default logger.
//
// Format is "SEVERITY | Description [| k1='v1' k2='v2' k3=]"
// with key/value pairs being optional, depending on whether args are provided
func logMessage(logger *log.Logger, id, severity, description string, args ...interface{}) {
	// A full log statement is <id> | <severity> | <description> | <keys and values>
	items := make([]interface{}, 0, 7)
	items = append(items, severity)
	items = append(items, "|")

	if len(id) > 0 {
		items = append(items, id)
		items = append(items, "|")
	}

	items = append(items, description)

	if len(args) > 0 {
		keysAndValues := expandKeyValuePairs(args)
		items = append(items, "|")
		items = append(items, keysAndValues)
	}

	logger.Println(items...)
}

// expandKeyValuePairs converts a list of arguments into a string with
// the format "k='v' foo='bar' bar=".
//
// When the final value is missing, the format "bar=" is used.
func expandKeyValuePairs(keyValuePairs []interface{}) string {
	argCount := len(keyValuePairs)

	kvPairCount := int(math.Ceil(float64(argCount) / 2)) // math, y u do dis.
	kvPairs := make([]string, kvPairCount)
	for i := 0; i < kvPairCount; i++ {
		keyIndex := i * 2
		valueIndex := keyIndex + 1
		key := keyValuePairs[keyIndex]
		if valueIndex < argCount {
			value := keyValuePairs[valueIndex]
			kvPairs[i] = fmt.Sprintf("%v='%v'", key, value)
		} else {
			kvPairs[i] = fmt.Sprintf("%v=", key)
		}
	}

	return strings.Join(kvPairs, " ")
}
