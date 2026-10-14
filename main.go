package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ── shared ──────────────────────────────────────────────────────────────────

type charInfo struct {
	pos  string
	char string
	typ  string
}

func classifyRune(r rune) string {
	switch {
	case unicode.IsLower(r):
		return "lowercase"
	case unicode.IsUpper(r):
		return "uppercase"
	case unicode.IsDigit(r):
		return "digit"
	case r == ' ':
		return "space"
	case r == '\n':
		return "newline"
	case r == '\t':
		return "tab"
	case unicode.IsPunct(r):
		return "punctuation"
	case unicode.IsSymbol(r):
		return "symbol"
	default:
		return "other"
	}
}

func displayRune(r rune) string {
	switch r {
	case ' ':
		return "[sp]"
	case '\n':
		return "[nl]"
	case '\t':
		return "[tb]"
	default:
		return string(r)
	}
}

func analyzeText(s string) []charInfo {
	var result []charInfo
	for i, r := range []rune(s) {
		result = append(result, charInfo{
			pos:  fmt.Sprintf("%d", i+1),
			char: displayRune(r),
			typ:  classifyRune(r),
		})
	}
	return result
}

func imax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// columnWidths returns the display width of each grid column, counted in
// characters rather than bytes so multi-byte runes don't skew alignment.
func columnWidths(chars []charInfo) []int {
	widths := make([]int, len(chars))
	for i, c := range chars {
		widths[i] = imax(utf8.RuneCountInString(c.pos),
			imax(utf8.RuneCountInString(c.char), utf8.RuneCountInString(c.typ)))
	}
	return widths
}

// ── cli ─────────────────────────────────────────────────────────────────────

// symbolNames (-s) names each type by an example character.
var symbolNames = map[string]string{
	"lowercase":   "a",
	"uppercase":   "A",
	"digit":       "9",
	"space":       " ",
	"newline":     "\\n",
	"tab":         "\\t",
	"punctuation": ".",
	"symbol":      "$",
	"other":       "?",
}

// codeNames (-c) names each type by a two-letter code.
var codeNames = map[string]string{
	"lowercase":   "lc",
	"uppercase":   "uc",
	"digit":       "di",
	"space":       "sp",
	"newline":     "nl",
	"tab":         "tb",
	"punctuation": "pu",
	"symbol":      "sy",
	"other":       "ot",
}

const usage = `usage: charsaywhat [-s | -c] [--] [text...]

  (no arguments)  open the GUI
  text...         print the characters of text (joined with spaces)
  -s              short type names by example: a A 9 $ . and a space
  -c              short type names by code:    lc uc di sy pu sp
  --              treat everything after as text, e.g. charsaywhat -- -s
  -h              show this help

With no text argument, text is read from stdin if it is piped.`

type cliOptions struct {
	names map[string]string // nil means full type names
	help  bool
	text  []string
}

// parseArgs reads leading switches and returns the rest as text. Parsing stops
// at the first argument that isn't a known switch, so text that happens to
// start with "-" (common in passwords) is analysed rather than rejected.
func parseArgs(args []string) (cliOptions, error) {
	var opts cliOptions
	for i, a := range args {
		switch a {
		case "-s", "-c":
			if opts.names != nil {
				return opts, fmt.Errorf("-s and -c can't be combined")
			}
			opts.names = symbolNames
			if a == "-c" {
				opts.names = codeNames
			}
		case "-h", "--help":
			opts.help = true
		case "--":
			opts.text = args[i+1:]
			return opts, nil
		default:
			opts.text = args[i:]
			return opts, nil
		}
	}
	return opts, nil
}

func runCLI(text string, names map[string]string) {
	chars := analyzeText(text)
	if len(chars) == 0 {
		fmt.Fprintln(os.Stderr, "no input")
		os.Exit(1)
	}
	if names != nil {
		for i := range chars {
			chars[i].typ = names[chars[i].typ]
		}
	}

	widths := columnWidths(chars)

	const labelW = 4
	printRow := func(label string, vals []string) {
		fmt.Printf("%-*s", labelW+2, label)
		for i, v := range vals {
			fmt.Print(v, strings.Repeat(" ", widths[i]-utf8.RuneCountInString(v)+2))
		}
		fmt.Println()
	}

	pos := make([]string, len(chars))
	chs := make([]string, len(chars))
	typ := make([]string, len(chars))
	for i, c := range chars {
		pos[i] = c.pos
		chs[i] = c.char
		typ[i] = c.typ
	}

	fmt.Println()
	printRow("Pos", pos)
	printRow("Char", chs)
	printRow("Type", typ)
	fmt.Println()
}

// ── main ────────────────────────────────────────────────────────────────────

func isPiped() bool {
	stat, err := os.Stdin.Stat()
	return err == nil && stat.Mode()&os.ModeCharDevice == 0
}

func main() {
	if len(os.Args) == 1 && !isPiped() {
		if prepareGUI() {
			runGUI()
		}
		return
	}

	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	if opts.help {
		fmt.Println(usage)
		return
	}
	if len(opts.text) > 0 {
		runCLI(strings.Join(opts.text, " "), opts.names)
		return
	}
	if !isPiped() {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	runCLI(strings.Join(lines, "\n"), opts.names)
}
