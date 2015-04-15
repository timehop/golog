// The MIT License (MIT)
//
// Copyright (c) 2014 timehop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package log provids a severity-based key/value logging replacement for Go's
// standard logger.
//
// The output is a simple and clean logging format that strikes the perfect
// balance between being human readable and easy to write parsing tools for.
//
// Examples:
//   ERROR | MyLibrary | Could not connect to server. | url='http://timehop.com/' error='timed out'
//   INFO  | MyLibrary | Something happened.
package log

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	LevelFatal = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

const (
	FlagsNone          = 0
	FlagsDate          = log.Ldate
	FlagsTime          = log.Ltime
	FlagsPrecisionTime = log.Lmicroseconds
	FlagsLongFile      = log.Llongfile
	FlagsShortFile     = log.Lshortfile
	FlagsDefault       = log.LstdFlags
)

var (
	Level int
	Flags int

	DefaultLogger *Logger

	defaultPrefix string
	defaultOutput io.Writer
)

func init() {
	defaultPrefix = os.Getenv("LOG_PREFIX")
	defaultOutput = os.Stdout

	Level = LevelInfo
	switch os.Getenv("LOG_LEVEL") {
	case "FATAL":
		Level = LevelFatal
	case "ERROR":
		Level = LevelError
	case "WARN":
		Level = LevelWarn
	case "DEBUG":
		Level = LevelDebug
	case "TRACE":
		Level = LevelTrace
	}

	var err error
	Flags, err = strconv.Atoi(os.Getenv("LOG_FORMAT"))
	if err != nil {
		Flags = FlagsDefault
	}

	DefaultLogger = New()
}

// Changes the global prefix for all log statements.
//
// New logger instances created after this method is called will be affected.
// Prefix is useful for multi-tail scenarios (tailing logs across multiple
// machines, to help distinguish which is which.)
func SetPrefix(prefix string) {
	defaultPrefix = prefix
	// Must recreate the default logger so it can pickup the prefix.
	DefaultLogger = New()
}

// Fatal outputs a severe error message just before terminating the process.
// Use judiciously.
func Fatal(id, description string, keysAndValues ...interface{}) {
	if Level < LevelFatal {
		return
	}
	logMessage(DefaultLogger.l, id, "FATAL", description, keysAndValues...)
	os.Exit(1)
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

// Trace outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelTrace, calling this method will yield no
// side effects.
func Trace(id, description string, keysAndValues ...interface{}) {
	if Level < LevelTrace {
		return
	}
	logMessage(DefaultLogger.l, id, "TRACE", description, keysAndValues...)
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

// SetFlags changes the timestamp flags on the output of the default logger.
func SetTimestampFlags(flags int) {
	Flags = flags
	DefaultLogger.SetTimestampFlags(flags)
}

// New creates a new logger instance.
func New() *Logger {
	return NewWithID("")
}

// NewWithID creates a new logger instance that will output use the supplied id
// as prefix for all the log messages.
// The format is:
//   Level | Prefix | Message | key='value' key2=value2, ...
func NewWithID(id string) *Logger {
	return &Logger{
		ID:    id,
		Level: Level,
		// don't touch the default logger on 'log' package
		l: log.New(defaultOutput, defaultPrefix, Flags),
	}
}

// Logger represents a logger, through which output is generated.
//
// It holds an ID, the minimum severity level to generate output (all calls
// with inferior severity will yield no effect) and wraps the underlying
// logger, which is a standard lib's *log.Logger instance.
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

// Trace outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelTrace, calling this method will yield no
// side effects.
func (s *Logger) Trace(description string, keysAndValues ...interface{}) {
	if s.Level < LevelTrace {
		return
	}
	logMessage(s.l, s.ID, "TRACE", description, keysAndValues...)
}

// SetOutput sets the output destination for the logger.
//
// Useful to change where the log stream ends up being written to.
func (s *Logger) SetOutput(w io.Writer) {
	s.l = log.New(w, defaultPrefix, s.l.Flags())
}

// SetFlags changes the timestamp flags on the output of the logger.
func (s *Logger) SetTimestampFlags(flags int) {
	s.l.SetFlags(flags)
}

// logMessage writes a formatted message to the default logger.
//
// Format is "SEVERITY | Description [| k1='v1' k2='v2' k3=]"
// with key/value pairs being optional, depending on whether args are provided
func logMessage(logger *log.Logger, id, severity, description string, args ...interface{}) {
	// A full log statement is <id> | <severity> | <description> | <keys and values>
	items := make([]interface{}, 0, 8)
	if logger.Flags() > FlagsNone {
		items = append(items, "|")
	}

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

// expandKeyValuePairs converts a list of arguments into a string with the
// format "k='v' foo='bar' bar=".
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
