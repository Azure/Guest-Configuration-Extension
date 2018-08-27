package main

const (
	// Error codes for each command
	installCode                = 100
	enableCode                 = 200
	agentHealthCheckFailedCode = 201
	updateCode                 = 300
	disableCode                = 400
	uninstallCode              = 500

	// Generic error codes
	successCode    = 0
	failureCode    = -1
	invalidCmdCode = 2

	// Logging operations
	logOutput    = "output"
	logEvent     = "event"
	logError     = "error"
	logAgentName = "AgentName"
	logVersion   = "version"
	logPath      = "path"

	// telemetry operations
	TelemetryScenario = "scenario"

	// DataDir is where we store the downloaded files, logs and state for
	// the extension handler
	DataDir = "./"

	// mrseq holds the processed highest sequence number to make sure
	// we do not run the command more than once for the same sequence
	// number. Stored under DataDir. This file is auto-preserved by the agent.
	MostRecentSequence = "mrseq"

	// UnzipAgentDir is where the agent is stored
	// stored until DataDir
	UnzipAgentDir = "GCAgent"

	// AgentZipDir is the directory where the agent package is stored
	// it will be unzipped into {DataDir}/GCAgent/{version}/agent
	AgentZipDir = "agent"

	// AgentName contains the .sh files
	// stored under the agent version
	AgentName = "DesiredStateConfiguration"

	// ExtensionHandlerLogFileName is the log file name.
	ExtensionHandlerLogFileName = "gcextn-handler.log"

	ExtensionDirPrefix = "Microsoft.GuestConfiguration.Edp.ConfigurationForLinux"
)
