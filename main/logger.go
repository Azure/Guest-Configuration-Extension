package main

import (
    "os"
    "path"
    //"log"

    golog "log"
)

// ExtensionLogger for all the extension-related events
type ExtensionLogger struct {
    //logger      *logrus.Logger
    logFilePath string
}

type NoopWriter struct{}

func (n *NoopWriter) Write(p []byte) (int, error) {
    return len(p), nil
}

// create a new ExtensionLogger
func newLogger(logDir string) ExtensionLogger {
        if err := os.MkdirAll(logDir, 0755); err != nil {
        golog.Printf("ERROR: Cannot create log folder %s: %v \r\n", logDir, err)
    }

    extensionLogPath := path.Join(logDir, ExtensionHandlerLogFileName)
    golog.Printf("Logging in file %s: in directory %s: .\r\n", ExtensionHandlerLogFileName, logDir)

    fileHandle, err := os.OpenFile(extensionLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        golog.Fatalf("ERROR: Cannot open log file: %v \r\n", err)
    }

    // Create a new Logrus logger
    // logger := logrus.New()
    // logger.SetOutput(fileHandle) // Log to the file
    // logger.SetLevel(logrus.InfoLevel)

    // Redirect standard output and error to the log file
    log.SetOutput(fileHandle)

    // Return the ExtensionLogger
    return ExtensionLogger{logFilePath: extensionLogPath}
}

func newNoopLogger() ExtensionLogger {
    return ExtensionLogger{logFilePath: ""}
}

// Add a key-value pair to the logger
func (lg ExtensionLogger) with(key string, value string) {
    //lg.logger.WithField(key, value).Info("")
    golog.Printf("Added context: %s=%s\n", key, value)
}

// Log an event
func (lg ExtensionLogger) event(event string) {
    //lg.logger.Info(event)
    golog.Println(output)
}

// Log an error event
func (lg ExtensionLogger) eventError(event string, err error) {
    //lg.logger.WithError(err).Error(event)
    golog.Println("ERROR: %s: %v \r\n", event, err)
}

// Log custom key-value pairs
func (lg ExtensionLogger) customLog(keyvals ...interface{}) {
    fields := logrus.Fields{}
    for i := 0; i < len(keyvals)-1; i += 2 {
        key, ok := keyvals[i].(string)
        if !ok {
            continue
        }
        //fields[key] = keyvals[i+1]
        golog.Printf("%s=%v ", key, keyvals[i+1])
    }
    //lg.logger.WithFields(fields).Info("Custom log")
    golog.Println()
}
