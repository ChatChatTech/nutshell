package nutshell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupBundleDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "Compression Test"
	m.Task.Summary = "Test compression analysis"
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)

	os.MkdirAll(filepath.Join(dir, "context"), 0755)
	os.WriteFile(filepath.Join(dir, "context", "requirements.md"), []byte("# Requirements\nBuild a widget.\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "data.bin"), []byte{0x00, 0x01, 0x02, 0x03, 0xFF}, 0644)
	return dir
}

func TestAnalyzeCompression(t *testing.T) {
	dir := setupBundleDir(t)

	plan, err := AnalyzeCompression(dir)
	if err != nil {
		t.Fatalf("AnalyzeCompression failed: %v", err)
	}
	if len(plan.Files) == 0 {
		t.Fatal("expected files in compression plan")
	}
	if plan.TotalOriginal == 0 {
		t.Fatal("expected non-zero total original size")
	}
	if plan.TextBytes == 0 {
		t.Fatal("expected non-zero text bytes (have .go and .md files)")
	}
	if plan.EstimatedTokens == 0 {
		t.Fatal("expected non-zero token estimate")
	}

	// Verify file categories
	categories := make(map[string]bool)
	for _, f := range plan.Files {
		categories[f.Category] = true
		if f.Path == "" {
			t.Fatal("file strategy has empty path")
		}
	}
	if !categories["text"] {
		t.Fatal("expected at least one 'text' category file")
	}
}

func TestClassifyFile(t *testing.T) {
	tests := []struct {
		path     string
		size     int64
		wantCat  string
	}{
		{"main.go", 100, "text"},
		{"readme.md", 200, "text"},
		{"config.json", 50, "text"},
		{"photo.jpg", 1000, "precompressed"},
		{"archive.zip", 5000, "precompressed"},
		{"data.bin", 500, "binary"},
		{"video.mp4", 10000, "precompressed"},
	}
	for _, tt := range tests {
		s := classifyFile(tt.path, tt.size)
		if s.Category != tt.wantCat {
			t.Errorf("classifyFile(%q) category = %q, want %q", tt.path, s.Category, tt.wantCat)
		}
	}
}

func TestPackWithCompressionNone(t *testing.T) {
	dir := setupBundleDir(t)
	out := filepath.Join(t.TempDir(), "test-none.nut")

	m, plan, err := PackWithCompression(dir, out, CompressNone)
	if err != nil {
		t.Fatalf("PackWithCompression(none) failed: %v", err)
	}
	if m == nil {
		t.Fatal("expected manifest")
	}
	if plan == nil {
		t.Fatal("expected compression plan")
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
}

func TestPackWithCompressionBest(t *testing.T) {
	dir := setupBundleDir(t)
	out := filepath.Join(t.TempDir(), "test-best.nut")

	m, _, err := PackWithCompression(dir, out, CompressBest)
	if err != nil {
		t.Fatalf("PackWithCompression(best) failed: %v", err)
	}
	if m.Files.TotalCount < 3 {
		t.Fatalf("expected at least 3 files, got %d", m.Files.TotalCount)
	}
}

func TestPackWithCompressionRoundtrip(t *testing.T) {
	dir := setupBundleDir(t)
	out := filepath.Join(t.TempDir(), "test-rt.nut")

	_, _, err := PackWithCompression(dir, out, CompressFast)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}

	unpackDir := filepath.Join(t.TempDir(), "unpacked")
	m, err := Unpack(out, unpackDir)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if m.Task.Title != "Compression Test" {
		t.Fatalf("task title mismatch: %q", m.Task.Title)
	}

	// Verify files exist
	if _, err := os.Stat(filepath.Join(unpackDir, "main.go")); err != nil {
		t.Fatal("main.go not found after unpack")
	}
}
