package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/Azure/azure-docker-extension/pkg/vmextension/status"
	"go.uber.org/zap"
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
	lg *zap.SugaredLogger
)

func main() {
	// Initialize Zap logger
	logger, _ := zap.NewProduction() // Use zap.NewDevelopment() for development
	defer logger.Sync()              // Flushes buffer, if any
	lg = logger.Sugar()

	// Log an example startup message
	lg.Infow("Starting application",
		"version", "1.0.0",
		"operation", "main",
	)

	// Parse extension environment
	hEnv, handlerErr := vmextension.GetHandlerEnv()
	if handlerErr != nil {
		lg.Errorw("Failed to parse handlerEnv file.", "error", handlerErr)
		os.Exit(failureCode)
	}

	lg.Infow("Handler environment parsed", "logFolder", hEnv.HandlerEnvironment.LogFolder)

	// Parse the command line arguments
	flag.Parse()
	cmd := parseCmd(flag.Args())
	lg.Infow("Parsed command", "operation", cmd.name)

	seqNum, seqErr := vmextension.FindSeqNum(hEnv.HandlerEnvironment.ConfigFolder)
	if seqErr != nil {
		lg.Errorw("Failed to find sequence number", "error", seqErr)
		if cmd.name != "install" {
			os.Exit(cmd.failExitCode)
		}
	}
	lg.Infow("Sequence number found", "seqNum", seqNum)

	// Check sub-command preconditions, if any, before executing
	lg.Info("Start operation")
	if cmd.pre != nil {
		lg.Info("Running pre-check")
		if preErr := cmd.pre(lg, seqNum); preErr != nil {
			lg.Errorw("Pre-check failed", "error", preErr)
			telemetry(TelemetryScenario, "enable pre-check failed: "+preErr.Error(), false, 0)
			os.Exit(cmd.failExitCode)
		}
	}

	// Execute the command
	lg.Info("Reporting transitioning status...")
	reportStatus(lg, hEnv, seqNum, status.StatusTransitioning, cmd, "Transitioning")

	if cmdErr := cmd.f(lg, hEnv, seqNum); cmdErr != nil {
		message := "Operation '" + cmd.name + "' failed."
		lg.Errorw(message, "error", cmdErr)
		telemetry(TelemetryScenario, message+" Error: '"+cmdErr.Error()+"'.", false, 0)
		if cmd.name != "disable" {
			reportStatus(lg, hEnv, seqNum, status.StatusError, cmd, cmdErr.Error())
			os.Exit(cmd.failExitCode)
		}
	} else {
		message := "Operation '" + cmd.name + "' succeeded."
		lg.Info(message)
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
