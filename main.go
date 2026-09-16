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
	Flags:    flags(),
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

// flags builds a FRESH flag set on every call.
//
// urfave/cli flags carry mutable parse state: once a flag is supplied, its
// hasBeenSet stays true for the life of the value. A package-level slice shared
// across two parses therefore reports the second parse as having flags only the
// FIRST one set, and IsSet stops meaning "the user asked for this".
//
// The shared gate runs every suite with -count=2 in one process, which is
// exactly that shape — the first pass sets the flag, the second sees it already
// set. Building fresh keeps one definition while giving each parse its own
// state.
func flags() []urf.Flag {
	return []urf.Flag{
		&urf.IntFlag{
			Name:    flagLines,
			Aliases: []string{"n"},
			Value:   10,
			Usage:   "print the first NUM lines instead of the first 10",
			Sources: urf.EnvVars("YUP_HEAD_LINES"),
		},
		&urf.IntFlag{
			Name:    flagBytes,
			Aliases: []string{"c"},
			Usage:   "print the first NUM bytes",
			Value:   0,
			Sources: urf.EnvVars("YUP_HEAD_BYTES"),
		},
	}
}
