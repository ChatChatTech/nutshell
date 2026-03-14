package nutshell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenBundle(t *testing.T) {
	// Create and pack a bundle
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "SDK Test"
	m.Task.Summary = "Testing the SDK"
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)
	os.MkdirAll(filepath.Join(dir, "context"), 0755)
	os.WriteFile(filepath.Join(dir, "context", "requirements.md"), []byte("# Reqs\nDo the thing.\n"), 0644)
	os.WriteFile(filepath.Join(dir, "src.go"), []byte("package main\n"), 0644)

	outFile := filepath.Join(t.TempDir(), "sdk-test.nut")
	Pack(dir, outFile)

	// Open
	b, err := Open(outFile)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	// Manifest
	if b.Manifest().Task.Title != "SDK Test" {
		t.Fatalf("expected 'SDK Test', got '%s'", b.Manifest().Task.Title)
	}

	// ListFiles
	files := b.ListFiles()
	if len(files) < 3 {
		t.Fatalf("expected at least 3 files, got %d: %v", len(files), files)
	}

	// ReadFile
	content, err := b.ReadFile("src.go")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(content) != "package main\n" {
		t.Fatalf("unexpected content: %s", string(content))
	}

	// ReadContext
	ctx, err := b.ReadContext()
	if err != nil {
		t.Fatalf("ReadContext failed: %v", err)
	}
	if len(ctx) == 0 {
		t.Fatal("expected non-empty context")
	}

	// ReadFileString
	s, err := b.ReadFileString("src.go")
	if err != nil {
		t.Fatalf("ReadFileString failed: %v", err)
	}
	if s != "package main\n" {
		t.Fatalf("unexpected: %s", s)
	}

	// HasFile
	if !b.HasFile("nutshell.json") {
		t.Fatal("should have nutshell.json")
	}
	if b.HasFile("nonexistent.txt") {
		t.Fatal("should not have nonexistent.txt")
	}

	// FilesByPrefix
	ctxFiles := b.FilesByPrefix("context/")
	if len(ctxFiles) != 1 {
		t.Fatalf("expected 1 context file, got %d: %v", len(ctxFiles), ctxFiles)
	}

	// ManifestJSON
	mjson, err := b.ManifestJSON()
	if err != nil {
		t.Fatalf("ManifestJSON failed: %v", err)
	}
	if len(mjson) == 0 {
		t.Fatal("expected non-empty manifest JSON")
	}

	// ReadFile not found
	_, err = b.ReadFile("no-such-file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestOpenInvalidBundle(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad.nut")
	os.WriteFile(tmpFile, []byte("not a bundle"), 0644)
	_, err := Open(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid bundle")
	}
}

func TestOpenNonexistentFile(t *testing.T) {
	_, err := Open("/tmp/does-not-exist-12345.nut")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
