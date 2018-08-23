package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-docker-extension/pkg/vmextension/status"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"strconv"
)

// flags for debugging and printing detailed reports
type flags struct {
	verbose bool
	debug   bool
}

var (
	verbose = flag.Bool("verbose", false, "Return a detailed report")
	debug   = flag.Bool("debug", false, "Return a debug report")

	// dataDir is where we store the downloaded files, logs and state for
	// the extension handler
	dataDir = "./"

	// mrseq holds the processed highest sequence number to make sure
	// we do not run the command more than once for the same sequence
	// number. Stored under dataDir. This file is auto-preserved by the agent.
	mostRecentSequence = "mrseq"

	// agentDir is where the agent is stored
	// stored until dataDir
	agentDir = "GCAgent"

	// agentZip is the directory where the agent package is stored
	// it will be unzipped into {dataDir}/GCAgent/{version}/agent
	agentZip = "agent/DesiredStateConfiguration_1.0.0.zip"

	// agentName contains the .sh files
	// stored under the agent version
	agentName = "DesiredStateConfiguration"

	// the logger that will be used throughout
	lg = newLogger()
)

func main() {
	// parse the command line arguments
	flag.Parse()
	flags := flags{*verbose, *debug}
	cmd := parseCmd(flag.Args())
	lg.with("operation", cmd.name)

	// log flag settings and command name
	lg.customLog(logMessage, "flags settings", "verbose", strconv.FormatBool(flags.verbose),
		"debug", strconv.FormatBool(flags.debug))
	lg.customLog("command name", cmd.name)

	// parse extension environment
	hEnv, err := vmextension.GetHandlerEnv()
	if err != nil {
		lg.messageAndError("failed to parse handlerEnv", err)
		os.Exit(cmd.failExitCode)
	}

	// get sequence number
	seqNum, err := vmextension.FindSeqNum(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		lg.messageAndError("failed to find sequence number", err)
		// only throw a fatal error if the command is not install
		if cmd.name != "install" {
			os.Exit(cmd.failExitCode)
		}
	}
	lg.with("seqNum", strconv.Itoa(seqNum))

	// check sub-command preconditions, if any, before executing
	lg.event("start", "")
	if cmd.pre != nil {
		lg.event("pre-check", "")
		if err := cmd.pre(seqNum); err != nil {
			lg.messageAndError("pre-check failed", err)
			telemetry(telemetryScenario, "enable pre-check failed", false, 0)
			os.Exit(cmd.failExitCode)
		}
	}

	// execute the command
	lg.event("reporting status", "")
	reportStatus(hEnv, seqNum, status.StatusTransitioning, cmd, "")
	msg, err := cmd.f(hEnv, seqNum)
	if err != nil {
		lg.messageAndError("command failed", err)
		reportStatus(hEnv, seqNum, status.StatusError, cmd, err.Error()+msg)
		os.Exit(cmd.failExitCode)
	}
	reportStatus(hEnv, seqNum, status.StatusSuccess, cmd, msg)
	lg.event("end", "")
	os.Exit(successCode)
}

// parseCmd looks at the input array and parses the subcommand. If it is invalid,
// it prints the usage string and an error message and exits with code 2.
func parseCmd(args []string) cmd {
	if len(args) != 1 {
		if len(args) < 1 {
			fmt.Println("Not enough arguments")
		} else {
			fmt.Println("Too many arguments")
		}
		printUsage(args)
		os.Exit(invalidCmdCode)
	}
	// ensure arguments passed are all lower case
	cmd, ok := cmds[strings.ToLower(args[0])]
	if !ok {
		printUsage(args)
		fmt.Printf("Incorrect command: %q\n", args[0])
		os.Exit(invalidCmdCode)
	}
	return cmd
}

// printUsage prints the help string and version of the program to stdout with a
// trailing new line.
func printUsage(args []string) {
	fmt.Printf("Usage: %s ", "main.exe")
	i := 0
	for k := range cmds {
		fmt.Print(k)
		if i != len(cmds)-1 {
			fmt.Printf(" | ")
		}
		i++
	}
	fmt.Println()

	fmt.Println("Optional flags: verbose | debug")
}
