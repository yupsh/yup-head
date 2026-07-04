#!/bin/sh
# Integration checks for yup-head, run inside a Debian (GNU coreutils) container.
#
# parity_stdin INPUT ARGS...  — yup-head reading stdin must match GNU `head`.
# parity_file  ARGS...        — yup-head reading file operands must match GNU.
# assert WANT  ARGS...        — yup-head must produce WANT exactly (used where
#                               yup-head diverges from GNU by design; see
#                               cmd-head COMPATIBILITY.md).
set -eu

fails=0
sample='1
2
3
4
5
6
7
8
9
10
11
12'

parity_stdin() {
	in=$1
	shift
	ours=$(printf '%s\n' "$in" | yup-head "$@" 2>/dev/null || true)
	gnu=$(printf '%s\n' "$in" | head "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  head %s < stdin\n' "$*"
	else
		printf 'FAIL  parity  head %s < stdin\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

parity_file() {
	ours=$(yup-head "$@" 2>/dev/null || true)
	gnu=$(head "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  head %s\n' "$*"
	else
		printf 'FAIL  parity  head %s\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

assert() {
	want=$1
	shift
	got=$(yup-head "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  head %s\n' "$*"
	else
		printf 'FAIL  assert  head %s\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# assert_bytes WANT ARGS... — assert yup-head's exact bytes (reading the
# `bytes_stdin` global) equal WANT, with a trailing sentinel appended to both so
# command substitution cannot strip a trailing newline. Used where the
# divergence is precisely that trailing byte.
assert_bytes() {
	want=$1
	shift
	got=$(printf '%s' "$bytes_stdin" | yup-head "$@" 2>/dev/null; printf X)
	if [ "$got" = "${want}X" ]; then
		printf 'ok    assert  head %s (raw bytes)\n' "$*"
	else
		printf 'FAIL  assert  head %s (raw bytes)\n        want: %sX\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# Default: first 10 lines. Line mode matches GNU byte-for-byte.
parity_stdin "$sample"
# -n: explicit line counts, including more lines than the input has.
parity_stdin "$sample" -n 3
parity_stdin "$sample" -n 1
parity_stdin "$sample" -n 100

# Single file operand: matches GNU (no header for a lone file).
printf '%s\n' "$sample" > /tmp/a.txt
printf 'one\ntwo\nthree\n' > /tmp/b.txt
parity_file -n 2 /tmp/a.txt

# Divergence (byte mode trailing newline): yup-head emits the leading NUM bytes
# as one value, which the []byte sink terminates with a newline; GNU `head -c 5`
# of "hello world\n" emits exactly "hello" (5 bytes, no added newline) whereas
# yup-head emits "hello\n" (6 bytes). Compare raw bytes so the trailing newline
# is not stripped by command substitution.
bytes_stdin='hello world
'
# WANT is the 6 bytes "hello\n" (a literal newline inside the quotes).
assert_bytes 'hello
' -c 5

# Divergence (multiple file operands): GNU head treats each file independently —
# it prints `-n N` lines OF EACH file, prefixed with a `==> NAME <==` header. The
# gloo framework instead concatenates all operands into ONE byte stream, so
# yup-head emits the first N lines of the CONCATENATION, with no headers. For
# a.txt (1..12) + b.txt (one,two,three), `-n 2` yields just the first two lines
# of a.txt; b.txt never contributes.
assert "$(printf '1\n2')" -n 2 /tmp/a.txt /tmp/b.txt

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
