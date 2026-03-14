package nutshell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupSplitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "Split Test Task"
	m.Task.Summary = "A task to be split"
	m.Tags.SkillsRequired = []string{"go", "python"}
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)

	os.MkdirAll(filepath.Join(dir, "context"), 0755)
	os.WriteFile(filepath.Join(dir, "context", "requirements.md"), []byte("# Reqs\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("package main\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "tests"), 0755)
	os.WriteFile(filepath.Join(dir, "tests", "test.go"), []byte("package tests\n"), 0644)
	return dir
}

func TestSplitAutoplan(t *testing.T) {
	dir := setupSplitDir(t)

	results, err := Split(dir, nil)
	if err != nil {
		t.Fatalf("Split(auto) failed: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 sub-tasks, got %d", len(results))
	}

	for _, r := range results {
		if r.ID == "" {
			t.Fatal("sub-task missing ID")
		}
		if r.Title == "" {
			t.Fatal("sub-task missing title")
		}
		if r.Directory == "" {
			t.Fatal("sub-task missing directory")
		}
		// Verify sub-task directory has a manifest
		mPath := filepath.Join(r.Directory, "nutshell.json")
		data, err := os.ReadFile(mPath)
		if err != nil {
			t.Fatalf("sub-task manifest not found: %v", err)
		}
		var sub Manifest
		if err := json.Unmarshal(data, &sub); err != nil {
			t.Fatalf("invalid sub-task manifest: %v", err)
		}
		if sub.BundleType != "partial" {
			t.Fatalf("expected bundle_type 'partial', got %q", sub.BundleType)
		}
		if sub.ParentID == "" {
			t.Fatal("sub-task missing parent_id")
		}
	}
}

func TestSplitWithPlan(t *testing.T) {
	dir := setupSplitDir(t)

	plan := &SplitPlan{
		SubTasks: []SubTask{
			{Title: "Backend", Skills: []string{"go"}, Files: []string{"src/*"}},
			{Title: "Testing", Skills: []string{"go"}, Files: []string{"tests/*"}},
		},
	}

	results, err := Split(dir, plan)
	if err != nil {
		t.Fatalf("Split(plan) failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 sub-tasks, got %d", len(results))
	}
	if results[0].Title != "Backend" {
		t.Fatalf("expected 'Backend', got %q", results[0].Title)
	}
	if results[1].Title != "Testing" {
		t.Fatalf("expected 'Testing', got %q", results[1].Title)
	}
}

func TestMerge(t *testing.T) {
	dir := setupSplitDir(t)

	results, err := Split(dir, nil)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}

	// Add delivery content to each sub-task
	for i, r := range results {
		content := []byte("delivery content " + string(rune('A'+i)))
		os.WriteFile(filepath.Join(r.Directory, "output.txt"), content, 0644)
	}

	// Merge
	mergeDir := filepath.Join(t.TempDir(), "merged")
	dirs := make([]string, len(results))
	for i, r := range results {
		dirs[i] = r.Directory
	}

	m, err := Merge(dirs, mergeDir)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}
	if m == nil {
		t.Fatal("expected merged manifest")
	}
	if m.BundleType != "delivery" {
		t.Fatalf("expected 'delivery', got %q", m.BundleType)
	}
	// Verify merged manifest exists
	if _, err := os.Stat(filepath.Join(mergeDir, "nutshell.json")); err != nil {
		t.Fatal("merged manifest not found")
	}
}
