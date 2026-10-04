package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("validate-skills", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: validate-skills [repository]")
		fmt.Fprintln(stderr, "Validate skill bundles, metadata, local references, and Bash/Go syntax.")
		fmt.Fprintln(stderr, "The repository defaults to the current directory. Bash and ShellCheck must be on PATH.")
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() > 1 {
		flags.Usage()
		return 2
	}
	root := "."
	if flags.NArg() == 1 {
		root = flags.Arg(0)
	}
	root, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	count, problems := validate(root)
	for _, problem := range problems {
		fmt.Fprintf(stderr, "error: %s\n", problem)
	}
	if len(problems) > 0 {
		return 1
	}
	fmt.Fprintf(stdout, "Validated %d skills, agent metadata, local references, and script syntax.\n", count)
	return 0
}
