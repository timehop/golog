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
	"os"
	"strconv"
	"strings"
	"time"
)

type LogLevel int

const (
	LevelFatal LogLevel = 0
	LevelError LogLevel = 1
	LevelWarn  LogLevel = 2
	LevelInfo  LogLevel = 3
	LevelDebug LogLevel = 4
	LevelTrace LogLevel = 5
)

type LogLevelName string

const (
	LevelFatalName LogLevelName = "FATAL"
	LevelErrorName LogLevelName = "ERROR"
	LevelWarnName  LogLevelName = "WARN"
	LevelInfoName  LogLevelName = "INFO"
	LevelDebugName LogLevelName = "DEBUG"
	LevelTraceName LogLevelName = "TRACE"
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
	DefaultLogger Logger

	defaultPrefix string
	defaultOutput io.Writer
	defaultLevel  LogLevel
	defaultFlags  int
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

	defaultLevel = LevelInfo
	switch LogLevelName(os.Getenv("LOG_LEVEL")) {
	case LevelFatalName:
		defaultLevel = LevelFatal
	case LevelErrorName:
		defaultLevel = LevelError
	case LevelWarnName:
		defaultLevel = LevelWarn
	case LevelDebugName:
		defaultLevel = LevelDebug
	case LevelTraceName:
		defaultLevel = LevelTrace
	}

	if flags, err := strconv.Atoi(os.Getenv("LOG_FORMAT")); err != nil {
		defaultFlags = FlagsDefault
	} else {
		defaultFlags = flags
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
	keysAndValues = append([]interface{}{"golog_id", id}, keysAndValues...)
	DefaultLogger.Fatal(description, keysAndValues...)
}

// Error outputs an error message with an optional list of key/value pairs.
func Error(id, description string, keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"golog_id", id}, keysAndValues...)
	DefaultLogger.Error(description, keysAndValues...)
}

// Warn outputs a warning message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelWarn, calling this method will yield no
// side effects.
func Warn(id, description string, keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"golog_id", id}, keysAndValues...)
	DefaultLogger.Warn(description, keysAndValues...)
}

// Info outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelInfo, calling this method will yield no
// side effects.
func Info(id, description string, keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"golog_id", id}, keysAndValues...)
	DefaultLogger.Info(description, keysAndValues...)
}

// Debug outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelDebug, calling this method will yield no
// side effects.
func Debug(id, description string, keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"golog_id", id}, keysAndValues...)
	DefaultLogger.Debug(description, keysAndValues...)
}

// Trace outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelTrace, calling this method will yield no
// side effects.
func Trace(id, description string, keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"golog_id", id}, keysAndValues...)
	DefaultLogger.Trace(description, keysAndValues...)
}

func SetLevel(level LogLevel) {
	defaultLevel = level
	DefaultLogger.SetLevel(level)
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
	defaultFlags = flags
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
		flags = defaultFlags
	}

	// Set 'ID' config as a static field, but before reading the varargs suplied
	// fields, so that they can override the config.
	if conf.ID != "" {
		staticArgs["golog_id"] = conf.ID
	}

	if len(staticKeysAndValues)%2 == 1 {
		// If there are an odd number of staticKeysAndValue, then there's probably one
		// missing, which means we'd interpret a value as a key, which can be bad for
		// logs-as-data, like metrics on staticKeys or elasticsearch. But, instead of
		// throwing the corrupt data out, serialize it into a string, which both
		// keeps the info, and maintains key-value integrity.
		staticKeysAndValues = []interface{}{"corruptStaticFields", flattenKeyValues(staticKeysAndValues)}
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

	return &logger{
		level: defaultLevel,

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
	s.logMessage(LevelFatalName, description, keysAndValues...)
	osExit(1)
}

// Error outputs an error message with an optional list of key/value pairs.
func (s *logger) Error(description string, keysAndValues ...interface{}) {
	if s.level < LevelError {
		return
	}
	s.logMessage(LevelErrorName, description, keysAndValues...)
}

// Warn outputs a warning message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelWarn, calling this method will yield no
// side effects.
func (s *logger) Warn(description string, keysAndValues ...interface{}) {
	if s.level < LevelWarn {
		return
	}
	s.logMessage(LevelWarnName, description, keysAndValues...)
}

// Info outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelInfo, calling this method will yield no
// side effects.
func (s *logger) Info(description string, keysAndValues ...interface{}) {
	if s.level < LevelInfo {
		return
	}
	s.logMessage(LevelInfoName, description, keysAndValues...)
}

// Debug outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelDebug, calling this method will yield no
// side effects.
func (s *logger) Debug(description string, keysAndValues ...interface{}) {
	if s.level < LevelDebug {
		return
	}
	s.logMessage(LevelDebugName, description, keysAndValues...)
}

// Trace outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelTrace, calling this method will yield no
// side effects.
func (s *logger) Trace(description string, keysAndValues ...interface{}) {
	if s.level < LevelTrace {
		return
	}
	s.logMessage(LevelTraceName, description, keysAndValues...)
}

func (s *logger) logMessage(level LogLevelName, description string, keysAndValues ...interface{}) {
	// If there are an odd number of keysAndValue, then there's probably one
	// missing, which means we'd interpret a value as a key, which can be bad for
	// logs-as-data, like metrics on keys or elasticsearch. But, instead of
	// throwing the corrupt data out, serialize it into a string, which both
	// keeps the info, and maintains key-value integrity.
	if len(keysAndValues)%2 == 1 {
		// But, before checking for corrupt keys, remove golog_id, if present, cuz
		// that's an auto-field, so don't let user's missuse of keysAndValues mess up
		// the ID, which they didn't do incorrectly.
		if len(keysAndValues) >= 2 && keysAndValues[0] == "golog_id" {
			keysAndValues = []interface{}{
				"golog_id", keysAndValues[1],
				"corruptFields", flattenKeyValues(keysAndValues[2:]),
			}
		} else {
			keysAndValues = []interface{}{"corruptFields", flattenKeyValues(keysAndValues)}
		}
	}

	msg := s.formatLogEvent(s.flags, level, description, s.staticArgs, keysAndValues...)
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
	level LogLevelName,
	description string,
	staticFields map[string]string,
	extraFieldKeysAndValues ...interface{},
) string

// Format is "SEVERITY | Description [| k1='v1' k2='v2' k3=]"
// with key/value pairs being optional, depending on whether args are provided
func formatLogEventAsPlainText(flags int, level LogLevelName, description string, staticFields map[string]string, args ...interface{}) string {
	// A full log statement is <id> | <severity> | <description> | <keys and values>
	items := make([]string, 0, 8)

	// If there are flags, go's logger will prefix with stuff, so add an empty
	// initial item as a placeholder, so string join will prefix a separator.
	if flags > FlagsNone {
		items = append(items, "")
	}

	items = append(items, string(level))

	// Combine args and staticFields, allowing args to override staticFields.
	// But don't use yet, just use it for ID first.
	if len(args)+len(staticFields) > 0 {
		// Prefix with static fields, but make sure to allow args to override static.
		for key, value := range staticFields {
			var existsInArgs bool

			for i, arg := range args {
				if i%2 == 0 && key == arg {
					existsInArgs = true
				}
			}

			if !existsInArgs {
				args = append([]interface{}{key, value}, args...)
			}
		}
	}

	// Grab ID from args.
	var id string
	for i, arg := range args {
		if i%2 == 0 && fmt.Sprintf("%v", arg) == "golog_id" && i < len(args)-1 {
			// Set id and remove from fields
			id = fmt.Sprintf("%v", args[i+1])
			args = append(args[:i], args[i+2:]...)
			break
		}
	}
	if id != "" {
		items = append(items, id)
	}

	items = append(items, description)

	if len(args) > 0 {
		items = append(items, expandKeyValuePairs(args))
	}

	return strings.Join(items, " | ")
}

// expandKeyValuePairs converts a list of arguments into a string with the
// format "k='v' foo='bar' bar=".
func expandKeyValuePairs(keyValuePairs []interface{}) string {
	kvPairs := make([]string, 0, len(keyValuePairs)/2)

	// Just ignore the last dangling kv if odd #, cuz bug.
	for i, kv := range keyValuePairs {
		if i%2 == 1 {
			kvPairs = append(kvPairs, fmt.Sprintf("%v='%v'", keyValuePairs[i-1], kv))
		}
	}

	return strings.Join(kvPairs, " ")
}

func formatLogEventAsJson(flags int, level LogLevelName, msg string, staticFields map[string]string, extraFields ...interface{}) string {
	entry := jsonLogEntry{
		Timestamp: time.Now().String(),
		Level:     level,
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

	// log entry can't fail to marshal, it's just strings, so ignore error for 100% test coverage
	encodedEntry, _ := json.Marshal(entry)

	return string(encodedEntry)
}

type jsonLogEntry struct {
	Timestamp string            `json:"ts"`
	Level     LogLevelName      `json:"lvl"`
	Message   string            `json:"msg,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

func flattenKeyValues(keysAndValues []interface{}) string {
	stringKVs := make([]string, len(keysAndValues))
	for i, kv := range keysAndValues {
		stringKVs[i] = fmt.Sprintf("%v", kv)
	}

	return strings.Join(stringKVs, ", ")
}

var osExit func(int)
