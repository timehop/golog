package log

import (
	"bytes"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"
)

func resetLogging(t *testing.T) {
	t.Helper()
	t.Setenv("LOG_PREFIX", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "0")
	t.Setenv("LOG_ENCODING", "")
	t.Setenv("LOG_STACK_TRACE", "false")

	initLogging()

	osExit = func(code int) {}
}

type exitCapture struct {
	didExit  bool
	exitCode int
}

func captureExit(t *testing.T) *exitCapture {
	t.Helper()
	ec := &exitCapture{}
	osExit = func(code int) {
		ec.didExit = true
		ec.exitCode = code
	}
	return ec
}

func TestJsonFormat(t *testing.T) {
	resetLogging(t)

	t.Run("logs a jsonLogEntry struct in json format", func(t *testing.T) {
		resetLogging(t)
		output := new(bytes.Buffer)
		SetOutput(output)

		timeBefore := time.Now()
		New(Config{Format: JsonFormat, ID: "id"}).Error("oh no")
		timeAfter := time.Now()

		var entry jsonLogEntry
		if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		t.Run("has a timestamp", func(t *testing.T) {
			start := strings.Index(entry.Timestamp, " m=+")
			chunk := []byte(entry.Timestamp)[:start]
			entryTimestamp := string(chunk)

			timestamp, err := time.Parse("2006-01-02 15:04:05 -0700 MST", entryTimestamp)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if timestamp.Before(timeBefore) {
				t.Errorf("timestamp %v is before %v", timestamp, timeBefore)
			}
			if timestamp.After(timeAfter) {
				t.Errorf("timestamp %v is after %v", timestamp, timeAfter)
			}
		})

		t.Run("has a log level", func(t *testing.T) {
			if entry.Level != LevelErrorName {
				t.Errorf("got %v, want %v", entry.Level, LevelErrorName)
			}
		})

		t.Run("has a message", func(t *testing.T) {
			if entry.Message != "oh no" {
				t.Errorf("got %q, want %q", entry.Message, "oh no")
			}
		})

		t.Run("has an id field", func(t *testing.T) {
			if entry.Fields["golog_id"] != "id" {
				t.Errorf("got %q, want %q", entry.Fields["golog_id"], "id")
			}
		})
	})

	t.Run("default prefix is set", func(t *testing.T) {
		resetLogging(t)
		t.Setenv("LOG_PREFIX", "default_prefix")
		initLogging()

		output := new(bytes.Buffer)
		SetOutput(output)
		New(Config{Format: JsonFormat}).Error("oh no")

		var entry jsonLogEntry
		if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.Fields["prefix"] != "default_prefix" {
			t.Errorf("got %q, want %q", entry.Fields["prefix"], "default_prefix")
		}
	})
}

func TestFields(t *testing.T) {
	resetLogging(t)

	t.Run("without static fields", func(t *testing.T) {
		t.Run("without dynamic fields", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: JsonFormat})
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no")

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(entry.Fields) != 0 {
				t.Errorf("got %d fields, want 0", len(entry.Fields))
			}
		})

		t.Run("with dynamic fields using JsonFormat", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: JsonFormat})
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no", "field", "value")

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Fields["field"] != "value" {
				t.Errorf("got %q, want %q", entry.Fields["field"], "value")
			}
		})

		t.Run("with dynamic fields using PlainTextFormat", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: PlainTextFormat})
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no", "field", "value")

			if !strings.Contains(output.String(), "field") {
				t.Errorf("output %q does not contain %q", output.String(), "field")
			}
			if !strings.Contains(output.String(), "value") {
				t.Errorf("output %q does not contain %q", output.String(), "value")
			}
		})
	})

	t.Run("with static fields", func(t *testing.T) {
		t.Run("without dynamic fields", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: JsonFormat}, "static_field", "static_value")
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no")

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Fields["static_field"] != "static_value" {
				t.Errorf("got %q, want %q", entry.Fields["static_field"], "static_value")
			}
			if len(entry.Fields) != 1 {
				t.Errorf("got %d fields, want 1", len(entry.Fields))
			}
		})

		t.Run("with different dynamic fields using JsonFormat", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: JsonFormat}, "static_field", "static_value")
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no", "field", "value")

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Fields["static_field"] != "static_value" {
				t.Errorf("got %q, want %q", entry.Fields["static_field"], "static_value")
			}
			if entry.Fields["field"] != "value" {
				t.Errorf("got %q, want %q", entry.Fields["field"], "value")
			}
			if len(entry.Fields) != 2 {
				t.Errorf("got %d fields, want 2", len(entry.Fields))
			}
		})

		t.Run("with different dynamic fields using PlainTextFormat", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: PlainTextFormat}, "static_field", "static_value")
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no", "dynamic_field", "dynamic_value")

			out := output.String()
			for _, s := range []string{"static_field", "static_value", "dynamic_field", "dynamic_value"} {
				if !strings.Contains(out, s) {
					t.Errorf("output %q does not contain %q", out, s)
				}
			}
		})

		t.Run("dynamic field overrides static using JsonFormat", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: JsonFormat}, "static_field", "static_value")
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no", "static_field", "value")

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Fields["static_field"] != "value" {
				t.Errorf("got %q, want %q", entry.Fields["static_field"], "value")
			}
			if len(entry.Fields) != 1 {
				t.Errorf("got %d fields, want 1", len(entry.Fields))
			}
		})

		t.Run("dynamic field overrides static using PlainTextFormat", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: PlainTextFormat}, "static_field", "static_value")
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no", "static_field", "value")

			out := output.String()
			if !strings.Contains(out, "static_field") {
				t.Errorf("output %q does not contain %q", out, "static_field")
			}
			if !strings.Contains(out, "value") {
				t.Errorf("output %q does not contain %q", out, "value")
			}
			if strings.Contains(out, "static_value") {
				t.Errorf("output %q should not contain %q", out, "static_value")
			}
		})
	})

	t.Run("odd number of key-value pairs", func(t *testing.T) {
		t.Run("in static fields", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: JsonFormat}, "static_field", "static_value", "odd_key")
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no", "key", "value")

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Fields["corruptStaticFields"] != "static_field, static_value, odd_key" {
				t.Errorf("got %q, want %q", entry.Fields["corruptStaticFields"], "static_field, static_value, odd_key")
			}
			if entry.Fields["key"] != "value" {
				t.Errorf("got %q, want %q", entry.Fields["key"], "value")
			}
			if len(entry.Fields) != 2 {
				t.Errorf("got %d fields, want 2", len(entry.Fields))
			}
		})

		t.Run("in dynamic fields", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: JsonFormat}, "static_field", "static_value")
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no", "field", "value", "odd_key")

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Fields["static_field"] != "static_value" {
				t.Errorf("got %q, want %q", entry.Fields["static_field"], "static_value")
			}
			if entry.Fields["corruptFields"] != "field, value, odd_key" {
				t.Errorf("got %q, want %q", entry.Fields["corruptFields"], "field, value, odd_key")
			}
			if len(entry.Fields) != 2 {
				t.Errorf("got %d fields, want 2", len(entry.Fields))
			}
		})

		t.Run("in both static and dynamic fields", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: JsonFormat}, "static_field", "static_value", "static_odd_key")
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no", "field", "value", "dynamic_odd_key")

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Fields["corruptStaticFields"] != "static_field, static_value, static_odd_key" {
				t.Errorf("got %q, want %q", entry.Fields["corruptStaticFields"], "static_field, static_value, static_odd_key")
			}
			if entry.Fields["corruptFields"] != "field, value, dynamic_odd_key" {
				t.Errorf("got %q, want %q", entry.Fields["corruptFields"], "field, value, dynamic_odd_key")
			}
			if len(entry.Fields) != 2 {
				t.Errorf("got %d fields, want 2", len(entry.Fields))
			}
		})
	})

	t.Run("setting static fields after creating logger", func(t *testing.T) {
		resetLogging(t)
		logger := New(Config{Format: JsonFormat}, "old_static_field", "old_static_value")
		output := new(bytes.Buffer)
		logger.SetOutput(output)
		logger.SetStaticField("new_static_field", "new_static_value")

		t.Run("uses the static fields when logging", func(t *testing.T) {
			logger.Error("oh no")

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Fields["old_static_field"] != "old_static_value" {
				t.Errorf("got %q, want %q", entry.Fields["old_static_field"], "old_static_value")
			}
			if entry.Fields["new_static_field"] != "new_static_value" {
				t.Errorf("got %q, want %q", entry.Fields["new_static_field"], "new_static_value")
			}
			if len(entry.Fields) != 2 {
				t.Errorf("got %d fields, want 2", len(entry.Fields))
			}
		})

		t.Run("overrides the newly set static fields when logging", func(t *testing.T) {
			output := new(bytes.Buffer)
			logger.SetOutput(output)

			logger.Error("oh no", "new_static_field", "dynamic_value")

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Fields["old_static_field"] != "old_static_value" {
				t.Errorf("got %q, want %q", entry.Fields["old_static_field"], "old_static_value")
			}
			if entry.Fields["new_static_field"] != "dynamic_value" {
				t.Errorf("got %q, want %q", entry.Fields["new_static_field"], "dynamic_value")
			}
			if len(entry.Fields) != 2 {
				t.Errorf("got %d fields, want 2", len(entry.Fields))
			}
		})
	})
}

func TestSanitizeFormat(t *testing.T) {
	resetLogging(t)

	t.Run("known format returns that format", func(t *testing.T) {
		if got := SanitizeFormat(PlainTextFormat); got != PlainTextFormat {
			t.Errorf("got %v, want %v", got, PlainTextFormat)
		}
		if got := SanitizeFormat(JsonFormat); got != JsonFormat {
			t.Errorf("got %v, want %v", got, JsonFormat)
		}
	})

	t.Run("unknown format with env default", func(t *testing.T) {
		t.Setenv("LOG_ENCODING", "json")
		initLogging()
		if got := SanitizeFormat(LogFormat("whut")); got != JsonFormat {
			t.Errorf("got %v, want %v", got, JsonFormat)
		}

		t.Setenv("LOG_ENCODING", "text")
		initLogging()
		if got := SanitizeFormat(LogFormat("whut")); got != PlainTextFormat {
			t.Errorf("got %v, want %v", got, PlainTextFormat)
		}
	})

	t.Run("unknown format without env default", func(t *testing.T) {
		t.Setenv("LOG_ENCODING", "")
		initLogging()
		if got := SanitizeFormat(LogFormat("whut")); got != PlainTextFormat {
			t.Errorf("got %v, want %v", got, PlainTextFormat)
		}
	})
}

func TestFatalLogging(t *testing.T) {
	resetLogging(t)

	t.Run("package level Fatal with level set low", func(t *testing.T) {
		resetLogging(t)
		SetLevel(LogLevel(-1))
		ec := captureExit(t)
		output := new(bytes.Buffer)
		SetOutput(output)

		Fatal("id", "msg")

		if output.String() != "" {
			t.Errorf("expected empty output, got %q", output.String())
		}
		if ec.didExit {
			t.Error("expected no exit")
		}
		if ec.exitCode != 0 {
			t.Errorf("got exit code %d, want 0", ec.exitCode)
		}
	})

	t.Run("logger Fatal with level set low", func(t *testing.T) {
		resetLogging(t)
		SetLevel(LogLevel(-1))
		logger := NewDefault()
		output := new(bytes.Buffer)
		logger.SetOutput(output)
		ec := captureExit(t)

		logger.Fatal("msg")

		if output.String() != "" {
			t.Errorf("expected empty output, got %q", output.String())
		}
		if ec.didExit {
			t.Error("expected no exit")
		}
		if ec.exitCode != 0 {
			t.Errorf("got exit code %d, want 0", ec.exitCode)
		}
	})
}

func TestSetPrefix(t *testing.T) {
	resetLogging(t)
	t.Setenv("LOG_PREFIX", "env_prefix ")
	initLogging()
	SetPrefix("fn_prefix ")

	t.Run("default logger uses new prefix", func(t *testing.T) {
		output := new(bytes.Buffer)
		SetOutput(output)
		Error("id", "msg")
		if got := output.String(); got != "fn_prefix ERROR | id | msg\n" {
			t.Errorf("got %q, want %q", got, "fn_prefix ERROR | id | msg\n")
		}
	})

	t.Run("new loggers use new prefix", func(t *testing.T) {
		logger := New(Config{ID: "id"})
		output := new(bytes.Buffer)
		logger.SetOutput(output)
		logger.Error("msg")
		if got := output.String(); got != "fn_prefix ERROR | id | msg\n" {
			t.Errorf("got %q, want %q", got, "fn_prefix ERROR | id | msg\n")
		}
	})
}

func TestInit(t *testing.T) {
	t.Run("with prefix from env var", func(t *testing.T) {
		resetLogging(t)
		t.Setenv("LOG_PREFIX", "env_prefix ")
		initLogging()
		output := new(bytes.Buffer)
		SetOutput(output)

		Error("id", "msg")
		if got := output.String(); got != "env_prefix ERROR | id | msg\n" {
			t.Errorf("got %q, want %q", got, "env_prefix ERROR | id | msg\n")
		}
	})

	t.Run("with format from env var", func(t *testing.T) {
		resetLogging(t)
		t.Setenv("LOG_FORMAT", strconv.FormatInt(log.Ldate|log.Lshortfile, 10))
		initLogging()
		output := new(bytes.Buffer)
		SetOutput(output)

		Error("id", "msg")
		// Check output matches expected regex pattern
		out := output.String()
		if !strings.Contains(out, "ERROR") || !strings.Contains(out, "id") || !strings.Contains(out, "msg") {
			t.Errorf("output %q missing expected content", out)
		}
	})

	t.Run("log levels", func(t *testing.T) {
		levels := []struct {
			name        string
			higherCount int
			lowerFns    func()
		}{
			{"FATAL", 1, func() { Error("id", "msg"); Warn("id", "msg"); Info("id", "msg"); Debug("id", "msg"); Trace("id", "msg") }},
			{"ERROR", 2, func() { Warn("id", "msg"); Info("id", "msg"); Debug("id", "msg"); Trace("id", "msg") }},
			{"WARN", 3, func() { Info("id", "msg"); Debug("id", "msg"); Trace("id", "msg") }},
			{"INFO", 4, func() { Debug("id", "msg"); Trace("id", "msg") }},
			{"DEBUG", 5, func() { Trace("id", "msg") }},
			{"TRACE", 6, nil},
		}

		for _, tc := range levels {
			t.Run(tc.name, func(t *testing.T) {
				resetLogging(t)
				t.Setenv("LOG_LEVEL", tc.name)
				initLogging()

				t.Run("logs higher levels", func(t *testing.T) {
					ec := captureExit(t)
					output := new(bytes.Buffer)
					SetOutput(output)

					Fatal("id", "msg")
					Error("id", "msg")
					Warn("id", "msg")
					Info("id", "msg")
					Debug("id", "msg")
					Trace("id", "msg")

					if output.String() == "" {
						t.Error("expected non-empty output")
					}
					if got := strings.Count(output.String(), "\n"); got != tc.higherCount {
						t.Errorf("got %d lines, want %d", got, tc.higherCount)
					}
					if !ec.didExit {
						t.Error("expected exit from Fatal")
					}
					if ec.exitCode <= 0 {
						t.Errorf("expected positive exit code, got %d", ec.exitCode)
					}
				})

				if tc.lowerFns != nil {
					t.Run("does not log lower levels", func(t *testing.T) {
						resetLogging(t)
						t.Setenv("LOG_LEVEL", tc.name)
						initLogging()
						output := new(bytes.Buffer)
						SetOutput(output)

						tc.lowerFns()

						if output.String() != "" {
							t.Errorf("expected empty output, got %q", output.String())
						}
					})
				}
			})
		}
	})
}

func TestLoggerLevels(t *testing.T) {
	resetLogging(t)

	type levelTest struct {
		name      string
		level     LogLevel
		logFn     func(Logger)
		wantLevel LogLevelName
	}

	tests := []levelTest{
		{"Fatal below level", LevelFatal, func(l Logger) { l.Fatal("oh no") }, LevelFatalName},
		{"Error below level", LevelError, func(l Logger) { l.Error("oh no") }, LevelErrorName},
		{"Warn below level", LevelWarn, func(l Logger) { l.Warn("oh no") }, LevelWarnName},
		{"Info below level", LevelInfo, func(l Logger) { l.Info("oh no") }, LevelInfoName},
		{"Debug below level", LevelDebug, func(l Logger) { l.Debug("oh no") }, LevelDebugName},
		{"Trace below level", LevelTrace, func(l Logger) { l.Trace("oh no") }, LevelTraceName},
	}

	for _, tc := range tests {
		t.Run(tc.name+" should log", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: JsonFormat})
			output := new(bytes.Buffer)
			logger.SetOutput(output)
			logger.SetLevel(tc.level)

			tc.logFn(logger)

			var entry jsonLogEntry
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Level != tc.wantLevel {
				t.Errorf("got %v, want %v", entry.Level, tc.wantLevel)
			}
		})

		t.Run(tc.name+" above level should not log", func(t *testing.T) {
			resetLogging(t)
			logger := New(Config{Format: JsonFormat})
			output := new(bytes.Buffer)
			logger.SetOutput(output)
			logger.SetLevel(LevelFatal)
			if tc.level == LevelFatal {
				logger.SetLevel(LogLevel(-1))
			}

			tc.logFn(logger)

			if output.Len() != 0 {
				t.Errorf("expected empty output, got %q", output.String())
			}
		})
	}
}

func TestPackageLevelLogging(t *testing.T) {
	setup := func(t *testing.T) *bytes.Buffer {
		t.Helper()
		resetLogging(t)
		output := new(bytes.Buffer)
		SetLevel(LevelTrace)
		SetTimestampFlags(FlagsNone)
		SetOutput(output)
		return output
	}

	t.Run("Fatal", func(t *testing.T) {
		t.Run("without ID", func(t *testing.T) {
			output := setup(t)
			captureExit(t)
			Fatal("", "Not all those who wander are lost.")
			if got := output.String(); got != "FATAL | Not all those who wander are lost.\n" {
				t.Errorf("got %q", got)
			}
		})

		t.Run("with ID", func(t *testing.T) {
			output := setup(t)
			captureExit(t)
			Fatal("Bilbo", "Not all those who wander are lost.")
			if got := output.String(); got != "FATAL | Bilbo | Not all those who wander are lost.\n" {
				t.Errorf("got %q", got)
			}
		})

		t.Run("with key values", func(t *testing.T) {
			output := setup(t)
			captureExit(t)
			Fatal("Bilbo", "Not all those who wander are lost.", "key", "value", "foo", "bar")
			if got := output.String(); got != "FATAL | Bilbo | Not all those who wander are lost. | key='value' foo='bar'\n" {
				t.Errorf("got %q", got)
			}
		})

		t.Run("with corrupt fields", func(t *testing.T) {
			output := setup(t)
			captureExit(t)
			Fatal("", "Not all those who wander are lost.", "key", "value", "foo")
			if got := output.String(); got != "FATAL | Not all those who wander are lost. | corruptFields='key, value, foo'\n" {
				t.Errorf("got %q", got)
			}
		})
	})

	t.Run("Error", func(t *testing.T) {
		t.Run("without ID", func(t *testing.T) {
			output := setup(t)
			Error("", "Not all those who wander are lost.")
			if got := output.String(); got != "ERROR | Not all those who wander are lost.\n" {
				t.Errorf("got %q", got)
			}
		})

		t.Run("with ID", func(t *testing.T) {
			output := setup(t)
			Error("Bilbo", "Not all those who wander are lost.")
			if got := output.String(); got != "ERROR | Bilbo | Not all those who wander are lost.\n" {
				t.Errorf("got %q", got)
			}
		})

		t.Run("with key values", func(t *testing.T) {
			output := setup(t)
			Error("Bilbo", "Not all those who wander are lost.", "key", "value", "foo", "bar")
			if got := output.String(); got != "ERROR | Bilbo | Not all those who wander are lost. | key='value' foo='bar'\n" {
				t.Errorf("got %q", got)
			}
		})

		t.Run("with corrupt fields", func(t *testing.T) {
			output := setup(t)
			Error("", "Not all those who wander are lost.", "key", "value", "foo")
			if got := output.String(); got != "ERROR | Not all those who wander are lost. | corruptFields='key, value, foo'\n" {
				t.Errorf("got %q", got)
			}
		})
	})

	t.Run("Warn", func(t *testing.T) {
		t.Run("logs message", func(t *testing.T) {
			output := setup(t)
			Warn("", "Not all those who wander are lost.")
			if got := output.String(); got != "WARN | Not all those who wander are lost.\n" {
				t.Errorf("got %q", got)
			}
		})

		t.Run("does not log below level", func(t *testing.T) {
			output := setup(t)
			SetLevel(LevelError)
			Warn("", "Not all those who wander are lost.")
			if output.String() != "" {
				t.Errorf("expected empty output, got %q", output.String())
			}
		})
	})

	t.Run("Info", func(t *testing.T) {
		t.Run("logs message", func(t *testing.T) {
			output := setup(t)
			Info("", "Not all those who wander are lost.")
			if got := output.String(); got != "INFO | Not all those who wander are lost.\n" {
				t.Errorf("got %q", got)
			}
		})

		t.Run("does not log below level", func(t *testing.T) {
			output := setup(t)
			SetLevel(LevelWarn)
			Info("", "Not all those who wander are lost.")
			if output.String() != "" {
				t.Errorf("expected empty output, got %q", output.String())
			}
		})
	})

	t.Run("Debug", func(t *testing.T) {
		t.Run("logs message", func(t *testing.T) {
			output := setup(t)
			Debug("", "Not all those who wander are lost.")
			if got := output.String(); got != "DEBUG | Not all those who wander are lost.\n" {
				t.Errorf("got %q", got)
			}
		})

		t.Run("does not log below level", func(t *testing.T) {
			output := setup(t)
			SetLevel(LevelInfo)
			Debug("", "Not all those who wander are lost.")
			if output.String() != "" {
				t.Errorf("expected empty output, got %q", output.String())
			}
		})
	})

	t.Run("Trace", func(t *testing.T) {
		t.Run("logs message", func(t *testing.T) {
			output := setup(t)
			Trace("", "Not all those who wander are lost.")
			if got := output.String(); got != "TRACE | Not all those who wander are lost.\n" {
				t.Errorf("got %q", got)
			}
		})

		t.Run("does not log below level", func(t *testing.T) {
			output := setup(t)
			SetLevel(LevelInfo)
			Trace("", "Not all those who wander are lost.")
			if output.String() != "" {
				t.Errorf("expected empty output, got %q", output.String())
			}
		})
	})

	t.Run("logMessage stack trace", func(t *testing.T) {
		t.Run("adds stack trace", func(t *testing.T) {
			output := setup(t)
			SetStackTrace(true)
			Error("", "Not all those who wander are lost.")
			if defaultStackTrace != true {
				t.Error("expected defaultStackTrace to be true")
			}
			if !strings.Contains(output.String(), "line=") {
				t.Errorf("output %q does not contain stack trace", output.String())
			}
		})

		t.Run("removes stack trace", func(t *testing.T) {
			output := setup(t)
			SetStackTrace(false)
			Error("", "Not all those who wander are lost.")
			if defaultStackTrace != false {
				t.Error("expected defaultStackTrace to be false")
			}
			if strings.Contains(output.String(), "file") {
				t.Errorf("output %q should not contain file", output.String())
			}
			if strings.Contains(output.String(), "line") {
				t.Errorf("output %q should not contain line", output.String())
			}
		})
	})
}

func TestLoggerSetTimestampFlags(t *testing.T) {
	resetLogging(t)

	output := new(bytes.Buffer)
	logger := New(Config{ID: "bilbo"})
	logger.SetLevel(LevelDebug)
	logger.SetTimestampFlags(FlagsDate)
	logger.SetOutput(output)

	message := "Not all those who wander are lost."
	logger.Debug(message)
	out := output.String()
	if !strings.Contains(out, "DEBUG | bilbo | "+message) {
		t.Errorf("output %q does not contain expected content", out)
	}
	if !strings.HasPrefix(out, time.Now().Format("2006/01/02")) {
		t.Errorf("output %q does not start with date prefix", out)
	}

	// Change flags and verify
	output = new(bytes.Buffer)
	logger.SetTimestampFlags(FlagsNone)
	logger.SetOutput(output)
	logger.Debug(message)
	out = output.String()
	if !strings.HasPrefix(out, "DEBUG | bilbo | "+message) {
		t.Errorf("output %q does not start with expected content", out)
	}
}
