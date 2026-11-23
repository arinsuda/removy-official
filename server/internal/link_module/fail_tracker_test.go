package link_module

import "testing"

func TestFailTrackerBlocksAfterMax(t *testing.T) {
	var tr failTracker
	for i := 0; i < verifyFailMax; i++ {
		if !tr.allow("a") {
			t.Fatalf("blocked early at %d", i)
		}
		tr.fail("a")
	}
	if tr.allow("a") {
		t.Fatal("31st attempt should be blocked")
	}
	if !tr.allow("b") {
		t.Fatal("other key must be unaffected")
	}
}
