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

func TestReadyScriptWaitsForHydratedComposerControls(t *testing.T) {
	if !strings.Contains(ReadyScript, "querySelectorAll('button').length>=3") {
		t.Fatalf("ReadyScript does not wait for hydrated composer controls")
	}
}
