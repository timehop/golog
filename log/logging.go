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

// Fatal outputs an error message with an optional list of key/value pairs and exits
func Fatal(description string, keysAndValues ...interface{}) {
	if Level < LevelFatal {
		return
	}
	logMessage("FATAL", description, keysAndValues...)
	os.Exit(1)
}

// Error outputs an error message with an optional list of key/value pairs.
func Error(description string, keysAndValues ...interface{}) {
	if Level < LevelError {
		return
	}
	logMessage("ERROR", description, keysAndValues...)
}

// Error outputs a warning message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelWarn, calling this method will yield no
// side effects.
func Warn(description string, keysAndValues ...interface{}) {
	if Level < LevelWarn {
		return
	}
	logMessage("WARN ", description, keysAndValues...)
}

// Error outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelInfo, calling this method will yield no
// side effects.
func Info(description string, keysAndValues ...interface{}) {
	if Level < LevelInfo {
		return
	}
	logMessage("INFO ", description, keysAndValues...)
}

// Error outputs an info message with an optional list of key/value pairs.
//
// If LogLevel is set below LevelDebug, calling this method will yield no
// side effects.
func Debug(description string, keysAndValues ...interface{}) {
	if Level < LevelDebug {
		return
	}
	logMessage("DEBUG", description, keysAndValues...)
}

// SetOutput sets the output destination for the logger.
//
// Useful to change where the log stream ends up being written to.
func SetOutput(w io.Writer) {
	logger = log.New(w, "", 0)
}

// Internal private logger (so we don't touch default logger on 'log' package)
var logger = log.New(os.Stdout, "", 0)

// logMessage writes a formatted message to the default logger.
//
// Format is "SEVERITY | Description [| k1='v1' k2='v2' k3=]"
// with key/value pairs being optional, depending on whether args are provided
func logMessage(severityPrefix string, description string, args ...interface{}) {
	if len(args) == 0 {
		logger.Println(severityPrefix, "|", description)
		return
	}

	keysAndValues := expandKeyValuePairs(args)
	logger.Println(severityPrefix, "|", description, "|", keysAndValues)
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
