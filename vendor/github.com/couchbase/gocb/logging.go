package gocb

import (
	"fmt"
	"gopkg.in/couchbase/gocbcore.v7"
	"log"
	"strings"
)

// LogLevel specifies the severity of a log message.
type LogLevel gocbcore.LogLevel

// Various logging levels (or subsystems) which can categorize the message.
// Currently these are ordered in decreasing severity.
const (
	LogError        = LogLevel(gocbcore.LogError)
	LogWarn         = LogLevel(gocbcore.LogWarn)
	LogInfo         = LogLevel(gocbcore.LogInfo)
	LogDebug        = LogLevel(gocbcore.LogDebug)
	LogTrace        = LogLevel(gocbcore.LogTrace)
	LogSched        = LogLevel(gocbcore.LogSched)
	LogMaxVerbosity = LogLevel(gocbcore.LogMaxVerbosity)
)

// LogRedactLevel specifies the degree with which to redact the logs.
type LogRedactLevel int

const (
	// RedactNone indicates to perform no redactions
	RedactNone = LogRedactLevel(0)

	// RedactPartial indicates to redact all possible user-identifying information from logs.
	RedactPartial = LogRedactLevel(1)

	// RedactFull indicates to fully redact all possible identifying information from logs.
	RedactFull = LogRedactLevel(1)
)

// SetLogRedactionLevel specifies the level with which logs should be redacted.
func SetLogRedactionLevel(level LogRedactLevel) {
	// We don't current log any data that falls under our current redaction rules.
	// This function is included as a stub for future implementations of log redaction
	// that act at a higher level and may need to perform actual redaction's.
}

// Logger defines a logging interface. You can either use one of the default loggers
// (DefaultStdioLogger(), VerboseStdioLogger()) or implement your own.
type Logger interface {
	// Outputs logging information:
	// level is the verbosity level
	// offset is the position within the calling stack from which the message
	// originated. This is useful for contextual loggers which retrieve file/line
	// information.
	Log(level LogLevel, offset int, format string, v ...interface{}) error
}

var (
	globalLogger Logger
)

type coreLogWrapper struct {
	wrapped gocbcore.Logger
}

func (wrapper coreLogWrapper) Log(level LogLevel, offset int, format string, v ...interface{}) error {
	return wrapper.wrapped.Log(gocbcore.LogLevel(level), offset+2, format, v...)
}

// DefaultStdioLogger gets the default standard I/O logger.
//  gocb.SetLogger(gocb.DefaultStdioLogger())
func DefaultStdioLogger() Logger {
	return &coreLogWrapper{
		wrapped: gocbcore.DefaultStdioLogger(),
	}
}

// VerboseStdioLogger is a more verbose level of DefaultStdioLogger(). Messages
// pertaining to the scheduling of ordinary commands (and their responses) will
// also be emitted.
//  gocb.SetLogger(gocb.VerboseStdioLogger())
func VerboseStdioLogger() Logger {
	return coreLogWrapper{
		wrapped: gocbcore.VerboseStdioLogger(),
	}
}

type coreLogger struct {
	wrapped Logger
}

func (wrapper coreLogger) Log(level gocbcore.LogLevel, offset int, format string, v ...interface{}) error {
	return wrapper.wrapped.Log(LogLevel(level), offset+2, format, v...)
}

func getCoreLogger(logger Logger) gocbcore.Logger {
	typedLogger, isCoreLogger := logger.(*coreLogWrapper)
	if isCoreLogger {
		return typedLogger.wrapped
	}

	return &coreLogger{
		wrapped: logger,
	}
}

// SetLogger sets a logger to be used by the library. A logger can be obtained via
// the DefaultStdioLogger() or VerboseStdioLogger() functions. You can also implement
// your own logger using the Logger interface.
func SetLogger(logger Logger) {
	globalLogger = logger
	gocbcore.SetLogger(getCoreLogger(logger))
}

func logExf(level LogLevel, offset int, format string, v ...interface{}) {
	if globalLogger != nil {
		err := globalLogger.Log(level, offset+1, format, v...)
		if err != nil {
			log.Printf("Logger error occurred (%s)\n", err)
		}
	}
}

func logInfof(format string, v ...interface{}) {
	logExf(LogInfo, 1, format, v...)
}

func logDebugf(format string, v ...interface{}) {
	logExf(LogDebug, 1, format, v...)
}

func logSchedf(format string, v ...interface{}) {
	logExf(LogSched, 1, format, v...)
}

func logWarnf(format string, v ...interface{}) {
	logExf(LogWarn, 1, format, v...)
}

func logErrorf(format string, v ...interface{}) {
	logExf(LogError, 1, format, v...)
}

func reindentLog(indent, message string) string {
	reindentedMessage := strings.Replace(message, "\n", "\n"+indent, -1)
	return fmt.Sprintf("%s%s", indent, reindentedMessage)
}
