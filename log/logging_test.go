package log

import (
	"bytes"
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
		osExit = func(code int) {
			didExit = true
			exitCode = code
		}
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
				Expect(output.String()).To(Equal(time.Now().Format("2006/01/02") + " logging.go:458:  | ERROR | id | msg\n"))
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

	Context("Using package level logging", func() {
		BeforeEach(func() {
			output = new(bytes.Buffer)
			Level = LevelTrace
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

				Expect(output.String()).To(Equal("FATAL | Not all those who wander are lost. | key='value' foo=\n"))
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

			It("should print a message with ERROR prefix and key/value pairs and a valueless key", func() {
				Error("", "Not all those who wander are lost.", "key", "value", "foo")

				Expect(output.String()).To(Equal("ERROR | Not all those who wander are lost. | key='value' foo=\n"))
			})
		})

		Describe(".Warn", func() {
			It("should print a formatted message with WARN prefix", func() {
				Warn("", "Not all those who wander are lost.")

				Expect(output.String()).To(Equal("WARN | Not all those who wander are lost.\n"))
			})

			It("should not output anything if log level is lower than LevelWarn", func() {
				Level = LevelError
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
				Level = LevelWarn
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
				Level = LevelInfo
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
				Level = LevelInfo
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
