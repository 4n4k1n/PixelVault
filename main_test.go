package main

import "testing"

func TestParseUsers(t *testing.T) {
	got := parseUsers(" 123:anakin , 456 : bob ")
	if got[123] != "anakin" || got[456] != "bob" || len(got) != 2 {
		t.Fatalf("got %v", got)
	}
	if len(parseUsers("")) != 0 || len(parseUsers("nope,12x:a")) != 0 {
		t.Fatal("garbage must not authorize anyone")
	}
	if parseUsers("123:../../etc")[123] != "etc" {
		t.Fatal("folder must not escape baseDir")
	}
}
