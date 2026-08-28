package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
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
	for i, r := range s {
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

// ── cli ─────────────────────────────────────────────────────────────────────

func runCLI(text string) {
	chars := analyzeText(text)
	if len(chars) == 0 {
		fmt.Fprintln(os.Stderr, "no input")
		os.Exit(1)
	}

	widths := make([]int, len(chars))
	for i, c := range chars {
		widths[i] = imax(len(c.pos), imax(len(c.char), len(c.typ)))
	}

	const labelW = 4
	printRow := func(label string, vals []string) {
		fmt.Printf("%-*s", labelW+2, label)
		for i, v := range vals {
			fmt.Printf("%-*s  ", widths[i], v)
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
	if len(os.Args) > 1 {
		attachConsole()
		runCLI(strings.Join(os.Args[1:], " "))
		return
	}
	if isPiped() {
		attachConsole()
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		runCLI(strings.Join(lines, "\n"))
		return
	}
	runGUI()
}
