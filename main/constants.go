package main

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

	// Error codes for each command
	installCode = 100

	enableCode                 = 200
	agentHealthCheckFailedCode = 201

	updateCode = 300

	disableCode = 400

	uninstallCode = 500

	// Generic error codes
	successCode    = 0
	invalidCmdCode = 2
)
