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

	// TelemetryScenario is the operation for telemetry
	TelemetryScenario = "scenario"

	// DataDir is where we store the downloaded files, logs and state for
	// the extension handler52
	DataDir = "./"

	// MostRecentSequence (mrseq) holds the processed highest sequence number to make sure
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
	AgentName = "DSC"

	// ExtensionHandlerLogFileName is the log file name.
	ExtensionHandlerLogFileName = "gcextn-handler.log"

	// ExtensionDirRegex Regex for finding only Extension directories.
	ExtensionDirRegex = "Microsoft.GuestConfiguration.?(Edp)?.ConfigurationForLinux-([0-9.]*)"

	// GCExtensionVersionRegex returns the version of the extension
	GCExtensionVersionRegex = "^([./a-zA-Z]*)-([0-9.]*)?$"

	// AgentVersionRegex helps return the version of the extension
	AgentVersionRegex = "^([./a-zA-Z0-9]*)_([0-9.]*)?[.](.*)$"

	// If we return failure from update, the Guest Agent goes into an infinite loop. Fixed in the next GA deployment.
	UpdateFailFileName = "./update_failed"
)
