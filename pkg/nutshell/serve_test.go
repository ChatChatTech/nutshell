package nutshell

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServeViewerDirectory(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "Serve Test"
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world"), 0644)

	addr, srv, err := ServeViewer(dir, 0)
	if err != nil {
		t.Fatalf("ServeViewer failed: %v", err)
	}
	defer srv.Close()

	base := "http://" + addr

	// Test root page
	resp, err := http.Get(base + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("GET / status %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), "Serve Test") {
		t.Fatal("expected task title in HTML")
	}

	// Test manifest API
	resp, err = http.Get(base + "/api/manifest")
	if err != nil {
		t.Fatalf("GET /api/manifest failed: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("GET /api/manifest status %d", resp.StatusCode)
	}
	var mOut Manifest
	if err := json.Unmarshal(body, &mOut); err != nil {
		t.Fatalf("invalid manifest JSON: %v", err)
	}
	if mOut.Task.Title != "Serve Test" {
		t.Fatalf("manifest title %q", mOut.Task.Title)
	}

	// Test files API
	resp, err = http.Get(base + "/api/files")
	if err != nil {
		t.Fatalf("GET /api/files failed: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	var files []string
	if err := json.Unmarshal(body, &files); err != nil {
		t.Fatalf("invalid files JSON: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one file")
	}

	// Test file content API (directory mode)
	resp, err = http.Get(base + "/api/file/hello.txt")
	if err != nil {
		t.Fatalf("GET /api/file/hello.txt failed: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "hello world" {
		t.Fatalf("file content %q", string(body))
	}
}

func TestServeViewerBundle(t *testing.T) {
	// Create a bundle first
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "Bundle Serve Test"
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)
	os.MkdirAll(filepath.Join(dir, "context"), 0755)
	os.WriteFile(filepath.Join(dir, "context", "requirements.md"), []byte("# Reqs\n"), 0644)

	nutFile := filepath.Join(t.TempDir(), "test.nut")
	_, err := Pack(dir, nutFile)
	if err != nil {
		t.Fatalf("Pack failed: %v", err)
	}

	addr, srv, err := ServeViewer(nutFile, 0)
	if err != nil {
		t.Fatalf("ServeViewer(bundle) failed: %v", err)
	}
	defer srv.Close()

	base := "http://" + addr

	resp, err := http.Get(base + "/api/manifest")
	if err != nil {
		t.Fatalf("GET /api/manifest failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var mOut Manifest
	if err := json.Unmarshal(body, &mOut); err != nil {
		t.Fatalf("invalid manifest JSON: %v", err)
	}
	if mOut.Task.Title != "Bundle Serve Test" {
		t.Fatalf("unexpected title %q", mOut.Task.Title)
	}
}

func TestServeViewerInvalidTarget(t *testing.T) {
	_, _, err := ServeViewer("/nonexistent/path", 0)
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}
