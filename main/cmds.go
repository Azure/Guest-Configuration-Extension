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
	f            cmdfunc // associated function
	name         string  // human readable string
	failExitCode int     // exitCode to use when commands fail
}

var (

	// allowed user inputs
	cmds = map[string]cmd{
		"install":   {install, "install", 52},
		"uninstall": {uninstall, "uninstall", 3},
		"enable":    {enable, "enable", 3},
		"update":    {update, "update", 3},
		"disable":   {disable, "disable", 3},
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
	return "", nil
}

func update(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	return "", nil
}

func disable(logger log.Logger, hEnv vmextension.HandlerEnvironment, seqNum int) (string, error) {
	return "", nil
}
