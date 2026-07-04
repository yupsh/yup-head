// Command yup-head is the CLI wrapper around github.com/gloo-foo/cmd-head.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-head"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name      = "head"
	flagLines = "lines"
	flagBytes = "bytes"
)

// synopsis is the multi-line --help usage block. urfave/cli indents the whole
// block three spaces, so the lines stay flush-left.
const synopsis = `head [OPTIONS] [FILE...]

Print the first 10 lines of each FILE to standard output.
With no FILE, or when FILE is -, read standard input.`

// spec declares the head wrapper: a file-or-stdin filter with -n line and -c
// byte limits.
var spec = clix.Spec{
	Name:     name,
	Summary:  "output the first part of files",
	Synopsis: synopsis,
	Build:    build,
	Flags: []urf.Flag{
		&urf.IntFlag{
			Name:    flagLines,
			Aliases: []string{"n"},
			Value:   10,
			Usage:   "print the first NUM lines instead of the first 10",
		},
		&urf.IntFlag{Name: flagBytes, Aliases: []string{"c"}, Usage: "print the first NUM bytes"},
	},
}

// build maps the invocation to head's pipeline: a file-or-stdin source into the
// head command configured by the parsed flags.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	return clix.OperandsOrStdin(inv), command.Head(options(inv.Args)...), nil
}

// options translates the selected flags into head's option values. -c/byte mode
// takes precedence over -n/line mode, matching cmd-head's contract.
func options(c *urf.Command) []any {
	if c.IsSet(flagBytes) {
		return []any{command.HeadBytes(c.Int(flagBytes))}
	}
	if c.IsSet(flagLines) {
		return []any{command.HeadLines(c.Int(flagLines))}
	}
	return nil
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
