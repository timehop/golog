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

type LogLevel int

const (
	LevelFatal LogLevel = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

type LogLevelName string

const (
	LevelFatalName LogLevelName = "FATAL"
	LevelErrorName              = "ERROR"
	LevelWarnName               = "WARN"
	LevelInfoName               = "INFO"
	LevelDebugName              = "DEBUG"
	LevelTraceName              = "TRACE"
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
	Level LogLevel
	Flags int

	DefaultLogger Logger

	defaultPrefix string
	defaultOutput io.Writer
)

func init() {
	// Call a separate func to actually do anything so it can be tested. This
	// function itself automatically gets called on package load.
	initLogging()

	// Setting up mockable fns is fine to do here, though.
	osExit = os.Exit
}

func initLogging() {
	defaultPrefix = os.Getenv("LOG_PREFIX")
	defaultOutput = os.Stdout

	Level = LevelInfo
	switch LogLevelName(os.Getenv("LOG_LEVEL")) {
	case LevelFatalName:
		Level = LevelFatal
	case LevelErrorName:
		Level = LevelError
	case LevelWarnName:
		Level = LevelWarn
	case LevelDebugName:
		Level = LevelDebug
	case LevelTraceName:
		Level = LevelTrace
	}

	var err error
	Flags, err = strconv.Atoi(os.Getenv("LOG_FORMAT"))
	if err != nil {
		Flags = FlagsDefault
	}

	DefaultLogger = NewDefault()
}

// Changes the global prefix for all log statements.
//
// New logger instances created after this method is called will be affected.
// Prefix is useful for multi-tail scenarios (tailing logs across multiple
// machines, to help distinguish which is which.)
func SetPrefix(prefix string) {
	defaultPrefix = prefix
	// Must recreate the default logger so it can pickup the prefix.
	DefaultLogger = NewDefault()
}

// Fatal outputs a severe error message just before terminating the process.
// Use judiciously.
func Fatal(id, description string, keysAndValues ...interface{}) {
	if Level < LevelFatal {
		return
	}
	DefaultLogger.logMessage(LevelFatalName, id, description, keysAndValues...)
	osExit(1)
}

// Error outputs an error message with an optional list of key/value pairs.
func Error(id, description string, keysAndValues ...interface{}) {
	if Level < LevelError {
		return
	}
	DefaultLogger.logMessage(LevelErrorName, id, description, keysAndValues...)
}

// Warn outputs a warning message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelWarn, calling this method will yield no
// side effects.
func Warn(id, description string, keysAndValues ...interface{}) {
	if Level < LevelWarn {
		return
	}
	DefaultLogger.logMessage(LevelWarnName, id, description, keysAndValues...)
}

// Info outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelInfo, calling this method will yield no
// side effects.
func Info(id, description string, keysAndValues ...interface{}) {
	if Level < LevelInfo {
		return
	}
	DefaultLogger.logMessage(LevelInfoName, id, description, keysAndValues...)
}

// Debug outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelDebug, calling this method will yield no
// side effects.
func Debug(id, description string, keysAndValues ...interface{}) {
	if Level < LevelDebug {
		return
	}
	DefaultLogger.logMessage(LevelDebugName, id, description, keysAndValues...)
}

// Trace outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelTrace, calling this method will yield no
// side effects.
func Trace(id, description string, keysAndValues ...interface{}) {
	if Level < LevelTrace {
		return
	}
	DefaultLogger.logMessage(LevelTraceName, id, description, keysAndValues...)
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

type Logger interface {
	Fatal(description string, keysAndValues ...interface{})
	Error(description string, keysAndValues ...interface{})
	Warn(description string, keysAndValues ...interface{})
	Info(description string, keysAndValues ...interface{})
	Debug(description string, keysAndValues ...interface{})
	Trace(description string, keysAndValues ...interface{})

	SetLevel(level LogLevel)
	SetOutput(w io.Writer)
	SetTimestampFlags(flags int)
	SetStaticField(name string, value interface{})

	logMessage(level LogLevelName, id string, description string, keysAndValues ...interface{})
}

// Logger config. Default/unset values for each attribute are safe.
type Config struct {
	Format LogFormat
	ID     string
}

type LogFormat string

const (
	DefaultFormat   LogFormat = "" // Use env variable, defaulting to PlainTextFormat
	PlainTextFormat LogFormat = "text"
	JsonFormat      LogFormat = "json"
)

// New creates a new logger instance.
func New(conf Config, staticKeysAndValues ...interface{}) Logger {
	var prefix string
	var flags int
	var formatter formatLogEvent
	staticArgs := make(map[string]string, 0)

	format := SanitizeFormat(conf.Format)

	if format == JsonFormat {
		formatter = formatLogEventAsJson

		// Don't mess up the json by letting logger print these:
		prefix = ""
		flags = 0

		// Instead put them into the staticArgs
		if defaultPrefix != "" {
			staticArgs["prefix"] = defaultPrefix
		}
	} else {
		formatter = formatLogEvent(formatLogEventAsPlainText)
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

	return &logger{
		id:    conf.ID,
		level: Level,

		formatLogEvent: formatter,
		staticArgs:     staticArgs,

		// don't touch the default logger on 'log' package
		// cache args to make a logger, in case it's changes with SetOutput()
		prefix: prefix,
		flags:  flags,
		l:      log.New(defaultOutput, prefix, flags),
	}
}

func NewDefault() Logger {
	return New(Config{})
}

func SanitizeFormat(format LogFormat) LogFormat {
	if format == PlainTextFormat || format == JsonFormat {
		return format
	} else {
		// Whether it's explicitly a DefaultFormat, or it's an unrecognized value,
		// try to take from env var.
		envFormat := os.Getenv("LOG_ENCODING")
		if envFormat == string(JsonFormat) || envFormat == string(PlainTextFormat) {
			return LogFormat(envFormat)
		}
	}

	// Fall back to text
	return PlainTextFormat
}

// Logger represents a logger, through which output is generated.
//
// It holds an ID, the minimum severity level to generate output (all calls
// with inferior severity will yield no effect) and wraps the underlying
// logger, which is a standard lib's *log.Logger instance.
type logger struct {
	id    string
	level LogLevel

	formatLogEvent formatLogEvent
	staticArgs     map[string]string

	prefix string
	flags  int
	l      *log.Logger
}

// Fatal outputs an error message with an optional list of key/value pairs and exits
func (s *logger) Fatal(description string, keysAndValues ...interface{}) {
	if s.level < LevelFatal {
		return
	}
	s.logMessage(LevelFatalName, s.id, description, keysAndValues...)
	osExit(1)
}

// Error outputs an error message with an optional list of key/value pairs.
func (s *logger) Error(description string, keysAndValues ...interface{}) {
	if s.level < LevelError {
		return
	}
	s.logMessage(LevelErrorName, s.id, description, keysAndValues...)
}

// Warn outputs a warning message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelWarn, calling this method will yield no
// side effects.
func (s *logger) Warn(description string, keysAndValues ...interface{}) {
	if s.level < LevelWarn {
		return
	}
	s.logMessage(LevelWarnName, s.id, description, keysAndValues...)
}

// Info outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelInfo, calling this method will yield no
// side effects.
func (s *logger) Info(description string, keysAndValues ...interface{}) {
	if s.level < LevelInfo {
		return
	}
	s.logMessage(LevelInfoName, s.id, description, keysAndValues...)
}

// Debug outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelDebug, calling this method will yield no
// side effects.
func (s *logger) Debug(description string, keysAndValues ...interface{}) {
	if s.level < LevelDebug {
		return
	}
	s.logMessage(LevelDebugName, s.id, description, keysAndValues...)
}

// Trace outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelTrace, calling this method will yield no
// side effects.
func (s *logger) Trace(description string, keysAndValues ...interface{}) {
	if s.level < LevelTrace {
		return
	}
	s.logMessage(LevelTraceName, s.id, description, keysAndValues...)
}

func (s *logger) logMessage(level LogLevelName, id string, description string, keysAndValues ...interface{}) {
	msg := s.formatLogEvent(s.flags, id, level, description, s.staticArgs, keysAndValues...)
	s.l.Println(msg)
}

func (s *logger) SetLevel(level LogLevel) {
	s.level = level
}

// SetOutput sets the output destination for the logger.
//
// Useful to change where the log stream ends up being written to.
func (s *logger) SetOutput(w io.Writer) {
	s.l = log.New(w, s.prefix, s.flags)
}

// SetFlags changes the timestamp flags on the output of the logger.
func (s *logger) SetTimestampFlags(flags int) {
	s.flags = flags
	s.l.SetFlags(flags)
}

// Add a key/value field to every log line from this logger.
func (s *logger) SetStaticField(name string, value interface{}) {
	s.staticArgs[name] = fmt.Sprintf("%v", value)
}

type formatLogEvent func(
	flags int,
	id string,
	level LogLevelName,
	description string,
	staticFields map[string]string,
	extraFieldKeysAndValues ...interface{},
) string

// Format is "SEVERITY | Description [| k1='v1' k2='v2' k3=]"
// with key/value pairs being optional, depending on whether args are provided
func formatLogEventAsPlainText(flags int, id string, level LogLevelName, description string, staticFields map[string]string, args ...interface{}) string {
	// A full log statement is <id> | <severity> | <description> | <keys and values>
	items := make([]string, 0, 8)

	// If there are flags, go's logger will prefix with stuff, so add an empty
	// initial item as a placeholder, so string join will prefix a separator.
	if flags > FlagsNone {
		items = append(items, "")
	}

	items = append(items, string(level))

	if id != "" {
		items = append(items, id)
	}

	items = append(items, description)

	if len(args)+len(staticFields) > 0 {
		// Prefix with static fields.
		for key, value := range staticFields {
			args = append([]interface{}{key, value}, args...)
		}

		items = append(items, expandKeyValuePairs(args))
	}

	return strings.Join(items, " | ")
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

func formatLogEventAsJson(flags int, name string, level LogLevelName, msg string, staticFields map[string]string, extraFields ...interface{}) string {
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
		return fmt.Sprintf("{\"ts\": %v, \"msg\": \"failed to marshal log entry\"}", entry.Timestamp)
	}

	return string(encodedEntry)
}

type jsonLogEntry struct {
	Timestamp string            `json:"ts"`
	Level     LogLevelName      `json:"lvl"`
	Name      string            `json:"name,omitempty"`
	Message   string            `json:"msg,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

var osExit func(int)
