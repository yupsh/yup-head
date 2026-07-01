package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		files      map[string]string
		name       string
		version    string
		stdin      string
		wantOut    string
		wantErrSub string
		args       []string
		wantCode   int
	}{
		{
			name:    "default count",
			args:    []string{"head"},
			stdin:   "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n",
			wantOut: "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n",
		},
		{
			name:    "explicit lines",
			args:    []string{"head", "-n", "3"},
			stdin:   "a\nb\nc\nd\ne\n",
			wantOut: "a\nb\nc\n",
		},
		{
			name:    "file source",
			args:    []string{"head", "-n", "2", "/in.txt"},
			files:   map[string]string{"/in.txt": "one\ntwo\nthree\n"},
			wantOut: "one\ntwo\n",
		},
		{
			// Byte mode emits the leading NUM bytes as a single value; the
			// []byte sink terminates that value with a newline.
			name:    "byte count takes the leading bytes",
			args:    []string{"head", "-c", "5"},
			stdin:   "hello world\n",
			wantOut: "hello\n",
		},
		{
			name:    "byte mode overrides line mode",
			args:    []string{"head", "-c", "3", "-n", "1"},
			stdin:   "abcdef\nghij\n",
			wantOut: "abc\n",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"head", "--version"},
			wantOut: "head version 1.2.3\n",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"head", "--nope"},
			wantCode:   1,
			wantErrSub: "head:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for path, content := range tc.files {
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Fatalf("write fixture %s: %v", path, err)
				}
			}

			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(tc.stdin), &out, &errOut, fs)

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
