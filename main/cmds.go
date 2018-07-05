package main

import (
	"os"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"

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
	pre 			   prefunc // executed before any status is reported
	failExitCode       int     // exitCode to use when commands fail
}

var (
	// allowed user inputs
	cmds = map[string]cmd{
		"install":   {install, "install", false, nil, 52},
		"uninstall": {uninstall, "uninstall", false, nil, 3},
		"enable":    {enable, "enable", true, enablePre, 3},
		"update":    {update, "update", true, nil, 3},
		"disable":   {disable, "disable", true, nil, 3},
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

func uninstall(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	return "", nil
}

func enablePre(logger log.Logger, seqNum int) error {
	// exit if this sequence number is already processed
	// if not, save the sequence number before proceeding
	if shouldExit, err := checkAndSaveSeqNum(logger, seqNum, mostRecentSequence); err != nil {
		return errors.Wrap(err, "failed to process seqnum")
	} else if shouldExit {
		logger.Log("event", "exit", "message", "this guest configuration is already processed, will not run again")
	}
	return nil
}

func enable(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// parse the extension handler settings (file not available prior to 'enable')
	_, err := parseAndValidateSettings(logger, hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	// unzip the agent

	// run agent scripts

	// collect and output available logs

	logger.Log("event", "enabled")

	return "", nil
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
	logger.Log("event", "seqnum saved", "path", mrseqPath)

	return false, nil
}

func update(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	return "", nil
}

func disable(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	return "", nil
}
