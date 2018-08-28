package main

import (
	"os"

	"github.com/go-kit/kit/log"
	"io"
	golog "log"
	"path"
)

// ExtensionLogger for all the extension related events
type ExtensionLogger struct {
	logger      log.Logger
	logFilePath string
}

// create a new ExtensionLogger
func newLogger(logDir string) ExtensionLogger {
	if err := os.MkdirAll(logPath, 0644); err != nil {
		golog.Printf("ERROR: Cannot create log folder %s: %v \r\n", logDir, err)
	}

	extensionLogPath := path.Join(logPath, ExtensionHandlerLogFileName)
	golog.Printf("Logging in file %s: in directory %s: .\r\n", ExtensionHandlerLogFileName, logPath)

	fileHandle, err := os.OpenFile(extensionLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		golog.Fatalf("ERROR: Cannot open log file: %v \r\n", err)
	}

	fileLogger := log.With(
		log.With(
			log.NewSyncLogger(
				log.NewLogfmtLogger(
					io.MultiWriter(
						os.Stdout,
						os.Stderr,
						fileHandle))),
			"time", log.DefaultTimestamp),
		"version", VersionString())
	lg := ExtensionLogger{fileLogger, extensionLogPath}

	lg.event("ExtensionLogPath: " + extensionLogPath)

	return lg
}

func (lg ExtensionLogger) with(key string, value string) {
	lg.logger = log.With(lg.logger, key, value)
}

func (lg ExtensionLogger) output(output string) {
	lg.logger.Log(logOutput, output)
}

func (lg ExtensionLogger) event(event string) {
	lg.logger.Log(logEvent, event)
}

func (lg ExtensionLogger) eventError(event string, error error) {
	lg.logger.Log(logEvent, event, logError, error)
}

func (lg ExtensionLogger) customLog(keyvals ...interface{}) {
	lg.logger.Log(keyvals)
}
