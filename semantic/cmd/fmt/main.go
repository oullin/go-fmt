package main

import (
	"fmt"
	"io"
	"os"

	"github.com/oullin/go-fmt/semantic/cli"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)

		return 1
	}

	switch args[0] {
	case "check":
		return cli.
			NewRunner(stdout, stderr).
			Run(cli.CheckMode, args[1:])
	case "format":
		return cli.
			NewRunner(stdout, stderr).
			Run(cli.FormatMode, args[1:])
	case "version", "--version", "-version":
		fmt.Printf("go-fmt %s\n", version)

		return 0
	case "help", "--help", "-h":
		printUsage(stderr)

		return 0
	default:
		fmt.Printf("unknown subcommand - {%q}\n\n", args[0])
		printUsage(stderr)

		return 1
	}
}

func printUsage(w io.Writer) {
	fmt.Printf("go-fmt check [--host-path /absolute/host/path] [paths...] - {%v}\n\n", w)
	fmt.Printf("go-fmt format [--host-path /absolute/host/path] [paths...] - {%v}\n\n", w)
}
