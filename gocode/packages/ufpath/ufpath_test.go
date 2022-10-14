package ufpath

import "testing"

func TestJoin(t *testing.T) {
	if Join("") != "" || Join("a") != "a" || Join("a", "b") != "a/b" {
		t.Fatal("Join")
	}
}

func TestSplit(t *testing.T) {
	tSplit := func(p, d, f string) {
		dr, fr := Split(p)
		if dr != d || fr != f {
			t.Fatalf("Split %s -> '%s', '%s'", p, dr, fr)
		}
	}
	tSplit("", "", "")
	tSplit("a/", "a/", "")
	tSplit("a/b", "a/", "b")
}

func TestExt(t *testing.T) {
	if Ext("a/b.c") != ".c" || Ext("d.d/c") != "" {
		t.Fatal("Ext")
	}
}

func TestBase(t *testing.T) {
	if Base("a/b") != "b" || Base("") != "." || Base("/") != "/" || Base("a/b/") != "b" {
		t.Fatal("Base")
	}
}

func TestDir(t *testing.T) {
	if Dir("") != "." {
		t.Fatal("Dir")
	}
	if Dir("a") != "." {
		t.Fatal("Dir")
	}
	if Dir("/") != "/" {
		t.Fatal("Dir")
	}
	if Dir("a/") != "a" {
		t.Fatalf("Dir %s", Dir("a/"))
	}
	if Dir("a/b") != "a" {
		t.Fatalf("Dir %s", Dir("a/b"))
	}
}
