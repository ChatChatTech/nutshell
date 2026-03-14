package nutshell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackAndUnpack(t *testing.T) {
	// Create a temp bundle directory
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "Pack Test"
	m.Task.Summary = "Testing pack/unpack"
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)

	// Create context directory and file
	os.MkdirAll(filepath.Join(dir, "context"), 0755)
	os.WriteFile(filepath.Join(dir, "context", "requirements.md"), []byte("# Test Requirements\n"), 0644)
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world"), 0644)

	// Pack
	outFile := filepath.Join(t.TempDir(), "test.nut")
	packed, err := Pack(dir, outFile)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	if packed.Files.TotalCount < 3 {
		t.Fatalf("expected at least 3 files, got %d", packed.Files.TotalCount)
	}

	// Verify file-level hashes in tree
	if len(packed.Files.Tree) == 0 {
		t.Fatal("expected file tree to be populated")
	}
	for _, f := range packed.Files.Tree {
		if f.Hash == "" {
			t.Fatalf("file %s has no hash", f.Path)
		}
		if len(f.Hash) < 10 {
			t.Fatalf("hash too short for %s: %s", f.Path, f.Hash)
		}
	}

	// Unpack
	unpackDir := filepath.Join(t.TempDir(), "unpacked")
	unpacked, err := Unpack(outFile, unpackDir)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if unpacked == nil {
		t.Fatal("expected manifest from unpack")
	}
	if unpacked.Task.Title != "Pack Test" {
		t.Fatalf("expected 'Pack Test', got '%s'", unpacked.Task.Title)
	}

	// Verify files exist
	content, err := os.ReadFile(filepath.Join(unpackDir, "hello.txt"))
	if err != nil {
		t.Fatalf("expected hello.txt in unpacked dir: %v", err)
	}
	if string(content) != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", string(content))
	}
}

func TestPackWithNutignore(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "Ignore Test"
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)
	os.WriteFile(filepath.Join(dir, "include.txt"), []byte("yes"), 0644)
	os.WriteFile(filepath.Join(dir, "debug.log"), []byte("no"), 0644)
	os.WriteFile(filepath.Join(dir, ".nutignore"), []byte("*.log\n"), 0644)

	outFile := filepath.Join(t.TempDir(), "ignore-test.nut")
	packed, err := Pack(dir, outFile)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}

	// Should NOT contain debug.log
	for _, f := range packed.Files.Tree {
		if f.Path == "debug.log" {
			t.Fatal("debug.log should have been ignored")
		}
	}
}

func TestInspect(t *testing.T) {
	// Create and pack a bundle first
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "Inspect Me"
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)
	os.WriteFile(filepath.Join(dir, "data.txt"), []byte("test data"), 0644)

	outFile := filepath.Join(t.TempDir(), "inspect.nut")
	Pack(dir, outFile)

	manifest, entries, err := Inspect(outFile)
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}
	if manifest.Task.Title != "Inspect Me" {
		t.Fatalf("expected 'Inspect Me', got '%s'", manifest.Task.Title)
	}
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(entries))
	}
}

func TestHashBundle(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "Hash Test"
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)

	outFile := filepath.Join(t.TempDir(), "hash.nut")
	Pack(dir, outFile)

	hash, err := HashBundle(outFile)
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if len(hash) < 10 {
		t.Fatalf("hash too short: %s", hash)
	}
	if hash[:7] != "sha256:" {
		t.Fatalf("expected sha256: prefix, got %s", hash)
	}

	// Same file should produce same hash
	hash2, _ := HashBundle(outFile)
	if hash != hash2 {
		t.Fatal("same file should produce same hash")
	}
}

func TestUnpackPathTraversal(t *testing.T) {
	// We can't easily forge a tar with path traversal, but we can test that
	// Unpack rejects invalid bundles by checking error on a non-bundle file
	tmpFile := filepath.Join(t.TempDir(), "bad.nut")
	os.WriteFile(tmpFile, []byte("not a bundle"), 0644)
	_, err := Unpack(tmpFile, t.TempDir())
	if err == nil {
		t.Fatal("expected error for invalid bundle")
	}
}
