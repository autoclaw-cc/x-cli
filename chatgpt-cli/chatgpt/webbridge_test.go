package chatgpt

import (
	"strings"
	"testing"
)

func TestSnapshotScriptWaitsForCompleteToolMenuWithoutReclicking(t *testing.T) {
	for _, fragment := range []string{"Date.now()+6000", "labels.length<3", "plus.click()"} {
		if !strings.Contains(SnapshotScript, fragment) {
			t.Fatalf("SnapshotScript missing %q", fragment)
		}
	}
	if strings.Count(SnapshotScript, "plus.click()") != 1 {
		t.Fatalf("SnapshotScript must click the menu exactly once")
	}
}
