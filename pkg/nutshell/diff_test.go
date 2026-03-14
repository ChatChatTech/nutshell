package nutshell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDiffSameManifest(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "Same Task"
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)

	diffs, err := Diff(dir, dir)
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if len(diffs) != 0 {
		t.Fatalf("expected no diffs for same dir, got %d: %v", len(diffs), diffs)
	}
}

func TestDiffDifferentManifests(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	mA := NewManifest()
	mA.Task.Title = "Request Task"
	mA.BundleType = "request"
	dataA, _ := json.MarshalIndent(mA, "", "  ")
	os.WriteFile(filepath.Join(dirA, "nutshell.json"), dataA, 0644)

	mB := NewManifest()
	mB.Task.Title = "Delivery Task"
	mB.BundleType = "delivery"
	mB.ParentID = mA.ID
	dataB, _ := json.MarshalIndent(mB, "", "  ")
	os.WriteFile(filepath.Join(dirB, "nutshell.json"), dataB, 0644)

	diffs, err := Diff(dirA, dirB)
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected diffs between different manifests")
	}

	// Should have title, bundle_type, id, parent_id diffs at minimum
	fields := make(map[string]bool)
	for _, d := range diffs {
		fields[d.Field] = true
	}
	for _, expected := range []string{"task.title", "bundle_type", "parent_id"} {
		if !fields[expected] {
			t.Errorf("expected diff for field %s", expected)
		}
	}
}

func TestDiffBundleFiles(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	mA := NewManifest()
	mA.Task.Title = "A"
	mA.BundleType = "request"
	dataA, _ := json.MarshalIndent(mA, "", "  ")
	os.WriteFile(filepath.Join(dirA, "nutshell.json"), dataA, 0644)

	mB := NewManifest()
	mB.Task.Title = "A"
	mB.BundleType = "request"
	dataB, _ := json.MarshalIndent(mB, "", "  ")
	os.WriteFile(filepath.Join(dirB, "nutshell.json"), dataB, 0644)

	// Pack both
	nutA := filepath.Join(t.TempDir(), "a.nut")
	nutB := filepath.Join(t.TempDir(), "b.nut")
	Pack(dirA, nutA)
	Pack(dirB, nutB)

	diffs, err := Diff(nutA, nutB)
	if err != nil {
		t.Fatalf("diff on .nut files failed: %v", err)
	}
	// IDs will differ, but titles are same
	for _, d := range diffs {
		if d.Field == "task.title" {
			t.Fatal("titles should be same")
		}
	}
}
