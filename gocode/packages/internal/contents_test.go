package internal

import "testing"

func TestNs2Content(t *testing.T) {
	c1, s1, s31 := Ns2Content([]string{"a", "b", "c", "d/"}, "l")
	c2, s2, s32 := Ns2Content([]string{"c", "d/", "b", "a"}, "l")
	if c1 != c2 || s1 != s2 || s31 != s32 {
		t.Fatalf("Ns2Content %s %s differ from %s %s", c1, s1, c2, s2)
	}
}
