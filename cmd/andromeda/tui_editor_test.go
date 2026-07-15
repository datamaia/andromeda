package main

import (
	"testing"
)

func TestEditorCommandResolution(t *testing.T) {
	// $VISUAL wins over $EDITOR.
	t.Setenv("VISUAL", "code --wait")
	t.Setenv("EDITOR", "nano")
	if name, args := editorCommand(); name != "code" || len(args) != 1 || args[0] != "--wait" {
		t.Fatalf("VISUAL with flags: name=%q args=%v", name, args)
	}
	// Falls back to $EDITOR.
	t.Setenv("VISUAL", "")
	if name, args := editorCommand(); name != "nano" || len(args) != 0 {
		t.Fatalf("EDITOR: name=%q args=%v", name, args)
	}
	// Default when neither is set.
	t.Setenv("EDITOR", "")
	if name, _ := editorCommand(); name != defaultEditor() {
		t.Fatalf("default editor: %q", name)
	}
}

func TestEditorActionReturnsCommand(t *testing.T) {
	s := &tuiSession{wd: t.TempDir()}
	if cmd := s.editorAction("seed text"); cmd == nil {
		t.Fatal("editorAction should return a runnable command")
	}
}
