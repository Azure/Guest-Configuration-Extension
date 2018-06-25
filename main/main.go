package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Add more fields as necessary
type cmd struct {
	name string // human readable string
}

type flags struct {
	verbose bool
	debug   bool
}

var (
	verbose = flag.Bool("verbose", false, "Return a detailed report")
	debug   = flag.Bool("debug", false, "Start debug mode")

	cmds = map[string]cmd{
		"install":   {"install"},
		"uninstall": {"uninstall"},
		"enable":    {"enable"},
		"update":    {"update"},
		"disable":   {"disable"},
	}
)

func main() {
	// parse the command line arguments
	flag.Parse()
	flags := flags{*verbose, *debug}
	cmd := parseCmd(flag.Args())

	fmt.Println("Verbose is " + strconv.FormatBool(flags.verbose))
	fmt.Println("Debug is " + strconv.FormatBool(flags.debug))
	fmt.Println(cmd.name + " extension")
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
