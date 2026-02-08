package main

import (
	"reflect"
	"testing"
)

func TestSplitArgsSimple(t *testing.T) {
	got := splitArgs("foo bar baz")
	want := []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitArgs(\"foo bar baz\") = %v, want %v", got, want)
	}
}

func TestSplitArgsQuoted(t *testing.T) {
	got := splitArgs(`foo "bar baz" qux`)
	want := []string{"foo", "bar baz", "qux"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitArgs with quotes = %v, want %v", got, want)
	}
}

func TestSplitArgsMakeArgs(t *testing.T) {
	got := splitArgs(`etl-analyze-laws ARGS="--limit 1"`)
	want := []string{"etl-analyze-laws", "ARGS=--limit 1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitArgs with ARGS= = %v, want %v", got, want)
	}
}

func TestSplitArgsEmpty(t *testing.T) {
	got := splitArgs("")
	if got != nil {
		t.Errorf("splitArgs(\"\") = %v, want nil", got)
	}

	got = splitArgs("   ")
	if got != nil {
		t.Errorf("splitArgs(\"   \") = %v, want nil", got)
	}
}

func TestSplitArgsSingleArg(t *testing.T) {
	got := splitArgs("hello")
	want := []string{"hello"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitArgs(\"hello\") = %v, want %v", got, want)
	}
}
