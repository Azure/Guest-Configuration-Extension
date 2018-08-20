package main

import (
	"os"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"

	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"time"

	"github.com/Azure/Guest-Configuration-Extension/pkg/seqnum"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type cmdfunc func(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error)
type prefunc func(logger log.Logger, seqNum int) error

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

func install(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	logger.Log(logEvent, "installed")

	return "", nil
}

func enablePre(logger log.Logger, seqNum int) error {
	// exit if this sequence number is already processed
	// if not, save the sequence number before proceeding
	if shouldExit, err := checkAndSaveSeqNum(logger, seqNum, mostRecentSequence); err != nil {
		return errors.Wrap(err, "failed to process seqnum")
	} else if shouldExit {
		logger.Log(logEvent, "exit", logMessage,
			"this sequence number smaller than the currently processed sequence number, will not run again")
		os.Exit(successCode)
	}
	return nil
}

func enable(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// parse the extension handler settings (file not available prior to 'enable')
	cfg, err := parseAndValidateSettings(logger, hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	// parse the version string, log it, and send it through telemetry
	version, err := parseVersionString(agentZip)
	if err != nil {
		logger.Log(logMessage, "failed to parse version string", logError, err, logAgentName, agentZip)
		return "", errors.Wrap(err, "failed to parse version string")
	}
	logger.Log(logMessage, "current agent version", logVersion, version)
	// TODO Scenarios and Message should be variables. Better Enums.
	telemetry(telemetryScenario, "Current agent version: "+version, true, 0)

	// check to see if agent directory exists
	unzipDir, agentDirectory := unzipAndAgentDirectories()
	var runErr error = nil
	if _, err := os.Stat(agentDirectory); err == nil {
		// directory exists, run enable.sh for agent health check
		logger.Log(logEvent, "running agent health check")

		runErr = runCmd(logger, "bash ./enable.sh", agentDirectory, cfg)
		if runErr != nil {
			logger.Log(logMessage, "agent health check failed", logError, runErr)
			os.Exit(agentHealthCheckFailedCode)
		}
		os.Exit(successCode)
	}

	// directory does not exist, unzip agent
	_, err = unzip(logger, agentZip, unzipDir)
	if err != nil {
		logger.Log(logMessage, "failed to unzip agent dir", logError, err)
		return "", errors.Wrap(err, "failed to unzip agent")
	}
	// run install.sh and enable.sh
	logger.Log(logEvent, "installing agent")
	runErr = runCmd(logger, "bash ./install.sh", agentDirectory, cfg)
	if runErr != nil {
		logger.Log(logMessage, "agent installation failed", logError, runErr)
	} else {
		logger.Log(logMessage, "agent installation succeeded", logEvent, "enabling agent")
		runErr = runCmd(logger, "bash ./enable.sh", agentDirectory, cfg)
		if runErr != nil {
			logger.Log(logMessage, "enable agent failed", logError, runErr)
		} else {
			logger.Log(logMessage, "enable agent succeeded")
		}
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(logger, unzipDir, runErr)

	// TODO: write message for portal
	msg := ""

	return msg, runErr
}

func update(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// parse the extension handler settings
	cfg, err := parseAndValidateSettings(logger, hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	// get old extension path

	// run update.sh to update the agent
	logger.Log(logEvent, "updating agent")
	unzipDir, agentDirectory := unzipAndAgentDirectories()
	runErr := runCmd(logger, "bash ./update.sh", agentDirectory, cfg)
	if runErr != nil {
		// TODO: User doesn't need to know this?
		logger.Log(logMessage, "agent update failed", logError, runErr)
	} else {
		logger.Log(logMessage, "agent update succeeded")
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(logger, unzipDir, runErr)
	// TODO sendTelemetryMessage()

	return "", runErr
}

func disable(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// parse the extension handler settings
	cfg, err := parseAndValidateSettings(logger, hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	// run disable.sh to disable the agent
	logger.Log(logEvent, "disabling agent")
	unzipDir, agentDirectory := unzipAndAgentDirectories()
	runErr := runCmd(logger, "bash ./disable.sh", agentDirectory, cfg)
	if runErr != nil {
		logger.Log(logMessage, "agent disable failed", logError, runErr)
	} else {
		logger.Log(logMessage, "agent disable succeeded")
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(logger, unzipDir, runErr)

	// TODO: change msg to "disable succeeded" or something
	msg := ""

	return msg, nil
}

func uninstall(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// parse the extension handler settings
	cfg, err := parseAndValidateSettings(logger, hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	// run uninstall.sh to uninstall the agent
	logger.Log(logEvent, "uninstalling agent")
	unzipDir, agentDirectory := unzipAndAgentDirectories()
	runErr := runCmd(logger, "bash ./uninstall.sh", agentDirectory, cfg)
	if runErr != nil {
		// TODO: user doesn't need to know (same for disable)
		logger.Log(logMessage, "agent uninstall failed", logError, runErr)
	} else {
		logger.Log(logMessage, "agent uninstall succeeded")
	}

	// collect the logs if available and send telemetry updates
	getStdPipesAndTelemetry(logger, unzipDir, runErr)

	// TODO: fix msg
	msg := ""

	return msg, nil
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// returns the filepaths for the unzip and agent directories
func unzipAndAgentDirectories() (unzipDirectory string, agentDirectory string) {
	unzipDirectory = filepath.Join(dataDir, agentDir)
	agentDirectory = filepath.Join(unzipDirectory, agentName)
	return unzipDirectory, agentDirectory
}

func parseVersionString(agentName string) (version string, err error) {
	r, _ := regexp.Compile("^([./a-zA-Z0-9]*)_([0-9.]*)?[.](.*)$")
	matches := r.FindStringSubmatch(agentName)
	if len(matches) != 4 {
		return "", errors.New("incorrect naming format for agent")
	}
	return matches[2], nil
}

// checkAndSaveSeqNum checks if the given seqNum is already processed
// according to the specified seqNumFile and if so, returns true,
// otherwise saves the given seqNum into seqNumFile returns false.
func checkAndSaveSeqNum(logger log.Logger, seqNum int, mrseqPath string) (shouldExit bool, _ error) {
	logger.Log(logEvent, "comparing seqnum", logPath, mrseqPath)
	smaller, err := seqnum.IsSmallerThan(mrseqPath, seqNum)
	if err != nil {
		return false, errors.Wrap(err, "failed to check sequence number")
	}
	if !smaller {
		// store sequence number is greater than the current sequence number
		return true, nil
	}
	if err := seqnum.Set(mrseqPath, seqNum); err != nil {
		return false, errors.Wrap(err, "failed to save the sequence number")
	}
	logger.Log(logMessage, "seqnum saved", logPath, mrseqPath)

	return false, nil
}

// runCmd runs the command (extracted from cfg) in the given dir (assumed to exist).
func runCmd(logger log.Logger, cmd string, dir string, cfg handlerSettings) (err error) {
	logger.Log(logEvent, "executing command", logOutput, dir)

	begin := time.Now()
	err = ExecCmdInDir(cmd, dir)
	elapsed := time.Now().Sub(begin)
	isSuccess := err == nil

	logger.Log(logMessage, "command executed", "command", cmd, "isSuccess", isSuccess, "time elapsed", elapsed)

	if err != nil {
		logger.Log(logMessage, "failed to execute command", logError, err, logOutput, dir)
		return errors.Wrap(err, "failed to execute command")
	}
	logger.Log(logEvent, "executed command", logOutput, dir)
	return nil
}

// decompresses a zip archive, moving all files and folders within the zip file
// to an output directory
func unzip(logger log.Logger, source string, dest string) ([]string, error) {
	logger.Log(logEvent, "begin unzipping agent")
	var filenames []string
	r, err := zip.OpenReader(source)
	if err != nil {
		return filenames, errors.Wrap(err, "failed to open zip")
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return filenames, errors.Wrap(err, "failed to open file")
		}
		defer rc.Close()

		// store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)
		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// make folder
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			// make file
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, errors.Wrap(err, "failed to create directory")
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, errors.Wrap(err, "failed to open directory at current path")
			}
			_, err = io.Copy(outFile, rc)
			// close the file without defer to close before next iteration of loop
			outFile.Close()
			if err != nil {
				return filenames, errors.Wrap(err, "failed to close file")
			}
		}
	}
	logger.Log(logEvent, "unzipping successful")
	return filenames, nil
}

func getStdPipesAndTelemetry(logger log.Logger, logDir string, runErr error) {
	stdoutF, stderrF := logPaths(logDir)
	stdoutTail, err := tailFile(stdoutF, maxTailLen)
	if err != nil {
		logger.Log(logMessage, "error tailing stdout logs", logError, err)
	}
	stderrTail, err := tailFile(stderrF, maxTailLen)
	if err != nil {
		logger.Log(logMessage, "error tailing stderr logs", logError, err)
	}

	minStdout := min(len(stdoutTail), maxTelemetryTailLen)
	minStderr := min(len(stderrTail), maxTelemetryTailLen)
	msgTelemetry := fmt.Sprintf("\n[stdout]\n%s\n[stderr]\n%s",
		string(stdoutTail[len(stdoutTail)-minStdout:]),
		string(stderrTail[len(stderrTail)-minStderr:]))

	isSuccess := runErr == nil
	telemetry("output", msgTelemetry, isSuccess, 0)
}
