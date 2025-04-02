package main

import (
    "os"
    "path"

    "github.com/sirupsen/logrus"
    golog "log"
)

// ExtensionLogger for all the extension-related events
type ExtensionLogger struct {
    logger      *logrus.Logger
    logFilePath string
}

type NoopWriter struct{}

func (n *NoopWriter) Write(p []byte) (int, error) {
    return len(p), nil
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

    // Create a new Logrus logger
    logger := logrus.New()
    logger.SetOutput(fileHandle) // Log to the file
    logger.SetLevel(logrus.InfoLevel)

    // Return the ExtensionLogger
    return ExtensionLogger{logger: logger, logFilePath: extensionLogPath}
}

// newNoopLogger creates a Logrus logger that discards all log output
func newNoopLogger() ExtensionLogger {
    logger := logrus.New()
    logger.SetOutput(&NoopWriter{}) // Discard all log output
    return ExtensionLogger{logger: logger, logFilePath: ""}
}

// Add a key-value pair to the logger
func (lg ExtensionLogger) with(key string, value string) {
    lg.logger.WithField(key, value).Info("")
}

// Log an event
func (lg ExtensionLogger) event(event string) {
    lg.logger.Info(event)
}

// Log an error event
func (lg ExtensionLogger) eventError(event string, err error) {
    lg.logger.WithError(err).Error(event)
}

// Log custom key-value pairs
func (lg ExtensionLogger) customLog(keyvals ...interface{}) {
    fields := logrus.Fields{}
    for i := 0; i < len(keyvals)-1; i += 2 {
        key, ok := keyvals[i].(string)
        if !ok {
            continue
        }
        fields[key] = keyvals[i+1]
    }
    lg.logger.WithFields(fields).Info("Custom log")
}
