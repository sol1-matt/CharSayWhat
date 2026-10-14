package main

import (
	"reflect"
	"testing"
)

func TestClassifyRune(t *testing.T) {
	tests := []struct {
		r    rune
		want string
	}{
		{'a', "lowercase"},
		{'é', "lowercase"},
		{'Z', "uppercase"},
		{'Ö', "uppercase"},
		{'0', "digit"},
		{'٣', "digit"},
		{' ', "space"},
		{'\n', "newline"},
		{'\t', "tab"},
		{'!', "punctuation"},
		{'@', "punctuation"},
		{'$', "symbol"},
		{'+', "symbol"},
		{' ', "other"},
		{'​', "other"},
	}
	for _, tt := range tests {
		if got := classifyRune(tt.r); got != tt.want {
			t.Errorf("classifyRune(%q) = %q, want %q", tt.r, got, tt.want)
		}
	}
}

func TestDisplayRune(t *testing.T) {
	tests := []struct {
		r    rune
		want string
	}{
		{' ', "[sp]"},
		{'\n', "[nl]"},
		{'\t', "[tb]"},
		{'x', "x"},
		{'é', "é"},
	}
	for _, tt := range tests {
		if got := displayRune(tt.r); got != tt.want {
			t.Errorf("displayRune(%q) = %q, want %q", tt.r, got, tt.want)
		}
	}
}

func TestAnalyzeText(t *testing.T) {
	tests := []struct {
		in   string
		want []charInfo
	}{
		{"", nil},
		{"a1", []charInfo{
			{"1", "a", "lowercase"},
			{"2", "1", "digit"},
		}},
		// Positions count characters, not bytes: é is two bytes in UTF-8.
		{"héllo", []charInfo{
			{"1", "h", "lowercase"},
			{"2", "é", "lowercase"},
			{"3", "l", "lowercase"},
			{"4", "l", "lowercase"},
			{"5", "o", "lowercase"},
		}},
		{"€ x", []charInfo{
			{"1", "€", "symbol"},
			{"2", "[sp]", "space"},
			{"3", "x", "lowercase"},
		}},
	}
	for _, tt := range tests {
		if got := analyzeText(tt.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("analyzeText(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestColumnWidths(t *testing.T) {
	tests := []struct {
		in   string
		want []int
	}{
		{"a", []int{9}}, // "lowercase"
		{"é", []int{9}}, // counted in characters, not bytes
		{"5", []int{5}}, // "digit"
		{"€", []int{6}}, // "symbol"
		{" ", []int{5}}, // "space" beats "[sp]"
	}
	for _, tt := range tests {
		got := columnWidths(analyzeText(tt.in))
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("columnWidths(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		args    []string
		names   map[string]string
		help    bool
		text    []string
		wantErr bool
	}{
		{args: nil},
		{args: []string{"P@ss"}, text: []string{"P@ss"}},
		{args: []string{"-s", "P@ss"}, names: symbolNames, text: []string{"P@ss"}},
		{args: []string{"-c", "a", "b"}, names: codeNames, text: []string{"a", "b"}},
		{args: []string{"-c"}, names: codeNames},
		{args: []string{"-h"}, help: true},
		// Text starting with "-" is analysed, not rejected.
		{args: []string{"-x1"}, text: []string{"-x1"}},
		// Switches after the text are part of the text.
		{args: []string{"abc", "-s"}, text: []string{"abc", "-s"}},
		{args: []string{"--", "-s"}, text: []string{"-s"}},
		{args: []string{"-s", "--", "-c"}, names: symbolNames, text: []string{"-c"}},
		{args: []string{"-s", "-c", "x"}, wantErr: true},
	}
	for _, tt := range tests {
		got, err := parseArgs(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseArgs(%q) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			continue
		}
		if tt.wantErr {
			continue
		}
		if !reflect.DeepEqual(got.names, tt.names) || got.help != tt.help || !reflect.DeepEqual(got.text, tt.text) {
			t.Errorf("parseArgs(%q) = %+v, want names=%v help=%v text=%q", tt.args, got, tt.names, tt.help, tt.text)
		}
	}
}

// Every classification must have a short name in both styles, or -s/-c would
// print an empty Type cell.
func TestShortNamesCoverAllTypes(t *testing.T) {
	types := []string{"lowercase", "uppercase", "digit", "space", "newline", "tab", "punctuation", "symbol", "other"}
	for _, m := range []map[string]string{symbolNames, codeNames} {
		if len(m) != len(types) {
			t.Errorf("short-name map has %d entries, want %d", len(m), len(types))
		}
		for _, typ := range types {
			if _, ok := m[typ]; !ok {
				t.Errorf("short-name map missing %q", typ)
			}
		}
	}
}
