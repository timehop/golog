package log_test

import (
	"bytes"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timehop/golog/log"
)

var _ = Describe("Logging functions", func() {
	var output *bytes.Buffer

	BeforeEach(func() {
		output = new(bytes.Buffer)
		log.Level = log.LevelTrace
		log.SetTimestampFlags(log.FlagsNone)
		log.SetOutput(output)
	})

	Describe(".Error", func() {
		It("should print a message with ERROR prefix without ID", func() {
			log.Error("", "Not all those who wander are lost.")

			Expect(output.String()).To(Equal("ERROR | Not all those who wander are lost.\n"))
		})

		It("should print a message with ERROR prefix and ID", func() {
			log.Error("Bilbo", "Not all those who wander are lost.")

			Expect(output.String()).To(Equal("ERROR | Bilbo | Not all those who wander are lost.\n"))
		})

		It("should print a message with ERROR prefix and key values", func() {
			log.Error("Bilbo", "Not all those who wander are lost.", "key", "value", "foo", "bar")

			Expect(output.String()).To(Equal("ERROR | Bilbo | Not all those who wander are lost. | key='value' foo='bar'\n"))
		})

		It("should print a message with ERROR prefix and key values", func() {
			log.Error("Bilbo", "Not all those who wander are lost.", "key", "value", "foo", "bar")

			Expect(output.String()).To(Equal("ERROR | Bilbo | Not all those who wander are lost. | key='value' foo='bar'\n"))
		})

		It("should print a message with ERROR prefix and key/value pairs and a valueless key", func() {
			log.Error("", "Not all those who wander are lost.", "key", "value", "foo")

			Expect(output.String()).To(Equal("ERROR | Not all those who wander are lost. | key='value' foo=\n"))
		})
	})

	Describe(".Warn", func() {
		It("should print a formatted message with WARN prefix", func() {
			log.Warn("", "Not all those who wander are lost.")

			Expect(output.String()).To(Equal("WARN  | Not all those who wander are lost.\n"))
		})

		It("should not output anything if log level is lower than LevelWarn", func() {
			log.Level = log.LevelError
			log.Warn("", "Not all those who wander are lost.")

			Expect(output.String()).To(BeEmpty())
		})
	})

	Describe(".Info", func() {
		It("should print a formatted message with INFO prefix", func() {
			log.Info("", "Not all those who wander are lost.")

			Expect(output.String()).To(Equal("INFO  | Not all those who wander are lost.\n"))
		})

		It("should not output anything if log level is lower than LevelInfo", func() {
			log.Level = log.LevelWarn
			log.Info("", "Not all those who wander are lost.")

			Expect(output.String()).To(BeEmpty())
		})
	})

	Describe(".Debug", func() {
		It("should print a formatted message with DEBUG prefix", func() {
			log.Debug("", "Not all those who wander are lost.")

			Expect(output.String()).To(Equal("DEBUG | Not all those who wander are lost.\n"))
		})

		It("should not output anything if log level is lower than LevelDebug", func() {
			log.Level = log.LevelInfo
			log.Debug("", "Not all those who wander are lost.")

			Expect(output.String()).To(BeEmpty())
		})
	})

	Describe(".Trace", func() {
		It("should print a formatted message with TRACE prefix", func() {
			log.Trace("", "Not all those who wander are lost.")

			Expect(output.String()).To(Equal("TRACE | Not all those who wander are lost.\n"))
		})

		It("should not output anything if log level is lower than LevelTrace", func() {
			log.Level = log.LevelInfo
			log.Trace("", "Not all those who wander are lost.")

			Expect(output.String()).To(BeEmpty())
		})
	})
})

var _ = Describe("Logger", func() {
	Describe("#SetTimestampFlags", func() {
		It("changes the output of the date", func() {
			output := new(bytes.Buffer)
			logger := log.NewWithID("bilbo")
			logger.Level = log.LevelDebug
			logger.SetTimestampFlags(log.FlagsDate)
			logger.SetOutput(output)

			message := "Not all those who wander are lost."
			logger.Debug(message)
			out := output.String()
			Expect(out).To(ContainSubstring("DEBUG | bilbo | " + message))
			Expect(strings.HasPrefix(out, time.Now().Format("2006/01/02"))).To(BeTrue())

			// And now changing the flags...
			output = new(bytes.Buffer)
			logger.SetTimestampFlags(log.FlagsNone)
			logger.SetOutput(output)
			logger.Debug(message)
			out = output.String()
			Expect(strings.HasPrefix(out, "DEBUG | bilbo | "+message)).To(BeTrue())
		})
	})
})
