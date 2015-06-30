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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
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
	if DefaultLogger.format == JsonFormat {
		logMessageInJson(DefaultLogger.l, id, "FATAL", description, nil, keysAndValues...)
	} else {
		logMessage(DefaultLogger.l, id, "FATAL", description, nil, keysAndValues...)
	}
	os.Exit(1)
}

// Error outputs an error message with an optional list of key/value pairs.
func Error(id, description string, keysAndValues ...interface{}) {
	if Level < LevelError {
		return
	}
	if DefaultLogger.format == JsonFormat {
		logMessageInJson(DefaultLogger.l, id, "ERROR", description, nil, keysAndValues...)
	} else {
		logMessage(DefaultLogger.l, id, "ERROR", description, nil, keysAndValues...)
	}
}

// Warn outputs a warning message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelWarn, calling this method will yield no
// side effects.
func Warn(id, description string, keysAndValues ...interface{}) {
	if Level < LevelWarn {
		return
	}
	if DefaultLogger.format == JsonFormat {
		logMessageInJson(DefaultLogger.l, id, "WARN", description, nil, keysAndValues...)
	} else {
		logMessage(DefaultLogger.l, id, "WARN ", description, nil, keysAndValues...)
	}
}

// Info outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelInfo, calling this method will yield no
// side effects.
func Info(id, description string, keysAndValues ...interface{}) {
	if Level < LevelInfo {
		return
	}
	if DefaultLogger.format == JsonFormat {
		logMessageInJson(DefaultLogger.l, id, "INFO", description, nil, keysAndValues...)
	} else {
		logMessage(DefaultLogger.l, id, "INFO ", description, nil, keysAndValues...)
	}
}

// Debug outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelDebug, calling this method will yield no
// side effects.
func Debug(id, description string, keysAndValues ...interface{}) {
	if Level < LevelDebug {
		return
	}
	if DefaultLogger.format == JsonFormat {
		logMessageInJson(DefaultLogger.l, id, "DEBUG", description, nil, keysAndValues...)
	} else {
		logMessage(DefaultLogger.l, id, "DEBUG", description, nil, keysAndValues...)
	}
}

// Trace outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelTrace, calling this method will yield no
// side effects.
func Trace(id, description string, keysAndValues ...interface{}) {
	if Level < LevelTrace {
		return
	}
	if DefaultLogger.format == JsonFormat {
		logMessageInJson(DefaultLogger.l, id, "TRACE", description, nil, keysAndValues...)
	} else {
		logMessage(DefaultLogger.l, id, "TRACE", description, nil, keysAndValues...)
	}
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

type LoggerInterface interface {
	Fatal(description string, keysAndValues ...interface{})
	Error(description string, keysAndValues ...interface{})
	Warn(description string, keysAndValues ...interface{})
	Info(description string, keysAndValues ...interface{})
	Debug(description string, keysAndValues ...interface{})
	Trace(description string, keysAndValues ...interface{})
	SetOutput(w io.Writer)
	SetTimestampFlags(flags int)
	SetField(name string, value interface{})
}

func NewLogger(format LogFormat, id string, staticKeysAndValues ...interface{}) LoggerInterface {
	return newLoggerStruct(format, id, staticKeysAndValues...)
}

func newLoggerStruct(format LogFormat, id string, staticKeysAndValues ...interface{}) *Logger {
	var prefix string
	var flags int
	staticArgs := make(map[string]string, 0)

	format = SanitizeFormat(format)

	if format == JsonFormat {
		// Don't mess up the json by letting logger print these:
		prefix = ""
		flags = 0

		// Instead put them into the staticArgs
		if defaultPrefix != "" {
			staticArgs["prefix"] = defaultPrefix
		}
	} else {
		prefix = defaultPrefix
		flags = Flags
	}

	// Do this after handling prefix, so that individual loggers can override
	// external env variable.
	currentKey := ""
	for i, arg := range staticKeysAndValues {
		if i%2 == 0 {
			currentKey = fmt.Sprintf("%v", arg)
		} else {
			staticArgs[currentKey] = fmt.Sprintf("%v", arg)
		}
	}

	// If there are an odd number of keys+values, add the dangling key with empty
	// value.
	if len(staticKeysAndValues)%2 == 1 {
		staticArgs[currentKey] = ""
	}

	return &Logger{
		ID:    id,
		Level: Level,

		format:     format,
		staticArgs: staticArgs,

		// don't touch the default logger on 'log' package
		l: log.New(defaultOutput, prefix, flags),
	}
}

type LogFormat string

const (
	DefaultFormat   LogFormat = "" // Use env variable, defaulting to PlainTextFormat
	PlainTextFormat           = "text"
	JsonFormat                = "json"
)

func SanitizeFormat(format LogFormat) LogFormat {
	if format == PlainTextFormat || format == JsonFormat {
		return format
	} else {
		// Whether it's explicitly a DefaultFormat, or it's an unrecognized value,
		// try to take from env var.
		envFormat := os.Getenv("DEFAULT_LOG_ENCODING_FORMAT")
		if envFormat == string(JsonFormat) || envFormat == string(PlainTextFormat) {
			return LogFormat(envFormat)
		}
	}

	// Fall back to text
	return PlainTextFormat
}

// New creates a new logger instance.
// DEPRECATED: use `NewLogger(...)` instead. That one returns an interface,
// which allows the underlying data structure to change without breaking
// clients.
func New() *Logger {
	return newLoggerStruct(DefaultFormat, "")
}

// NewWithID creates a new logger instance that will output use the supplied id
// as prefix for all the log messages.
// The format is:
//   Level | Prefix | Message | key='value' key2=value2, ...
//
// DEPRECATED: use `NewLogger(...)` instead. That one returns an interface,
// which allows the underlying data structure to change without breaking
// clients.
func NewWithID(id string) *Logger {
	return newLoggerStruct(DefaultFormat, id)
}

// Logger represents a logger, through which output is generated.
//
// It holds an ID, the minimum severity level to generate output (all calls
// with inferior severity will yield no effect) and wraps the underlying
// logger, which is a standard lib's *log.Logger instance.
type Logger struct {
	ID    string
	Level int

	format     LogFormat
	staticArgs map[string]string

	l *log.Logger
}

// Fatal outputs an error message with an optional list of key/value pairs and exits
func (s *Logger) Fatal(description string, keysAndValues ...interface{}) {
	if s.Level < LevelFatal {
		return
	}
	if s.format == JsonFormat {
		logMessageInJson(s.l, s.ID, "FATAL", description, s.staticArgs, keysAndValues...)
	} else {
		logMessage(s.l, s.ID, "FATAL", description, s.staticArgs, keysAndValues...)
	}
	os.Exit(1)
}

// Error outputs an error message with an optional list of key/value pairs.
func (s *Logger) Error(description string, keysAndValues ...interface{}) {
	if s.Level < LevelError {
		return
	}
	if s.format == JsonFormat {
		logMessageInJson(s.l, s.ID, "ERROR", description, s.staticArgs, keysAndValues...)
	} else {
		logMessage(s.l, s.ID, "ERROR", description, s.staticArgs, keysAndValues...)
	}
}

// Warn outputs a warning message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelWarn, calling this method will yield no
// side effects.
func (s *Logger) Warn(description string, keysAndValues ...interface{}) {
	if s.Level < LevelWarn {
		return
	}
	if s.format == JsonFormat {
		logMessageInJson(s.l, s.ID, "WARN", description, s.staticArgs, keysAndValues...)
	} else {
		logMessage(s.l, s.ID, "WARN ", description, s.staticArgs, keysAndValues...)
	}
}

// Info outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelInfo, calling this method will yield no
// side effects.
func (s *Logger) Info(description string, keysAndValues ...interface{}) {
	if s.Level < LevelInfo {
		return
	}
	if s.format == JsonFormat {
		logMessageInJson(s.l, s.ID, "INFO", description, s.staticArgs, keysAndValues...)
	} else {
		logMessage(s.l, s.ID, "INFO ", description, s.staticArgs, keysAndValues...)
	}
}

// Debug outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelDebug, calling this method will yield no
// side effects.
func (s *Logger) Debug(description string, keysAndValues ...interface{}) {
	if s.Level < LevelDebug {
		return
	}
	if s.format == JsonFormat {
		logMessageInJson(s.l, s.ID, "DEBUG", description, s.staticArgs, keysAndValues...)
	} else {
		logMessage(s.l, s.ID, "DEBUG", description, s.staticArgs, keysAndValues...)
	}
}

// Trace outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelTrace, calling this method will yield no
// side effects.
func (s *Logger) Trace(description string, keysAndValues ...interface{}) {
	if s.Level < LevelTrace {
		return
	}
	if s.format == JsonFormat {
		logMessageInJson(s.l, s.ID, "TRACE", description, s.staticArgs, keysAndValues...)
	} else {
		logMessage(s.l, s.ID, "TRACE", description, s.staticArgs, keysAndValues...)
	}
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

func (s *Logger) SetField(name string, value interface{}) {
	s.staticArgs[name] = fmt.Sprintf("%v", value)
}

// logMessage writes a formatted message to the default logger.
//
// Format is "SEVERITY | Description [| k1='v1' k2='v2' k3=]"
// with key/value pairs being optional, depending on whether args are provided
func logMessage(logger *log.Logger, id, severity, description string, staticFields map[string]string, args ...interface{}) {
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

	if len(args)+len(staticFields) > 0 {
		// Prefix with static fields.
		for key, value := range staticFields {
			args = append([]interface{}{key, value}, args...)
		}

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

type jsonLogEntry struct {
	Timestamp string            `json:"ts"`
	Level     string            `json:"lvl"`
	Name      string            `json:"name,omitempty"`
	Message   string            `json:"msg,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

func logMessageInJson(logger *log.Logger, name, level, msg string, staticFields map[string]string, extraFields ...interface{}) {
	entry := jsonLogEntry{
		Timestamp: time.Now().String(),
		Level:     level,
		Name:      name,
		Message:   msg,
	}

	// If there are an odd number of keys+values, round up, cuz empty key will still be added.
	var numExtraKeyValuePairs int = (len(extraFields) + 1) / 2

	entry.Fields = make(map[string]string, len(staticFields)+numExtraKeyValuePairs)
	for key, value := range staticFields {
		entry.Fields[key] = value
	}

	currentKey := ""
	for i, field := range extraFields {
		if i%2 == 0 {
			currentKey = fmt.Sprintf("%v", field)
		} else {
			entry.Fields[currentKey] = fmt.Sprintf("%v", field)
		}
	}

	// If there are an odd number of keys+values, add empty key
	if len(extraFields)%2 == 1 {
		entry.Fields[currentKey] = ""
	}

	encodedEntry, err := json.Marshal(entry)
	if err != nil {
		logger.Printf("{\"ts\": %v, \"msg\": \"failed to marshal log entry\"}", entry.Timestamp)
	} else {
		logger.Println(string(encodedEntry))
	}
}
