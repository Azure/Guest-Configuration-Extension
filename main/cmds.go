package main

import (
	"os"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"

	"github.com/pkg/errors"
)

type cmdfunc func(hEnv vmextension.HandlerEnvironment, seqNum int) (string, error)
type prefunc func(seqNum int) error

// Add more fields as necessary
type cmd struct {
	f                  cmdfunc // associated function
	name               string  // human readable string
	shouldReportStatus bool    // determines if running this should log to a .status file
	pre                prefunc // executed before any status is reported
	failExitCode       int     // exitCode to use when commands fail
}

const (
	fullName                = "Microsoft.Azure.Extensions.GuestConfigurationForLinux"
	maxTailLen              = 4 * 1024 // length of max stdout/stderr to be transmitted in .status file
	maxTelemetryTailLen int = 1800
)

var (
	telemetry = sendTelemetry(newTelemetryEventSender(), fullName, Version)

	// allowed user inputs
	cmds = map[string]cmd{
		"install":   {install, "install", false, nil, installCode},
		"enable":    {enable, "enable", true, enablePre, enableCode},
		"update":    {update, "update", true, nil, updateCode},
		"disable":   {disable, "disable", true, nil, disableCode},
		"uninstall": {uninstall, "uninstall", false, nil, uninstallCode},
	}
)

func install(hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	lg.event("installed", "")
	telemetry(telemetryScenario, "extension install succeeded", true, 0)

	return "", nil
}

func enablePre(seqNum int) error {
	// exit if this sequence number is already processed
	// if not, save the sequence number before proceeding
	if shouldExit, err := checkAndSaveSeqNum(seqNum, mostRecentSequence); err != nil {
		return errors.Wrap(err, "failed to process seqnum")
	} else if shouldExit {
		lg.event("exit", "this sequence number smaller than the currently processed sequence number, will not run again")
		os.Exit(successCode)
	}
	return nil
}

func enable(hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// parse the extension handler settings (file not available prior to 'enable')
	cfg, err := parseAndValidateSettings(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	// parse and log the agent version
	_, err = parseAndLogAgentVersion(agentZip)
	if err != nil {
		lg.customLog(logMessage, "failed to parse version string", logError, err, logAgentName, agentZip)
		return "", errors.Wrap(err, "failed to parse version string")
	}

	// check to see if agent directory exists
	unzipDir, agentDirectory := unzipAndAgentDirectories()
	var runErr error = nil
	if _, err := os.Stat(agentDirectory); err == nil {
		// directory exists, run enable.sh for agent health check
		lg.event("agent health check", "")

		runErr = runCmd("bash ./enable.sh", agentDirectory, cfg)
		if runErr != nil {
			lg.messageAndError("agent health check failed", runErr)
			os.Exit(agentHealthCheckFailedCode)
		}
		os.Exit(successCode)
	}

	// directory does not exist, unzip agent
	_, err = unzip(agentZip, unzipDir)
	if err != nil {
		lg.messageAndError("failed to unzip agent dir", err)
		return "", errors.Wrap(err, "failed to unzip agent")
	}
	// run install.sh and enable.sh
	lg.event("installing agent", "")
	runErr = runCmd("bash ./install.sh", agentDirectory, cfg)
	if runErr != nil {
		lg.messageAndError("agent installation failed", runErr)
		telemetry(telemetryScenario, "agent installation failed: "+runErr.Error(), false, 0)
	} else {
		lg.customLog(logMessage, "agent installation succeeded", logEvent, "enabling agent")
		telemetry(telemetryScenario, "agent installation succeeded", true, 0)
		runErr = runCmd("bash ./enable.sh", agentDirectory, cfg)
		if runErr != nil {
			lg.messageAndError("enable agent failed", runErr)
			telemetry(telemetryScenario, "agent enable failed: "+runErr.Error(), false, 0)
		} else {
			lg.message("enable agent succeeded")
			telemetry(telemetryScenario, "agent enable succeeded", true, 0)
		}
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(unzipDir, runErr)

	msg := ""

	return msg, runErr
}

func update(hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// parse the extension handler settings
	cfg, err := parseAndValidateSettings(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	// get old agent path
	oldAgent, err := getOldAgentPath()
	if err != nil {
		lg.messageAndError("failed to get old agent path", err)
		return "", errors.Wrap(err, "failed to get old agent path")
	}

	// parse and log the new agent version
	_, err = parseAndLogAgentVersion(agentZip)
	if err != nil {
		lg.customLog(logMessage, "failed to parse version string", logError, err, logAgentName, agentZip)
		return "", errors.Wrap(err, "failed to parse version string")
	}

	// unzip new agent
	unzipDir, agentDirectory := unzipAndAgentDirectories()
	_, err = unzip(agentZip, unzipDir)
	if err != nil {
		lg.messageAndError("failed to unzip agent dir", err)
		return "", errors.Wrap(err, "failed to unzip agent")
	}

	// run new update.sh to update the agent
	lg.event("updating agent", "")
	runErr := runCmd("bash ./update.sh "+oldAgent, agentDirectory, cfg)
	if runErr != nil {
		lg.messageAndError("agent update failed", runErr)
		telemetry(telemetryScenario, "agent update failed: "+runErr.Error(), false, 0)
	} else {
		lg.message("agent update succeeded")
		telemetry(telemetryScenario, "agent update succeeded", true, 0)
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(unzipDir, runErr)

	msg := ""

	return msg, runErr
}

func disable(hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// parse the extension handler settings
	cfg, err := parseAndValidateSettings(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	// run disable.sh to disable the agent
	lg.event("disabling agent", "")
	unzipDir, agentDirectory := unzipAndAgentDirectories()
	runErr := runCmd("bash ./disable.sh", agentDirectory, cfg)
	if runErr != nil {
		lg.messageAndError("agent disable failed", runErr)
		telemetry(telemetryScenario, "agent disable failed: "+runErr.Error(), false, 0)
	} else {
		lg.message("agent disable succeeded")
		telemetry(telemetryScenario, "agent disable succeeded", true, 0)
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(unzipDir, runErr)

	msg := ""

	return msg, nil
}

func uninstall(hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// parse the extension handler settings
	cfg, err := parseAndValidateSettings(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	// run uninstall.sh to uninstall the agent
	lg.event("uninstalling agent", "")
	unzipDir, agentDirectory := unzipAndAgentDirectories()
	runErr := runCmd("bash ./uninstall.sh", agentDirectory, cfg)
	if runErr != nil {
		lg.messageAndError("agent uninstall failed", runErr)
		telemetry(telemetryScenario, "agent uninstall failed: "+runErr.Error(), false, 0)
	} else {
		lg.message("agent uninstall succeeded")
		telemetry(telemetryScenario, "agent uninstall succeeded", true, 0)
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(unzipDir, runErr)

	msg := ""

	return msg, nil
}
