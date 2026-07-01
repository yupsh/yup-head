package main

import (
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-head"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const name = "head"

const (
	flagLines = "lines"
	flagBytes = "bytes"
)

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `head [OPTIONS] [FILE...]

Print the first 10 lines of each FILE to standard output.
With no FILE, or when FILE is -, read standard input.`

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags while still exposing
// the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// run builds and executes the head CLI against the injected version, I/O, and
// filesystem, returning the process exit code.
func run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, fs afero.Fs) int {
	cmd := newApp(version, stdin, stdout, fs)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, name+": %v\n", err)
		return 1
	}
	return 0
}

func newApp(version string, stdin io.Reader, stdout io.Writer, fs afero.Fs) *cli.Command {
	return &cli.Command{
		Name:            name,
		Version:         version,
		Usage:           "output the first part of files",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    flagLines,
				Aliases: []string{"n"},
				Value:   10,
				Usage:   "print the first NUM lines instead of the first 10",
			},
			&cli.IntFlag{Name: flagBytes, Aliases: []string{"c"}, Usage: "print the first NUM bytes"},
		},
		Action: action(stdin, stdout, fs),
	}
}

func action(stdin io.Reader, stdout io.Writer, fs afero.Fs) cli.ActionFunc {
	return func(_ context.Context, c *cli.Command) error {
		_, err := gloo.Run(source(c, stdin, fs), gloo.ByteWriteTo(stdout), command.Head(options(c)...))
		return err
	}
}

func source(c *cli.Command, stdin io.Reader, fs afero.Fs) any {
	if c.NArg() == 0 {
		return gloo.ByteReaderSource([]io.Reader{stdin})
	}
	files := make([]gloo.File, c.NArg())
	for i := range files {
		files[i] = gloo.File(c.Args().Get(i))
	}
	return gloo.ByteFileSource(fs, files)
}

// options translates the selected flags into constructor option values. -c/byte
// mode takes precedence over -n/line mode, matching cmd-head's contract.
func options(c *cli.Command) []any {
	if c.IsSet(flagBytes) {
		return []any{command.HeadBytes(c.Int(flagBytes))}
	}
	if c.IsSet(flagLines) {
		return []any{command.HeadLines(c.Int(flagLines))}
	}
	return nil
}
