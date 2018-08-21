package main

import (
	"os"

	"github.com/go-kit/kit/log"
)

const (
	// Logging and telemetry operations
	telemetryScenario = "scenario"

	logOutput    = "output"
	logMessage   = "message"
	logEvent     = "event"
	logError     = "error"
	logAgentName = "agentName"
	logVersion   = "version"
	logPath      = "path"
)

type LinuxLogger struct {
	lg log.Logger
}

// create a new LinuxLogger
func newLogger() LinuxLogger {
	nop := log.With(log.With(log.NewSyncLogger(log.NewLogfmtLogger(
		os.Stdout)), "time", log.DefaultTimestamp), "version", VersionString())
	logger := LinuxLogger{nop}

	return logger
}

func (logger LinuxLogger) getLogger() log.Logger {
	return logger.lg
}

func (logger LinuxLogger) with(key string, value string) {
	logger.lg = log.With(logger.lg, key, value)
}

func (logger LinuxLogger) output(output string) {
	logger.lg.Log(logOutput, output)
}

func (logger LinuxLogger) event(event string) {
	logger.lg.Log(logEvent, event)
}

func (logger LinuxLogger) message(message string) {
	logger.lg.Log(logMessage, message)
}

func (logger LinuxLogger) messageAndError(message string, error error) {
	logger.lg.Log(logMessage, message, logError, error)
}

func (logger LinuxLogger) eventAndMessage(event string, message string) {
	logger.lg.Log(logEvent, event, logMessage, message)
}

func (logger LinuxLogger) customLogSingle(key string, value string) {
	logger.lg.Log(key, value)
}

func (logger LinuxLogger) customLogMultiple(keyvals ...interface{}) {
	logger.lg.Log(keyvals)
}
