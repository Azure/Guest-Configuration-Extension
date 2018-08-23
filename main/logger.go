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
	logger log.Logger
}

// create a new LinuxLogger
func newLogger() LinuxLogger {
	nop := log.With(log.With(log.NewSyncLogger(log.NewLogfmtLogger(
		os.Stdout)), "time", log.DefaultTimestamp), "version", VersionString())
	lg := LinuxLogger{nop}

	return lg
}

func (lg LinuxLogger) with(key string, value string) {
	lg.logger = log.With(lg.logger, key, value)
}

func (lg LinuxLogger) output(output string) {
	lg.logger.Log(logOutput, output)
}

func (lg LinuxLogger) event(event string, message string) {
	if message != "" {
		lg.logger.Log(logEvent, event, logMessage, message)
	} else {
		lg.logger.Log(logEvent, event)
	}
}

func (lg LinuxLogger) message(message string) {
	lg.logger.Log(logMessage, message)
}

func (lg LinuxLogger) messageAndError(message string, error error) {
	lg.logger.Log(logMessage, message, logError, error)
}

func (lg LinuxLogger) customLog(keyvals ...interface{}) {
	lg.logger.Log(keyvals)
}
