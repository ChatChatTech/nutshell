package nutshell

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SubTask describes one piece of a split task.
type SubTask struct {
	Title       string   `json:"title"`
	Summary     string   `json:"summary,omitempty"`
	Skills      []string `json:"skills,omitempty"`
	Files       []string `json:"files,omitempty"` // file patterns to include in this sub-bundle
	Constraints []string `json:"constraints,omitempty"`
}

// SplitPlan describes how to split a task into parallel sub-tasks.
type SplitPlan struct {
	SubTasks []SubTask `json:"sub_tasks"`
}

// SplitResult holds information about one generated sub-bundle.
type SplitResult struct {
	Index     int    `json:"index"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	Directory string `json:"directory"`
}

// Split breaks a task directory into N sub-task directories based on a plan.
// If plan is nil, it auto-splits by file directories.
func Split(srcDir string, plan *SplitPlan) ([]SplitResult, error) {
	data, err := os.ReadFile(filepath.Join(srcDir, "nutshell.json"))
	if err != nil {
		return nil, fmt.Errorf("no nutshell.json in %s: %w", srcDir, err)
	}
	var parent Manifest
	if err := json.Unmarshal(data, &parent); err != nil {
		return nil, fmt.Errorf("invalid nutshell.json: %w", err)
	}

	if plan == nil {
		plan = autoSplitPlan(srcDir, &parent)
	}
	if len(plan.SubTasks) == 0 {
		return nil, fmt.Errorf("split plan has no sub-tasks")
	}

	var results []SplitResult
	parentID := parent.ID

	for i, sub := range plan.SubTasks {
		m := parent // shallow copy
		m.ID = GenerateID()
		m.ParentID = parentID
		m.BundleType = "partial"
		m.Task.Title = sub.Title
		if sub.Summary != "" {
			m.Task.Summary = sub.Summary
		}
		if len(sub.Skills) > 0 {
			m.Tags.SkillsRequired = sub.Skills
		}

		// Merge constraints
		if m.Harness == nil {
			m.Harness = &Harness{}
		}
		if len(sub.Constraints) > 0 {
			m.Harness.Constraints = append(m.Harness.Constraints, sub.Constraints...)
		}

		// Add split metadata extension
		if m.Extensions == nil {
			m.Extensions = make(map[string]json.RawMessage)
		}
		splitMeta, _ := json.Marshal(map[string]interface{}{
			"parent_id":  parentID,
			"part_index": i,
			"part_total": len(plan.SubTasks),
		})
		m.Extensions["split"] = splitMeta

		// Create sub-directory
		slug := strings.ToLower(sub.Title)
		slug = strings.ReplaceAll(slug, " ", "-")
		if len(slug) > 30 {
			slug = slug[:30]
		}
		subDir := filepath.Join(filepath.Dir(srcDir), fmt.Sprintf("%s-part%d-%s", filepath.Base(srcDir), i, slug))
		if err := os.MkdirAll(filepath.Join(subDir, "context"), 0755); err != nil {
			return nil, err
		}

		// Copy matching files from source to sub-directory
		copied := 0
		if len(sub.Files) > 0 {
			for _, pattern := range sub.Files {
				matches, _ := filepath.Glob(filepath.Join(srcDir, pattern))
				for _, match := range matches {
					rel, _ := filepath.Rel(srcDir, match)
					if rel == "nutshell.json" {
						continue
					}
					dst := filepath.Join(subDir, rel)
					os.MkdirAll(filepath.Dir(dst), 0755)
					if copyFile(match, dst) == nil {
						copied++
					}
				}
			}
		}

		// Always copy context/ into sub-bundles
		contextDir := filepath.Join(srcDir, "context")
		if info, err := os.Stat(contextDir); err == nil && info.IsDir() {
			filepath.WalkDir(contextDir, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				rel, _ := filepath.Rel(srcDir, path)
				dst := filepath.Join(subDir, rel)
				os.MkdirAll(filepath.Dir(dst), 0755)
				copyFile(path, dst)
				return nil
			})
		}

		// Write sub-manifest
		mData, _ := json.MarshalIndent(&m, "", "  ")
		os.WriteFile(filepath.Join(subDir, "nutshell.json"), mData, 0644)

		results = append(results, SplitResult{
			Index:     i,
			ID:        m.ID,
			Title:     sub.Title,
			Directory: subDir,
		})
	}

	return results, nil
}

// Merge combines multiple delivery sub-bundles back into one delivery bundle.
func Merge(dirs []string, outDir string) (*Manifest, error) {
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no directories to merge")
	}

	// Load first manifest as the base
	data, err := os.ReadFile(filepath.Join(dirs[0], "nutshell.json"))
	if err != nil {
		return nil, fmt.Errorf("no nutshell.json in %s: %w", dirs[0], err)
	}
	var merged Manifest
	if err := json.Unmarshal(data, &merged); err != nil {
		return nil, err
	}

	// Get the parent_id to use as the merged bundle's parent
	parentID := merged.ParentID

	merged.ID = GenerateID()
	merged.ParentID = parentID
	merged.BundleType = "delivery"

	// Remove split extension from merged
	delete(merged.Extensions, "split")

	// Create output directory
	if err := os.MkdirAll(filepath.Join(outDir, "context"), 0755); err != nil {
		return nil, err
	}

	// Merge files from all sub-bundles
	seenFiles := map[string]bool{}
	for i, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(dir, path)
			if rel == "nutshell.json" {
				return nil
			}
			// If same file exists from multiple sub-bundles, prefix with part index
			dst := filepath.Join(outDir, rel)
			if seenFiles[rel] {
				ext := filepath.Ext(rel)
				base := strings.TrimSuffix(rel, ext)
				dst = filepath.Join(outDir, fmt.Sprintf("%s-part%d%s", base, i, ext))
			}
			seenFiles[rel] = true
			os.MkdirAll(filepath.Dir(dst), 0755)
			copyFile(path, dst)
			return nil
		})
	}

	// Write merged manifest
	mData, _ := json.MarshalIndent(&merged, "", "  ")
	os.WriteFile(filepath.Join(outDir, "nutshell.json"), mData, 0644)

	return &merged, nil
}

// autoSplitPlan generates a split plan by grouping files into top-level directories.
func autoSplitPlan(srcDir string, m *Manifest) *SplitPlan {
	// Group files by top-level directory
	groups := map[string][]string{}
	filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(srcDir, path)
		if rel == "nutshell.json" || rel == ".nutignore" || strings.HasPrefix(rel, "context/") {
			return nil
		}
		parts := strings.SplitN(rel, string(os.PathSeparator), 2)
		group := "root"
		if len(parts) > 1 {
			group = parts[0]
		}
		groups[group] = append(groups[group], rel)
		return nil
	})

	if len(groups) <= 1 {
		// Not enough structure to split — create 2 halves
		var allFiles []string
		for _, files := range groups {
			allFiles = append(allFiles, files...)
		}
		mid := len(allFiles) / 2
		if mid == 0 {
			mid = 1
		}
		return &SplitPlan{
			SubTasks: []SubTask{
				{Title: m.Task.Title + " (Part 1)", Files: allFiles[:mid]},
				{Title: m.Task.Title + " (Part 2)", Files: allFiles[mid:]},
			},
		}
	}

	var subs []SubTask
	for group, files := range groups {
		patterns := []string{group + "/**"}
		subs = append(subs, SubTask{
			Title: fmt.Sprintf("%s — %s", m.Task.Title, group),
			Files: patterns,
		})
		_ = files
	}
	return &SplitPlan{SubTasks: subs}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
