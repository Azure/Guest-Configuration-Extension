package main

import (
	"os"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"

	"github.com/pkg/errors"
)

type cmdfunc func(lg ExtensionLogger, hEnv vmextension.HandlerEnvironment, seqNum int) error
type prefunc func(lg ExtensionLogger, seqNum int) error

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

func install(lg ExtensionLogger, hEnv vmextension.HandlerEnvironment, seqNum int) error {
	msg := "Extension install succeeded"
	lg.event(msg)
	telemetry(TelemetryScenario, msg, true, 0)
	return nil
}

func enablePre(lg ExtensionLogger, seqNum int) error {
	// exit if this sequence number is already processed
	// if not, save the sequence number before proceeding
	if shouldExit, err := checkAndSaveSeqNum(lg, seqNum, MostRecentSequence); err != nil {
		return errors.Wrap(err, "failed to process seqnum")
	} else if shouldExit {
		lg.eventError("exit", errors.New("this sequence number smaller than the currently processed sequence number, will not run again"))
		os.Exit(successCode)
	}
	return nil
}

func enable(lg ExtensionLogger, hEnv vmextension.HandlerEnvironment, seqNum int) error {
	// parse the extension handler settings (file not available prior to 'enable')
	cfg, err := parseAndValidateSettings(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return errors.Wrap(err, "failed to get configuration")
	}

	// parse and log the agent version
	//_, err = parseAndLogAgentVersion(AgentZipDir)
	//if err != nil {
	//	lg.customLog(logEvent, "failed to parse version string", logError, err, logAgentName, AgentZipDir)
	//	return errors.Wrap(err, "failed to parse version string")
	//}

	// check to see if agent directory exists
	unzipDir, agentDirectory := getAgentPaths()
	var runErr error = nil
	if _, err := os.Stat(agentDirectory); err == nil {
		// directory exists, run enable.sh for agent health check
		lg.event("agent health check")

		runErr = runCmd(lg, "bash ./enable.sh", agentDirectory, cfg)
		if runErr != nil {
			lg.eventError("agent health check failed", runErr)
			os.Exit(agentHealthCheckFailedCode)
		}
		os.Exit(successCode)
	}

	// directory does not exist, unzipAgent agent
	_, err = unzipAgent(lg, AgentZipDir, AgentName, unzipDir)
	if err != nil {
		lg.eventError("failed to unzipAgent agent dir", err)
		return errors.Wrap(err, "failed to unzipAgent agent")
	}
	// run install.sh and enable.sh
	lg.event("installing agent")
	runErr = runCmd(lg, "bash ./install.sh", agentDirectory, cfg)
	if runErr != nil {
		lg.eventError("agent installation failed", runErr)
		telemetry(TelemetryScenario, "agent installation failed: "+runErr.Error(), false, 0)
	} else {
		lg.customLog(logEvent, "agent installation succeeded", logEvent, "enabling agent")
		telemetry(TelemetryScenario, "agent installation succeeded", true, 0)
		runErr = runCmd(lg, "bash ./enable.sh", agentDirectory, cfg)
		if runErr != nil {
			lg.eventError("enable agent failed", runErr)
			telemetry(TelemetryScenario, "agent enable failed: "+runErr.Error(), false, 0)
		} else {
			lg.event("enable agent succeeded")
			telemetry(TelemetryScenario, "agent enable succeeded", true, 0)
		}
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(lg, unzipDir, runErr)

	return runErr
}

func update(lg ExtensionLogger, hEnv vmextension.HandlerEnvironment, seqNum int) error {
	// parse the extension handler settings
	cfg, err := parseAndValidateSettings(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return errors.Wrap(err, "failed to get configuration")
	}

	// get old agent path
	oldAgent, err := getOldAgentPath(lg)
	if err != nil {
		lg.eventError("failed to get old agent path", err)
		return errors.Wrap(err, "failed to get old agent path")
	}

	// parse and log the new agent version
	//_, err = parseAndLogAgentVersion(lg, AgentZipDir)
	//if err != nil {
	//	lg.customLog(logEvent, "failed to parse version string", logError, err, logAgentName, AgentZipDir)
	//	return errors.Wrap(err, "failed to parse version string")
	//}

	// unzipAgent new agent
	unzipDir, agentDirectory := getAgentPaths()
	_, err = unzipAgent(lg, AgentZipDir, AgentName, unzipDir)
	if err != nil {
		lg.eventError("failed to unzipAgent agent dir", err)
		return errors.Wrap(err, "failed to unzipAgent agent")
	}

	// run new update.sh to update the agent
	lg.event("updating agent")
	runErr := runCmd(lg, "bash ./update.sh "+oldAgent, agentDirectory, cfg)
	if runErr != nil {
		lg.eventError("agent update failed", runErr)
		telemetry(TelemetryScenario, "agent update failed: "+runErr.Error(), false, 0)
	} else {
		lg.event("agent update succeeded")
		telemetry(TelemetryScenario, "agent update succeeded", true, 0)
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(lg, unzipDir, runErr)

	return runErr
}

func disable(lg ExtensionLogger, hEnv vmextension.HandlerEnvironment, seqNum int) error {
	// parse the extension handler settings
	cfg, err := parseAndValidateSettings(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return errors.Wrap(err, "failed to get configuration")
	}

	// run disable.sh to disable the agent
	lg.event("disabling agent")
	unzipDir, agentDirectory := getAgentPaths()
	runErr := runCmd(lg, "bash ./disable.sh", agentDirectory, cfg)
	if runErr != nil {
		lg.eventError("agent disable failed", runErr)
		telemetry(TelemetryScenario, "agent disable failed: "+runErr.Error(), false, 0)
	} else {
		lg.event("agent disable succeeded")
		telemetry(TelemetryScenario, "agent disable succeeded", true, 0)
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(lg, unzipDir, runErr)

	return nil
}

func uninstall(lg ExtensionLogger, hEnv vmextension.HandlerEnvironment, seqNum int) error {
	// parse the extension handler settings
	cfg, err := parseAndValidateSettings(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return errors.Wrap(err, "failed to get configuration")
	}

	// run uninstall.sh to uninstall the agent
	lg.event("uninstalling agent")
	unzipDir, agentDirectory := getAgentPaths()
	runErr := runCmd(lg, "bash ./uninstall.sh", agentDirectory, cfg)
	if runErr != nil {
		lg.eventError("agent uninstall failed", runErr)
		telemetry(TelemetryScenario, "agent uninstall failed: "+runErr.Error(), false, 0)
	} else {
		lg.event("agent uninstall succeeded")
		telemetry(TelemetryScenario, "agent uninstall succeeded", true, 0)
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(lg, unzipDir, runErr)

	return nil
}
