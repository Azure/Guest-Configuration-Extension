package main

import (
	"os"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type cmdfunc func(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error)

// Add more fields as necessary
type cmd struct {
	f                  cmdfunc // associated function
	name               string  // human readable string
	shouldReportStatus bool    // determines if running this should log to a .status file
	failExitCode       int     // exitCode to use when commands fail
}

var (
	// allowed user inputs
	cmds = map[string]cmd{
		"install":   {install, "install", false, 52},
		"uninstall": {uninstall, "uninstall", true, 3},
		"enable":    {enable, "enable", true, 3},
		"update":    {update, "update", true, 3},
		"disable":   {disable, "disable", true, 3},
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

func enable(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	// pre-check for enable
	// check that sequence number has not already been processed
	// if not, save the sequence number
	// still need to check exit cases
	if shouldExit, err := checkAndSaveSeqNum(logger, seqNum, mostRecentSequence); err != nil {
		return "", errors.Wrap(err, "failed to process seqnum")
	} else if shouldExit {
		logger.Log("event", "exit", "message", "this guest configuration is already processed, will not run again")
	}

	// parse the extension handler settings
	// returns config but this is not needed
	_, err := parseAndValidateSettings(logger, hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		return "", errors.Wrap(err, "failed to get configuration")
	}

	logger.Log("event", "enabled")

	return "", nil
}

// checkAndSaveSeqNum checks if the given seqNum is already processed
// according to the specified seqNumFile and if so, returns true,
// otherwise saves the given seqNum into seqNumFile returns false.
func checkAndSaveSeqNum(logger log.Logger, seqNum int, meseqPath string) (shouldExit bool, _ error) {

	return true, nil
}

func update(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	return "", nil
}

func disable(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	return "", nil
}
