package nutshell

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ServeViewer starts a local HTTP server for inspecting a .nut bundle or directory.
// Returns the actual address and a shutdown function.
func ServeViewer(target string, port int) (string, *http.Server, error) {
	// Determine if target is a .nut file or directory
	var manifest *Manifest
	var entries []string
	var srcDir string
	var isBundle bool

	info, err := os.Stat(target)
	if err != nil {
		return "", nil, fmt.Errorf("cannot access %s: %w", target, err)
	}

	if info.IsDir() {
		srcDir = target
		data, err := os.ReadFile(filepath.Join(srcDir, "nutshell.json"))
		if err != nil {
			return "", nil, fmt.Errorf("no nutshell.json in %s", srcDir)
		}
		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return "", nil, fmt.Errorf("invalid nutshell.json: %w", err)
		}
		manifest = &m
		// List files
		filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(srcDir, path)
			entries = append(entries, rel)
			return nil
		})
	} else {
		isBundle = true
		manifest, entries, err = Inspect(target)
		if err != nil {
			return "", nil, err
		}
	}

	mux := http.NewServeMux()

	// API endpoint: manifest JSON
	mux.HandleFunc("/api/manifest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(manifest)
	})

	// API endpoint: file list
	mux.HandleFunc("/api/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})

	// API endpoint: completeness check
	mux.HandleFunc("/api/check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		result := Validate(manifest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":    result.IsValid(),
			"errors":   result.Errors,
			"warnings": result.Warnings,
		})
	})

	// API endpoint: read a specific file (only for directory mode)
	mux.HandleFunc("/api/file/", func(w http.ResponseWriter, r *http.Request) {
		if isBundle {
			http.Error(w, "file reading not available for packed bundles", http.StatusNotImplemented)
			return
		}
		rel := strings.TrimPrefix(r.URL.Path, "/api/file/")
		// Security: prevent path traversal
		clean := filepath.Clean(rel)
		if filepath.IsAbs(clean) || strings.Contains(clean, "..") {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		full := filepath.Join(srcDir, clean)
		data, err := os.ReadFile(full)
		if err != nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(data)
	})

	// Serve the HTML viewer
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, viewerHTML(manifest, entries, isBundle))
	})

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// Try any available port
		ln, err = net.Listen("tcp", "0.0.0.0:0")
		if err != nil {
			return "", nil, err
		}
	}

	actualAddr := ln.Addr().String()
	server := &http.Server{Handler: mux}
	go server.Serve(ln)

	return actualAddr, server, nil
}

func cleanTag(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "[]\"")
	s = strings.TrimSpace(s)
	return s
}

func viewerHTML(m *Manifest, entries []string, isBundle bool) string {
	manifestJSON, _ := json.MarshalIndent(m, "", "  ")
	filesJSON, _ := json.Marshal(entries)
	mode := "directory"
	if isBundle {
		mode = "bundle"
	}

	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>🐚 Nutshell Viewer — ` + escapeHTML(m.Task.Title) + `</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #0d1117; color: #c9d1d9; line-height: 1.6; }
  .header { background: linear-gradient(135deg, #161b22, #1c2333); padding: 2rem; border-bottom: 1px solid #30363d; }
  .header h1 { font-size: 1.5rem; color: #58a6ff; }
  .header .meta { color: #8b949e; font-size: 0.85rem; margin-top: 0.3rem; }
  .container { max-width: 1100px; margin: 0 auto; padding: 1.5rem; }
  .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; margin-top: 1.5rem; }
  @media (max-width: 768px) { .grid { grid-template-columns: 1fr; } }
  .card { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 1.2rem; }
  .card h2 { font-size: 1rem; color: #58a6ff; margin-bottom: 0.8rem; border-bottom: 1px solid #21262d; padding-bottom: 0.5rem; }
  .tag { display: inline-block; background: #1f6feb22; color: #58a6ff; padding: 2px 8px; border-radius: 12px; font-size: 0.8rem; margin: 2px; }
  .field { margin-bottom: 0.5rem; }
  .field .label { color: #8b949e; font-size: 0.8rem; }
  .field .value { color: #c9d1d9; }
  .status-valid { color: #3fb950; }
  .status-warn { color: #d29922; }
  .status-error { color: #f85149; }
  .status-draft { color: #8b949e; }
  .file-list { max-height: 300px; overflow-y: auto; font-family: monospace; font-size: 0.85rem; }
  .file-list div { padding: 3px 0; border-bottom: 1px solid #21262d; cursor: pointer; }
  .file-list div:hover { color: #58a6ff; }
  pre { background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 1rem; overflow-x: auto; font-size: 0.8rem; max-height: 400px; overflow-y: auto; }
  #file-content { display: none; margin-top: 1.5rem; }
  #file-content h3 { color: #58a6ff; margin-bottom: 0.5rem; font-family: monospace; }
  .emoji { font-size: 1.2rem; margin-right: 0.3rem; }
  .checklist { list-style: none; padding: 0; }
  .checklist li { padding: 4px 0; border-bottom: 1px solid #21262d; font-size: 0.9rem; }
  .checklist li::before { content: "☐ "; color: #58a6ff; }
  .constraint { padding: 4px 0; border-bottom: 1px solid #21262d; font-size: 0.9rem; color: #d29922; }
  .constraint::before { content: "⚠ "; }
  .badge { display: inline-block; padding: 2px 10px; border-radius: 12px; font-size: 0.8rem; font-weight: 600; }
  .badge-draft { background: #8b949e33; color: #8b949e; }
  .badge-ready { background: #3fb95033; color: #3fb950; }
  .badge-incomplete { background: #d2992233; color: #d29922; }
</style>
</head>
<body>
<div class="header">
  <div class="container">
    <h1>🐚 Nutshell Viewer</h1>
    <div class="meta">` + escapeHTML(m.Task.Title) + ` • ` + m.BundleType + ` • ` + mode + ` mode` +
		func() string {
			if m.Completeness != nil && m.Completeness.Status != "" {
				return ` • <span class="badge badge-` + m.Completeness.Status + `">` + m.Completeness.Status + `</span>`
			}
			return ""
		}() + `</div>
  </div>
</div>
<div class="container">
  <div class="grid">
    <div class="card">
      <h2><span class="emoji">📋</span>Task</h2>
      <div class="field"><div class="label">Title</div><div class="value">` + escapeHTML(m.Task.Title) + `</div></div>
      <div class="field"><div class="label">Summary</div><div class="value">` + escapeHTML(m.Task.Summary) + `</div></div>
      <div class="field"><div class="label">Priority</div><div class="value">` + m.Task.Priority + `</div></div>
      <div class="field"><div class="label">Effort</div><div class="value">` + m.Task.EstimatedEffort + `</div></div>
    </div>
    <div class="card">
      <h2><span class="emoji">ℹ️</span>Bundle Info</h2>
      <div class="field"><div class="label">Version</div><div class="value">` + m.NutshellVersion + `</div></div>
      <div class="field"><div class="label">Type</div><div class="value">` + m.BundleType + `</div></div>
      <div class="field"><div class="label">ID</div><div class="value" style="font-family:monospace;font-size:0.8rem">` + m.ID + `</div></div>
      <div class="field"><div class="label">Created</div><div class="value">` + m.CreatedAt + `</div></div>` +
		func() string {
			if m.ParentID != "" {
				return `<div class="field"><div class="label">Parent</div><div class="value" style="font-family:monospace;font-size:0.8rem">` + m.ParentID + `</div></div>`
			}
			return ""
		}() + `
    </div>
    <div class="card">
      <h2><span class="emoji">🏷️</span>Tags & Skills</h2>` +
		func() string {
			var s string
			for _, t := range m.Tags.SkillsRequired {
				cleaned := cleanTag(t)
				if cleaned != "" {
					s += `<span class="tag">` + escapeHTML(cleaned) + `</span>`
				}
			}
			if len(m.Tags.Domains) > 0 {
				s += `<div style="margin-top:0.5rem"><div class="label">Domains</div>`
				for _, d := range m.Tags.Domains {
					cleaned := cleanTag(d)
					if cleaned != "" {
						s += `<span class="tag">` + escapeHTML(cleaned) + `</span>`
					}
				}
				s += `</div>`
			}
			if s == "" {
				s = `<div style="color:#8b949e">No tags defined</div>`
			}
			return s
		}() + `
    </div>
    <div class="card">
      <h2><span class="emoji">📦</span>Files (` + fmt.Sprintf("%d", len(entries)) + `)</h2>
      <div class="file-list" id="file-list"></div>
    </div>
  </div>` +
		// Acceptance checklist
		func() string {
			if m.Acceptance != nil && len(m.Acceptance.Checklist) > 0 {
				s := `
  <div class="card" style="margin-top:1.5rem">
    <h2><span class="emoji">✅</span>Acceptance Checklist</h2>
    <ul class="checklist">`
				for _, item := range m.Acceptance.Checklist {
					s += `
      <li>` + escapeHTML(item) + `</li>`
				}
				s += `
    </ul>
  </div>`
				return s
			}
			return ""
		}() +
		// Harness constraints
		func() string {
			if m.Harness != nil {
				s := `
  <div class="card" style="margin-top:1.5rem">
    <h2><span class="emoji">⚙️</span>Agent Harness</h2>`
				if m.Harness.AgentTypeHint != "" {
					s += `
    <div class="field"><div class="label">Agent Type</div><div class="value">` + escapeHTML(m.Harness.AgentTypeHint) + `</div></div>`
				}
				if m.Harness.ExecutionStrategy != "" {
					s += `
    <div class="field"><div class="label">Strategy</div><div class="value">` + escapeHTML(m.Harness.ExecutionStrategy) + `</div></div>`
				}
				if m.Harness.ContextBudgetHint > 0 {
					s += `
    <div class="field"><div class="label">Context Budget</div><div class="value">` + fmt.Sprintf("%.0f%%", m.Harness.ContextBudgetHint*100) + `</div></div>`
				}
				if len(m.Harness.Constraints) > 0 {
					s += `
    <div style="margin-top:0.5rem"><div class="label" style="margin-bottom:0.3rem">Constraints</div>`
					for _, c := range m.Harness.Constraints {
						s += `
    <div class="constraint">` + escapeHTML(c) + `</div>`
					}
					s += `</div>`
				}
				s += `
  </div>`
				return s
			}
			return ""
		}() +
		// Extensions (e.g. clawnet)
		func() string {
			if len(m.Extensions) > 0 {
				s := `
  <div class="card" style="margin-top:1.5rem">
    <h2><span class="emoji">🔌</span>Extensions</h2>`
				for name, raw := range m.Extensions {
					var pretty interface{}
					json.Unmarshal(raw, &pretty)
					prettyJSON, _ := json.MarshalIndent(pretty, "    ", "  ")
					s += `
    <div class="field"><div class="label">` + escapeHTML(name) + `</div></div>
    <pre>` + escapeHTML(string(prettyJSON)) + `</pre>`
				}
				s += `
  </div>`
				return s
			}
			return ""
		}() +
		func() string {
			if !isBundle {
				return `
  <div id="file-content" class="card" style="margin-top:1.5rem">
    <h3 id="file-name"></h3>
    <pre id="file-body"></pre>
  </div>`
			}
			return ""
		}() + `
  <div class="card" style="margin-top:1.5rem">
    <h2><span class="emoji">📄</span>Full Manifest</h2>
    <pre>` + escapeHTML(string(manifestJSON)) + `</pre>
  </div>
</div>
<script>
const files = ` + string(filesJSON) + `;
const isBundle = ` + fmt.Sprintf("%v", isBundle) + `;
const fileList = document.getElementById("file-list");
files.forEach(f => {
  const div = document.createElement("div");
  div.textContent = f;
  if (!isBundle) {
    div.onclick = () => {
      fetch("/api/file/" + encodeURIComponent(f))
        .then(r => r.text())
        .then(text => {
          document.getElementById("file-content").style.display = "block";
          document.getElementById("file-name").textContent = f;
          document.getElementById("file-body").textContent = text;
        });
    };
  }
  fileList.appendChild(div);
});
</script>
</body>
</html>`
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
