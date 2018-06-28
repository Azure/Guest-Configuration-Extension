package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Azure/azure-docker-extension/pkg/vmextension"
	"github.com/go-kit/kit/log"
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
	dataDir = "/var/lib/waagent/guest-configuration"

	// mrseq holds the processed highest sequence number to make sure
	// we do not run the command more than once for the same sequence
	// number. Stored under dataDir. This file is auto-preserved by the agent.
	mostRecentSequence = "mrseq"

	// downloadDir is where we store the downloaded files in the "{downloadDir}/{seqnum}/file"
	// format and the logs as "{downloadDir}/{seqnum}/std(out|err)". Stored under dataDir
	downloadDir = "download"
)

func main() {
	logger := log.With(log.With(log.NewSyncLogger(log.NewLogfmtLogger(
		os.Stdout)), "time", log.DefaultTimestamp), "version", VersionString())

	// parse the command line arguments
	flag.Parse()
	flags := flags{*verbose, *debug}
	cmd := parseCmd(flag.Args())
	logger = log.With(logger, "operation", cmd.name)

	// print flags and command name
	fmt.Println("Verbose is " + strconv.FormatBool(flags.verbose))
	fmt.Println("Debug is " + strconv.FormatBool(flags.debug))
	fmt.Println(cmd.name + " extension")

	// parse extension environment
	// currently have an issue with this when I execute
	hEnv, err := vmextension.GetHandlerEnv()
	if err != nil {
		logger.Log("message", "failed to parse handlerenv", "error", err)
		os.Exit(cmd.failExitCode)
	}

	// get sequence number (should we be throwing a fatal error here?)
	seqNum, err := vmextension.FindSeqNum(hEnv.HandlerEnvironment.ConfigFolder)
	if err != nil {
		logger.Log("messsage", "failed to find sequence number", "error", err)
	}
	logger = log.With(logger, "seq", seqNum)

	// execute the command
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
		os.Exit(2)
	}
	// ensure arguments passed are all lower case
	cmd, ok := cmds[strings.ToLower(args[0])]
	if !ok {
		printUsage(args)
		fmt.Printf("Incorrect command: %q\n", args[0])
		os.Exit(2)
	}
	return cmd
}

// printUsage prints the help string and version of the program to stdout with a
// trailing new line.
func printUsage(args []string) {
	fmt.Printf("Usage: %s ", "main.exe")
	i := 0
	for k := range cmds {
		fmt.Printf(k)
		if i != len(cmds)-1 {
			fmt.Printf("|")
		}
		i++
	}
	fmt.Println()

	fmt.Println("Optional flags: verbose | debug")
}
