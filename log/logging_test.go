package log

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging functions", func() {
	var output *bytes.Buffer
	var didExit bool
	var exitCode int

	// Reset logging
	BeforeEach(func() {
		os.Setenv("LOG_PREFIX", "")
		os.Setenv("LOG_LEVEL", "")
		os.Setenv("LOG_FORMAT", "0")
		os.Setenv("LOG_ENCODING", "")

		initLogging()

		didExit = false
		exitCode = 0
		osExit = func(code int) {
			didExit = true
			exitCode = code
		}
	})

	Describe("JsonFormat", func() {
		Context("Logs a jsonLogEntry struct in json format", func() {
			var timeBefore, timeAfter time.Time
			var entry jsonLogEntry

			BeforeEach(func() {
				output := new(bytes.Buffer)
				SetOutput(output)

				timeBefore = time.Now()
				New(Config{Format: JsonFormat, ID: "id"}).Error("oh no")
				timeAfter = time.Now()

				Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
			})

			It("has a timestamp", func() {
				timestamp, err := time.Parse("2006-01-02 15:04:05 -0700 MST", entry.Timestamp)
				Expect(err).To(BeNil())
				Expect(timestamp).To(BeTemporally(">=", timeBefore))
				Expect(timestamp).To(BeTemporally("<=", timeAfter))
			})

			It("has a log level", func() {
				Expect(entry.Level).To(Equal(LevelErrorName))
			})

			It("has a message", func() {
				Expect(entry.Message).To(Equal("oh no"))
			})

			It("has an id field", func() {
				Expect(entry.Fields).To(HaveKeyWithValue("golog_id", "id"))
			})
		})

		Context("Default prefix is set", func() {
			It("Uses the default prefix as a static field", func() {
				os.Setenv("LOG_PREFIX", "default_prefix")
				initLogging()

				output := new(bytes.Buffer)
				SetOutput(output)
				New(Config{Format: JsonFormat}).Error("oh no")

				var entry jsonLogEntry
				Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
				Expect(entry.Fields).To(HaveKeyWithValue("prefix", "default_prefix"))
			})
		})
	})

	Describe("Fields", func() {
		var logger Logger

		Context("Without static fields", func() {
			var output *bytes.Buffer

			Context("Without dynamic fields", func() {
				BeforeEach(func() {
					logger = New(Config{Format: JsonFormat})
					output = new(bytes.Buffer)
					logger.SetOutput(output)
				})

				It("has empty fields in entry", func() {
					logger.Error("oh no")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Fields).To(BeEmpty())
				})
			})

			Context("With dynamic fields", func() {
				Context("Using a JsonFormat logger", func() {
					It("has fields in entry", func() {
						logger = New(Config{Format: JsonFormat})
						output = new(bytes.Buffer)
						logger.SetOutput(output)

						logger.Error("oh no", "field", "value")

						var entry jsonLogEntry
						Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
						Expect(entry.Fields).To(HaveKeyWithValue("field", "value"))
					})
				})

				Context("Using a PlainTextFormat logger", func() {
					It("has fields in entry", func() {
						logger = New(Config{Format: PlainTextFormat})
						output = new(bytes.Buffer)
						logger.SetOutput(output)

						logger.Error("oh no", "field", "value")

						Expect(output.String()).To(ContainSubstring("field"))
						Expect(output.String()).To(ContainSubstring("value"))
					})
				})
			})
		})

		Context("With static fields", func() {
			var output *bytes.Buffer

			Context("Without dynamic fields", func() {
				BeforeEach(func() {
					logger = New(Config{Format: JsonFormat}, "static_field", "static_value")
					output = new(bytes.Buffer)
					logger.SetOutput(output)
				})

				It("has just static fields in entry", func() {
					logger.Error("oh no")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Fields).To(HaveKeyWithValue("static_field", "static_value"))
					Expect(len(entry.Fields)).To(Equal(1))
				})
			})

			Context("With dynamic fields", func() {
				Context("That are different", func() {
					Context("Using a JsonFormat logger", func() {
						BeforeEach(func() {
							logger = New(Config{Format: JsonFormat}, "static_field", "static_value")
							output = new(bytes.Buffer)
							logger.SetOutput(output)
						})

						It("has both static and dynamic fields in entry", func() {
							logger.Error("oh no", "field", "value")

							var entry jsonLogEntry
							Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
							Expect(entry.Fields).To(HaveKeyWithValue("static_field", "static_value"))
							Expect(entry.Fields).To(HaveKeyWithValue("field", "value"))
							Expect(len(entry.Fields)).To(Equal(2))
						})
					})

					Context("Using a PlainTextFormat logger", func() {
						BeforeEach(func() {
							logger = New(Config{Format: PlainTextFormat}, "static_field", "static_value")
							output = new(bytes.Buffer)
							logger.SetOutput(output)
						})

						It("has both static and dynamic fields in entry", func() {
							logger.Error("oh no", "dynamic_field", "dynamic_value")

							Expect(output.String()).To(ContainSubstring("static_field"))
							Expect(output.String()).To(ContainSubstring("static_value"))
							Expect(output.String()).To(ContainSubstring("dynamic_field"))
							Expect(output.String()).To(ContainSubstring("dynamic_value"))
						})
					})
				})

				Context("That are same as static", func() {
					Context("Using a JsonFormat logger", func() {
						BeforeEach(func() {
							logger = New(Config{Format: JsonFormat}, "static_field", "static_value")
							output = new(bytes.Buffer)
							logger.SetOutput(output)
						})

						It("overrides static field with dynamic value", func() {
							logger.Error("oh no", "static_field", "value")

							var entry jsonLogEntry
							Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
							Expect(entry.Fields).To(HaveKeyWithValue("static_field", "value"))
							Expect(len(entry.Fields)).To(Equal(1))
						})
					})

					Context("Using a PlainTextFormat logger", func() {
						BeforeEach(func() {
							logger = New(Config{Format: PlainTextFormat}, "static_field", "static_value")
							output = new(bytes.Buffer)
							logger.SetOutput(output)
						})

						It("overrides static field with dynamic value", func() {
							logger.Error("oh no", "static_field", "value")

							Expect(output.String()).To(ContainSubstring("static_field"))
							Expect(output.String()).To(ContainSubstring("value"))
							Expect(output.String()).ToNot(ContainSubstring("static_value"))
						})
					})
				})

			})
		})

		Context("With odd number of key-value pairs of fields", func() {
			Context("In static fields", func() {
				It("uses corrupt static field key with regular dynamic fields", func() {
					logger = New(Config{Format: JsonFormat}, "static_field", "static_value", "odd_key")
					output := new(bytes.Buffer)
					logger.SetOutput(output)

					logger.Error("oh no", "key", "value")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Fields).To(HaveKeyWithValue("corruptStaticFields", "static_field, static_value, odd_key"))
					Expect(entry.Fields).To(HaveKeyWithValue("key", "value"))
					Expect(len(entry.Fields)).To(Equal(2))
				})
			})

			Context("In dynamic fields", func() {
				It("uses static fields with corrupt dynamic key", func() {
					logger = New(Config{Format: JsonFormat}, "static_field", "static_value")
					output := new(bytes.Buffer)
					logger.SetOutput(output)

					logger.Error("oh no", "field", "value", "odd_key")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Fields).To(HaveKeyWithValue("static_field", "static_value"))
					Expect(entry.Fields).To(HaveKeyWithValue("corruptFields", "field, value, odd_key"))
					Expect(len(entry.Fields)).To(Equal(2))
				})
			})

			Context("In both static and dynamic fields", func() {
				It("has separate corrupt static and dynamic keys", func() {
					logger = New(Config{Format: JsonFormat}, "static_field", "static_value", "static_odd_key")
					output := new(bytes.Buffer)
					logger.SetOutput(output)

					logger.Error("oh no", "field", "value", "dynamic_odd_key")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Fields).To(HaveKeyWithValue("corruptStaticFields", "static_field, static_value, static_odd_key"))
					Expect(entry.Fields).To(HaveKeyWithValue("corruptFields", "field, value, dynamic_odd_key"))
					Expect(len(entry.Fields)).To(Equal(2))
				})
			})
		})

		Describe("Setting static fields after creating logger", func() {
			var output *bytes.Buffer

			BeforeEach(func() {
				logger = New(Config{Format: JsonFormat}, "old_static_field", "old_static_value")
				output = new(bytes.Buffer)
				logger.SetOutput(output)
				logger.SetStaticField("new_static_field", "new_static_value")
			})

			It("uses the static fields when loggin", func() {
				logger.Error("oh no")

				var entry jsonLogEntry
				Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
				Expect(entry.Fields).To(HaveKeyWithValue("old_static_field", "old_static_value"))
				Expect(entry.Fields).To(HaveKeyWithValue("new_static_field", "new_static_value"))
				Expect(len(entry.Fields)).To(Equal(2))
			})

			It("override the newly set static fields when logging", func() {
				logger.Error("oh no", "new_static_field", "dynamic_value")

				var entry jsonLogEntry
				Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
				Expect(entry.Fields).To(HaveKeyWithValue("old_static_field", "old_static_value"))
				Expect(entry.Fields).To(HaveKeyWithValue("new_static_field", "dynamic_value"))
				Expect(len(entry.Fields)).To(Equal(2))
			})
		})
	})

	Describe("SanitizeFormat()", func() {
		Context("When called with a known format", func() {
			It("returns that format", func() {
				Expect(SanitizeFormat(PlainTextFormat)).To(Equal(PlainTextFormat))
				Expect(SanitizeFormat(JsonFormat)).To(Equal(JsonFormat))
			})
		})

		Context("When called with an unknown format", func() {
			Context("And there's a default from env var", func() {
				It("returns that default", func() {
					os.Setenv("LOG_ENCODING", "json")
					initLogging()

					Expect(SanitizeFormat(LogFormat("whut"))).To(Equal(JsonFormat))

					os.Setenv("LOG_ENCODING", "text")
					initLogging()

					Expect(SanitizeFormat(LogFormat("whut"))).To(Equal(PlainTextFormat))
				})
			})

			Context("And there's no default from env var", func() {
				It("returns global default", func() {
					os.Setenv("LOG_ENCODING", "")
					initLogging()

					Expect(SanitizeFormat(LogFormat("whut"))).To(Equal(PlainTextFormat))
				})
			})
		})
	})

	Describe("Fatal logging fns", func() {
		Context("When logging level artificially set low", func() {
			Describe("Package level Fatal()", func() {
				BeforeEach(func() {
					initLogging()
					SetLevel(LogLevel(-1))

					didExit = false
					exitCode = 0
					output = new(bytes.Buffer)
					SetOutput(output)

					Fatal("id", "msg")
				})

				It("should not log when called", func() {
					Expect(output.String()).To(BeEmpty())
				})

				It("should not exit when called", func() {
					Expect(didExit).To(BeFalse())
					Expect(exitCode).To(Equal(0))
				})
			})

			Describe("Logger Fatal()", func() {
				BeforeEach(func() {
					SetLevel(LogLevel(-1))
					logger := NewDefault()

					output = new(bytes.Buffer)
					logger.SetOutput(output)

					didExit = false
					exitCode = 0

					logger.Fatal("msg")
				})

				It("should not log when called", func() {
					Expect(output.String()).To(BeEmpty())
				})

				It("should not exit when called", func() {
					Expect(didExit).To(BeFalse())
					Expect(exitCode).To(Equal(0))
				})
			})
		})
	})

	Describe("SetPrefix()", func() {
		Context("When modifying existing prefix through package level fn", func() {
			BeforeEach(func() {
				os.Setenv("LOG_PREFIX", "env_prefix ")
				initLogging()
				SetPrefix("fn_prefix ")
			})

			Describe("DefaultLogger", func() {
				It("should use the new prefix", func() {
					output = new(bytes.Buffer)
					SetOutput(output)

					Error("id", "msg")
					Expect(output.String()).To(Equal("fn_prefix ERROR | id | msg\n"))
				})
			})

			Describe("New loggers", func() {
				It("should use the new prefix", func() {
					logger := New(Config{ID: "id"})
					output = new(bytes.Buffer)
					logger.SetOutput(output)

					logger.Error("msg")
					Expect(output.String()).To(Equal("fn_prefix ERROR | id | msg\n"))
				})
			})
		})
	})

	Describe("init()", func() {
		Context("When initialized with prefix set through env var", func() {
			BeforeEach(func() {
				os.Setenv("LOG_PREFIX", "env_prefix ")

				initLogging()
				output = new(bytes.Buffer)
				SetOutput(output)
			})

			It("should use the supplied prefix", func() {
				Error("id", "msg")
				Expect(output.String()).To(Equal("env_prefix ERROR | id | msg\n"))
			})
		})

		Context("When initialized with a format", func() {
			BeforeEach(func() {
				os.Setenv("LOG_FORMAT", strconv.FormatInt(log.Ldate|log.Lshortfile, 10))

				initLogging()
				output = new(bytes.Buffer)
				SetOutput(output)
			})

			It("should log with that format", func() {
				Error("id", "msg")
				Expect(output.String()).To(MatchRegexp("[0-9/]{10} logging.go:[0-9]+: +| ERROR | id | msg\n"))
			})
		})

		Context("When initialized with a log level set through env var", func() {
			Context("Set to FATAL", func() {
				BeforeEach(func() {
					os.Setenv("LOG_LEVEL", "FATAL")
					initLogging()
					output = new(bytes.Buffer)
					SetOutput(output)
				})

				It("should log higher levels", func() {
					Fatal("id", "msg")

					Expect(output.String()).ToNot(BeEmpty())
					Expect(strings.Count(output.String(), "\n")).To(Equal(1))
					Expect(didExit).To(BeTrue())
					Expect(exitCode).To(BeNumerically(">", 0))
				})

				It("should not log lower levels", func() {
					Error("id", "msg")
					Warn("id", "msg")
					Info("id", "msg")
					Debug("id", "msg")
					Trace("id", "msg")

					Expect(output.String()).To(BeEmpty())
				})
			})

			Context("Set to ERROR", func() {
				BeforeEach(func() {
					os.Setenv("LOG_LEVEL", "ERROR")
					initLogging()
					output = new(bytes.Buffer)
					SetOutput(output)
				})

				It("should log higher levels", func() {
					Fatal("id", "msg")
					Error("id", "msg")

					Expect(output.String()).ToNot(BeEmpty())
					Expect(strings.Count(output.String(), "\n")).To(Equal(2))
					Expect(didExit).To(BeTrue())
					Expect(exitCode).To(BeNumerically(">", 0))
				})

				It("should not log lower levels", func() {
					Warn("id", "msg")
					Info("id", "msg")
					Debug("id", "msg")
					Trace("id", "msg")

					Expect(output.String()).To(BeEmpty())
				})
			})

			Context("Set to WARN", func() {
				BeforeEach(func() {
					os.Setenv("LOG_LEVEL", "WARN")
					initLogging()
					output = new(bytes.Buffer)
					SetOutput(output)
				})

				It("should log higher levels", func() {
					Fatal("id", "msg")
					Error("id", "msg")
					Warn("id", "msg")

					Expect(output.String()).ToNot(BeEmpty())
					Expect(strings.Count(output.String(), "\n")).To(Equal(3))
					Expect(didExit).To(BeTrue())
					Expect(exitCode).To(BeNumerically(">", 0))
				})

				It("should not log lower levels", func() {
					Info("id", "msg")
					Debug("id", "msg")
					Trace("id", "msg")

					Expect(output.String()).To(BeEmpty())
				})
			})

			Context("Set to INFO", func() {
				BeforeEach(func() {
					os.Setenv("LOG_LEVEL", "INFO")
					initLogging()
					output = new(bytes.Buffer)
					SetOutput(output)
				})

				It("should log higher levels", func() {
					Fatal("id", "msg")
					Error("id", "msg")
					Warn("id", "msg")
					Info("id", "msg")

					Expect(output.String()).ToNot(BeEmpty())
					Expect(strings.Count(output.String(), "\n")).To(Equal(4))
					Expect(didExit).To(BeTrue())
					Expect(exitCode).To(BeNumerically(">", 0))
				})

				It("should not log lower levels", func() {
					Debug("id", "msg")
					Trace("id", "msg")

					Expect(output.String()).To(BeEmpty())
				})
			})

			Context("Set to DEBUG", func() {
				BeforeEach(func() {
					os.Setenv("LOG_LEVEL", "DEBUG")
					initLogging()
					output = new(bytes.Buffer)
					SetOutput(output)
				})

				It("should log higher levels", func() {
					Fatal("id", "msg")
					Error("id", "msg")
					Warn("id", "msg")
					Info("id", "msg")
					Debug("id", "msg")

					Expect(output.String()).ToNot(BeEmpty())
					Expect(strings.Count(output.String(), "\n")).To(Equal(5))
					Expect(didExit).To(BeTrue())
					Expect(exitCode).To(BeNumerically(">", 0))
				})

				It("should not log lower levels", func() {
					Trace("id", "msg")

					Expect(output.String()).To(BeEmpty())
				})
			})

			Context("Set to TRACE", func() {
				BeforeEach(func() {
					os.Setenv("LOG_LEVEL", "TRACE")
					initLogging()
					output = new(bytes.Buffer)
					SetOutput(output)
				})

				It("should log higher levels", func() {
					Fatal("id", "msg")
					Error("id", "msg")
					Warn("id", "msg")
					Info("id", "msg")
					Debug("id", "msg")
					Trace("id", "msg")

					Expect(output.String()).ToNot(BeEmpty())
					Expect(strings.Count(output.String(), "\n")).To(Equal(6))
					Expect(didExit).To(BeTrue())
					Expect(exitCode).To(BeNumerically(">", 0))
				})
			})
		})
	})

	Context("Using logger logging fns", func() {
		var logger Logger
		var output *bytes.Buffer

		BeforeEach(func() {
			logger = New(Config{Format: JsonFormat})
			output = new(bytes.Buffer)
			logger.SetOutput(output)
		})

		Describe("Fatal", func() {
			Context("Below level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelFatal)
				})

				It("should log", func() {
					logger.Fatal("oh no")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Level).To(Equal(LevelFatalName))
				})
			})

			Context("Above level", func() {
				BeforeEach(func() {
					logger.SetLevel(LogLevel(-1))
				})

				It("should not log", func() {
					logger.Fatal("oh no")
					Expect(output.Len()).To(Equal(0))
				})
			})
		})

		Describe("Error", func() {
			Context("Below level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelError)
				})

				It("should log", func() {
					logger.Error("oh no")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Level).To(Equal(LevelErrorName))
				})
			})

			Context("Above level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelFatal)
				})

				It("should not log", func() {
					logger.Error("oh no")
					Expect(output.Len()).To(Equal(0))
				})
			})
		})

		Describe("Warn", func() {
			Context("Below level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelWarn)
				})

				It("should log", func() {
					logger.Warn("oh no")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Level).To(Equal(LevelWarnName))
				})
			})

			Context("Above level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelFatal)
				})

				It("should not log", func() {
					logger.Warn("oh no")
					Expect(output.Len()).To(Equal(0))
				})
			})
		})

		Describe("Info", func() {
			Context("Below level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelInfo)
				})

				It("should log", func() {
					logger.Info("oh no")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Level).To(Equal(LevelInfoName))
				})
			})

			Context("Above level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelFatal)
				})

				It("should not log", func() {
					logger.Info("oh no")
					Expect(output.Len()).To(Equal(0))
				})
			})
		})

		Describe("Debug", func() {
			Context("Below level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelDebug)
				})

				It("should log", func() {
					logger.Debug("oh no")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Level).To(Equal(LevelDebugName))
				})
			})

			Context("Above level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelFatal)
				})

				It("should not log", func() {
					logger.Debug("oh no")
					Expect(output.Len()).To(Equal(0))
				})
			})
		})

		Describe("Trace", func() {
			Context("Below level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelTrace)
				})

				It("should log", func() {
					logger.Trace("oh no")

					var entry jsonLogEntry
					Expect(json.Unmarshal(output.Bytes(), &entry)).To(BeNil())
					Expect(entry.Level).To(Equal(LevelTraceName))
				})
			})

			Context("Above level", func() {
				BeforeEach(func() {
					logger.SetLevel(LevelFatal)
				})

				It("should not log", func() {
					logger.Trace("oh no")
					Expect(output.Len()).To(Equal(0))
				})
			})
		})
	})

	Context("Using package level logging", func() {
		BeforeEach(func() {
			output = new(bytes.Buffer)
			SetLevel(LevelTrace)
			SetTimestampFlags(FlagsNone)
			SetOutput(output)
		})

		Describe(".Fatal", func() {
			It("should print a message with FATAL prefix without ID", func() {
				Fatal("", "Not all those who wander are lost.")

				Expect(output.String()).To(Equal("FATAL | Not all those who wander are lost.\n"))
			})

			It("should print a message with FATAL prefix and ID", func() {
				Fatal("Bilbo", "Not all those who wander are lost.")

				Expect(output.String()).To(Equal("FATAL | Bilbo | Not all those who wander are lost.\n"))
			})

			It("should print a message with FATAL prefix and key values", func() {
				Fatal("Bilbo", "Not all those who wander are lost.", "key", "value", "foo", "bar")

				Expect(output.String()).To(Equal("FATAL | Bilbo | Not all those who wander are lost. | key='value' foo='bar'\n"))
			})

			It("should print a message with FATAL prefix and key values", func() {
				Fatal("Bilbo", "Not all those who wander are lost.", "key", "value", "foo", "bar")

				Expect(output.String()).To(Equal("FATAL | Bilbo | Not all those who wander are lost. | key='value' foo='bar'\n"))
			})

			It("should print a message with FATAL prefix and key/value pairs and a valueless key", func() {
				Fatal("", "Not all those who wander are lost.", "key", "value", "foo")

				Expect(output.String()).To(Equal("FATAL | Not all those who wander are lost. | corruptFields='key, value, foo'\n"))
			})
		})

		Describe(".Error", func() {
			It("should print a message with ERROR prefix without ID", func() {
				Error("", "Not all those who wander are lost.")

				Expect(output.String()).To(Equal("ERROR | Not all those who wander are lost.\n"))
			})

			It("should print a message with ERROR prefix and ID", func() {
				Error("Bilbo", "Not all those who wander are lost.")

				Expect(output.String()).To(Equal("ERROR | Bilbo | Not all those who wander are lost.\n"))
			})

			It("should print a message with ERROR prefix and key values", func() {
				Error("Bilbo", "Not all those who wander are lost.", "key", "value", "foo", "bar")

				Expect(output.String()).To(Equal("ERROR | Bilbo | Not all those who wander are lost. | key='value' foo='bar'\n"))
			})

			It("should print a message with ERROR prefix and key values", func() {
				Error("Bilbo", "Not all those who wander are lost.", "key", "value", "foo", "bar")

				Expect(output.String()).To(Equal("ERROR | Bilbo | Not all those who wander are lost. | key='value' foo='bar'\n"))
			})

			It("should print a message with ERROR prefix and corrupt fields", func() {
				Error("", "Not all those who wander are lost.", "key", "value", "foo")

				Expect(output.String()).To(Equal("ERROR | Not all those who wander are lost. | corruptFields='key, value, foo'\n"))
			})
		})

		Describe(".Warn", func() {
			It("should print a formatted message with WARN prefix", func() {
				Warn("", "Not all those who wander are lost.")

				Expect(output.String()).To(Equal("WARN | Not all those who wander are lost.\n"))
			})

			It("should not output anything if log level is lower than LevelWarn", func() {
				SetLevel(LevelError)
				Warn("", "Not all those who wander are lost.")

				Expect(output.String()).To(BeEmpty())
			})
		})

		Describe(".Info", func() {
			It("should print a formatted message with INFO prefix", func() {
				Info("", "Not all those who wander are lost.")

				Expect(output.String()).To(Equal("INFO | Not all those who wander are lost.\n"))
			})

			It("should not output anything if log level is lower than LevelInfo", func() {
				SetLevel(LevelWarn)
				Info("", "Not all those who wander are lost.")

				Expect(output.String()).To(BeEmpty())
			})
		})

		Describe(".Debug", func() {
			It("should print a formatted message with DEBUG prefix", func() {
				Debug("", "Not all those who wander are lost.")

				Expect(output.String()).To(Equal("DEBUG | Not all those who wander are lost.\n"))
			})

			It("should not output anything if log level is lower than LevelDebug", func() {
				SetLevel(LevelInfo)
				Debug("", "Not all those who wander are lost.")

				Expect(output.String()).To(BeEmpty())
			})
		})

		Describe(".Trace", func() {
			It("should print a formatted message with TRACE prefix", func() {
				Trace("", "Not all those who wander are lost.")

				Expect(output.String()).To(Equal("TRACE | Not all those who wander are lost.\n"))
			})

			It("should not output anything if log level is lower than LevelTrace", func() {
				SetLevel(LevelInfo)
				Trace("", "Not all those who wander are lost.")

				Expect(output.String()).To(BeEmpty())
			})
		})
	})
})

var _ = Describe("Logger", func() {
	Describe("#SetTimestampFlags", func() {
		It("changes the output of the date", func() {
			output := new(bytes.Buffer)
			logger := New(Config{ID: "bilbo"})
			logger.SetLevel(LevelDebug)
			logger.SetTimestampFlags(FlagsDate)
			logger.SetOutput(output)

			message := "Not all those who wander are lost."
			logger.Debug(message)
			out := output.String()
			Expect(out).To(ContainSubstring("DEBUG | bilbo | " + message))
			Expect(strings.HasPrefix(out, time.Now().Format("2006/01/02"))).To(BeTrue())

			// And now changing the flags...
			output = new(bytes.Buffer)
			logger.SetTimestampFlags(FlagsNone)
			logger.SetOutput(output)
			logger.Debug(message)
			out = output.String()
			Expect(strings.HasPrefix(out, "DEBUG | bilbo | "+message)).To(BeTrue())
		})
	})
})
