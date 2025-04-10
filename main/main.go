package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Azure/azure-docker-extension/pkg/vmextension/status"
	"github.com/Azure/azure-docker-extension/pkg/vmextension"
)

// flags for debugging and printing detailed reports
type flags struct {
	verbose bool
	debug   bool
}

var (
	verbose = flag.Bool("verbose", false, "Return a detailed report")
	debug   = flag.Bool("debug", false, "Return a debug report")

	// the logger that will be used throughout
	lg ExtensionLogger

	// this logger is used only for testing purposes
	noopLogger ExtensionLogger
)

func main() {

	// parse extension environment
	hEnv, handlerErr := vmextension.GetHandlerEnv()
	if handlerErr != nil {
		lg.eventError("Failed to parse handlerEnv file.", handlerErr)
		os.Exit(failureCode)
	}

	// Note that this should be logging to: hEnv.HandlerEnvironment.LogFolder, but
	// The original functionality had this logging at "./path" within the extension
	// directory, and we don't want to break this.
	lg = newLogger(logPath)
	noopLogger = newNoopLogger()
	
	// parse the command line arguments
	flag.Parse()
	cmd := parseCmd(flag.Args())
	lg.with("Operation: ", cmd.name)
	lg.customLog("Command: ", cmd.name)

	seqNum, seqErr := vmextension.FindSeqNum(hEnv.HandlerEnvironment.ConfigFolder)
	if seqErr != nil {
		lg.eventError("failed to find sequence number", seqErr)
		// only throw a fatal error if the command is not "install"
		if cmd.name != "install" {
			os.Exit(cmd.failExitCode)
		}
	}
	lg.event("seqNum: " + strconv.Itoa(seqNum))

	// check sub-command preconditions, if any, before executing
	lg.event("start operation")
	if cmd.pre != nil {
		lg.event("pre-check")
		if preErr := cmd.pre(lg, seqNum); preErr != nil {
			lg.eventError("pre-check failed", preErr)
			telemetry(TelemetryScenario, "enable pre-check failed: "+preErr.Error(), false, 0)
			os.Exit(cmd.failExitCode)
		}
	}

	// execute the command
	lg.event("Reporting transitioning status...")
	reportStatus(lg, hEnv, seqNum, status.StatusTransitioning, cmd, "Transitioning")

	if cmdErr := cmd.f(lg, hEnv, seqNum); cmdErr != nil {
		message := "Operation '" + cmd.name + "' failed."
		lg.eventError(message, cmdErr)
		telemetry(TelemetryScenario, message+" Error: '"+cmdErr.Error()+"'.", false, 0)
		// Never fail on disable due to a current bug in the Guest Agent
		if cmd.name != "disable" {
			reportStatus(lg, hEnv, seqNum, status.StatusError, cmd, cmdErr.Error())
			os.Exit(cmd.failExitCode)
		}
	} else {
		message := "Operation '" + cmd.name + "' succeeded."
		lg.event(message)
		telemetry(TelemetryScenario, message, false, 0)
	}

	reportStatus(lg, hEnv, seqNum, status.StatusSuccess, cmd, "")
	os.Exit(successCode)
}

// parseCmd looks at the input array and parses the subcommand. If it is invalid,
// it prints the usage string and an error message and exits with code 2.
func parseCmd(args []string) cmd {
	if len(args) != 1 {
		if len(args) < 1 {
			fmt.Printf("Not enough arguments, %d", len(args))
			fmt.Println()
			fmt.Printf("%v", args)
			fmt.Println()
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
