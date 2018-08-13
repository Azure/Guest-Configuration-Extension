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

	"github.com/Azure/custom-script-extension-linux/pkg/seqnum"
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
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create data dir")
	}

	logger.Log("event", "created data dir", "path", dataDir)
	logger.Log("event", "installed")

	return "", nil
}

func enablePre(logger log.Logger, seqNum int) error {
	// exit if this sequence number is already processed
	// if not, save the sequence number before proceeding
	if shouldExit, err := checkAndSaveSeqNum(logger, seqNum, mostRecentSequence); err != nil {
		return errors.Wrap(err, "failed to process seqnum")
	} else if shouldExit {
		logger.Log("event", "exit", "message", "this guest configuration is already processed, will not run again")
		// check agent health
		logger.Log("event", "checking agent health")
		agentDirectory := filepath.Join(dataDir, agentDir, agentName)
		shouldExit, err := installAndEnable(logger, agentDirectory)
		if err != nil {
			logger.Log("message", "error running installation and enable scripts", "error", err)
		}
		if shouldExit == true {
			os.Exit(0)
		}
	}
	return nil
}

func installAndEnable(logger log.Logger, agentDirectory string) (bool, error) {
	// check if agent directory exists
	shouldExit := true
	var runErr error = nil
	if _, err := os.Stat(agentDirectory); os.IsNotExist(err) {
		shouldExit = false
	} else {
		// run install.sh and enable.sh from the agent directory
		logger.Log("event", "Running the installation/enabling commands")
		runErr = runCmd(logger, "bash ./install.sh", agentDirectory)
		if runErr != nil {
			logger.Log("message", "error running install.sh", "error", runErr)
		} else {
			runErr = runCmd(logger, "bash ./enable.sh", agentDirectory)
			if runErr != nil {
				logger.Log("message", "error running enable.sh", "error", runErr)
			}
		}
	}
	return shouldExit, runErr
}

func enable(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// parse the extension handler settings (file not available prior to 'enable')
	_, err := parseAndValidateSettings(logger, hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	// parse the version string and log it
	version, err := parseVersionString(agentZip)
	if err != nil {
		logger.Log("message", "failed to parse version string", "error", err, "agentName", agentZip)
		return "", errors.Wrap(err, "failed to parse version string")
	}
	logger.Log("message", "current agent version", "version", version)

	// unzip the agent directory
	dir := filepath.Join(dataDir, agentDir)
	_, err = unzip(logger, agentZip, dir)
	if err != nil {
		logger.Log("message", "failed to unzip agent dir", "error", err)
		return "", errors.Wrap(err, "failed to unzip agent")
	}

	// run install.sh and enable.sh
	agentDirectory := filepath.Join(dir, agentName)
	_, runErr := installAndEnable(logger, agentDirectory)

	// collect the logs if available
	stdoutF, stderrF := logPaths(dir)
	stdoutTail, err := tailFile(stdoutF, maxTailLen)
	if err != nil {
		logger.Log("message", "error tailing stdout logs", "error", err)
	}
	stderrTail, err := tailFile(stderrF, maxTailLen)
	if err != nil {
		logger.Log("message", "error tailing stderr logs", "error", err)
	}

	msg := fmt.Sprintf("\n[stdout]\n%s\n[stderr]\n%s", string(stdoutTail), string(stderrTail))

	minStdout := min(len(stdoutTail), maxTelemetryTailLen)
	minStderr := min(len(stderrTail), maxTelemetryTailLen)
	msgTelemetry := fmt.Sprintf("\n[stdout]\n%s\n[stderr]\n%s",
		string(stdoutTail[len(stdoutTail)-minStdout:]),
		string(stderrTail[len(stderrTail)-minStderr:]))

	isSuccess := runErr == nil
	telemetry("Output", msgTelemetry, isSuccess, 0)

	if isSuccess {
		logger.Log("event", "enabled")
	} else {
		logger.Log("event", "enable failed")
	}
	return msg, runErr
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
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
	logger.Log("event", "comparing seqnum", "path", mrseqPath)
	smaller, err := seqnum.IsSmallerThan(mrseqPath, seqNum)
	if err != nil {
		return false, errors.Wrap(err, "failed to check sequence number")
	}
	if !smaller {
		// store sequence number is equal or greater than the current sequence number
		return true, nil
	}
	if err := seqnum.Set(mrseqPath, seqNum); err != nil {
		return false, errors.Wrap(err, "failed to save the sequence number")
	}
	logger.Log("event", "seqnum saved", "path", mrseqPath)

	return false, nil
}

// runCmd runs the command (extracted from cfg) in the given dir (assumed to exist).
func runCmd(logger log.Logger, cmd string, dir string) (err error) {
	logger.Log("event", "executing command", "output", dir)
	var scenario string

	begin := time.Now()
	err = ExecCmdInDir(cmd, dir)
	elapsed := time.Now().Sub(begin)
	isSuccess := err == nil

	telemetry("scenario", scenario, isSuccess, elapsed)

	if err != nil {
		logger.Log("event", "failed to execute command", "error", err, "output", dir)
		return errors.Wrap(err, "failed to execute command")
	}
	logger.Log("event", "executed command", "output", dir)
	return nil
}

// decompresses a zip archive, moving all files and folders within the zip file
// to an output directory
func unzip(logger log.Logger, source string, dest string) ([]string, error) {
	logger.Log("event", "begin unzipping agent")
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
	logger.Log("event", "unzipping successful")
	return filenames, nil
}

func update(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	return "", nil
}

func disable(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	return "", nil
}

func uninstall(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	return "", nil
}
