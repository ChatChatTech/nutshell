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

const nutshellIconSVG = `<svg viewBox="-4.32 -4.32 32.64 32.64" xmlns="http://www.w3.org/2000/svg" fill="#ffffff" stroke="#ffffff"><g stroke-width="0"><path transform="translate(-4.32,-4.32) scale(1.02)" d="M16,29.85C18.89,29.84,22.02,30.72,24.46,29.17C26.9,27.61,27.57,24.46,28.67,21.78C29.72,19.21,30.94,16.65,30.73,13.88C30.5,11.01,29.35,8.25,27.46,6.07C25.57,3.9,22.93,2.61,20.18,1.76C17.4,0.9,14.37,0.16,11.66,1.21C9.01,2.24,7.77,5.11,5.93,7.28C4.13,9.41,1.67,11.13,0.99,13.84C0.29,16.67,1.04,19.63,2.18,22.31C3.34,25.04,4.91,27.79,7.51,29.21C10.05,30.59,13.1,29.86,16,29.85" fill="#165DFF"/></g><g stroke-width="1.2" fill="none" fill-rule="evenodd" stroke-linecap="round" stroke="#ffffff"><path d="M19.1,9.24L19.51,9.39C20.8,9.86,21.47,11.29,21,12.58C20.89,12.89,20.72,13.17,20.5,13.41L20.34,13.59C18.79,15.29,16.99,16.75,15,17.91L12.05,19.63C12.02,19.65,11.98,19.65,11.95,19.63L9,17.91C7.01,16.75,5.21,15.29,3.66,13.59L3.5,13.41C2.57,12.39,2.65,10.81,3.67,9.89C3.91,9.67,4.19,9.5,4.49,9.39L4.9,9.24M8.69,15.25L5.5,10.99C4.5,9.66,4.77,7.78,6.1,6.79C6.36,6.59,6.64,6.44,6.95,6.34L7.5,6.16C7.88,6.03,8.27,5.92,8.65,5.83L8.84,5.78C8.51,6.39,8.41,7.11,8.61,7.81L10.45,14.24M13.55,14.24L15.39,7.81C15.59,7.11,15.49,6.39,15.16,5.78C15.61,5.89,16.06,6.01,16.5,6.16L17.05,6.34C18.62,6.86,19.47,8.56,18.95,10.14C18.85,10.44,18.7,10.73,18.5,10.99L15.31,15.25M10.45,14.24L8.61,7.81C8.26,6.56,8.84,5.24,10,4.66C11.26,4.03,12.74,4.03,14,4.66C15.16,5.24,15.74,6.56,15.39,7.81L13.55,14.24"/><path d="M17,16.59L17,18.5C17,19.33,16.33,20,15.5,20L8.5,20C7.67,20,7,19.33,7,18.5L7,16.59C7.64,17.07,8.31,17.5,9,17.91L11.95,19.63C11.98,19.65,12.02,19.65,12.05,19.63L15,17.91C15.69,17.5,16.36,17.07,17,16.59Z"/></g></svg>`

func viewerHTML(m *Manifest, entries []string, isBundle bool) string {
	manifestJSON, _ := json.MarshalIndent(m, "", "  ")
	filesJSON, _ := json.Marshal(entries)
	mode := "directory"
	if isBundle {
		mode = "bundle"
	}

	// Build tags HTML
	tagsHTML := ""
	for _, t := range m.Tags.SkillsRequired {
		cleaned := cleanTag(t)
		if cleaned != "" {
			tagsHTML += `<span class="tag">` + escapeHTML(cleaned) + `</span>`
		}
	}
	if len(m.Tags.Domains) > 0 {
		for _, d := range m.Tags.Domains {
			cleaned := cleanTag(d)
			if cleaned != "" {
				tagsHTML += `<span class="tag tag-domain">` + escapeHTML(cleaned) + `</span>`
			}
		}
	}
	if tagsHTML == "" {
		tagsHTML = `<span class="muted">No tags defined</span>`
	}

	// Build acceptance HTML
	acceptanceHTML := ""
	if m.Acceptance != nil && len(m.Acceptance.Checklist) > 0 {
		acceptanceHTML = `<div class="card full"><div class="card-header"><svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg><span>Acceptance Checklist</span></div><div class="card-body"><ul class="checklist">`
		for _, item := range m.Acceptance.Checklist {
			acceptanceHTML += `<li>` + escapeHTML(item) + `</li>`
		}
		acceptanceHTML += `</ul></div></div>`
	}

	// Build harness HTML
	harnessHTML := ""
	if m.Harness != nil {
		harnessHTML = `<div class="card full"><div class="card-header"><svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg><span>Agent Harness</span></div><div class="card-body">`
		if m.Harness.AgentTypeHint != "" {
			harnessHTML += `<div class="field"><span class="field-label">Agent Type</span><span class="field-value">` + escapeHTML(m.Harness.AgentTypeHint) + `</span></div>`
		}
		if m.Harness.ExecutionStrategy != "" {
			harnessHTML += `<div class="field"><span class="field-label">Strategy</span><span class="field-value">` + escapeHTML(m.Harness.ExecutionStrategy) + `</span></div>`
		}
		if m.Harness.ContextBudgetHint > 0 {
			harnessHTML += `<div class="field"><span class="field-label">Context Budget</span><span class="field-value">` + fmt.Sprintf("%.0f%%", m.Harness.ContextBudgetHint*100) + `</span></div>`
		}
		if len(m.Harness.Constraints) > 0 {
			harnessHTML += `<div class="constraints-label">Constraints</div><ul class="constraints">`
			for _, c := range m.Harness.Constraints {
				harnessHTML += `<li>` + escapeHTML(c) + `</li>`
			}
			harnessHTML += `</ul>`
		}
		harnessHTML += `</div></div>`
	}

	// Build extensions HTML
	extensionsHTML := ""
	if len(m.Extensions) > 0 {
		extensionsHTML = `<div class="card full"><div class="card-header"><svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M4 17l6-6-6-6"/><path d="M12 19h8"/></svg><span>Extensions</span></div><div class="card-body">`
		for name, raw := range m.Extensions {
			var pretty interface{}
			json.Unmarshal(raw, &pretty)
			prettyJSON, _ := json.MarshalIndent(pretty, "", "  ")
			extensionsHTML += `<div class="ext-name">` + escapeHTML(name) + `</div><pre>` + escapeHTML(string(prettyJSON)) + `</pre>`
		}
		extensionsHTML += `</div></div>`
	}

	// Completeness
	statusBadge := ""
	if m.Completeness != nil && m.Completeness.Status != "" {
		cls := "badge-muted"
		switch m.Completeness.Status {
		case "ready":
			cls = "badge-ready"
		case "incomplete":
			cls = "badge-warn"
		}
		statusBadge = `<span class="badge ` + cls + `">` + m.Completeness.Status + `</span>`
	}

	// Parent ID field
	parentHTML := ""
	if m.ParentID != "" {
		parentHTML = `<div class="field"><span class="field-label">Parent</span><span class="field-value mono">` + m.ParentID + `</span></div>`
	}

	// File content viewer (directory mode only)
	fileContentHTML := ""
	if !isBundle {
		fileContentHTML = `<div id="file-content" class="card full" style="display:none"><div class="card-header"><svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg><span id="file-name">File</span></div><div class="card-body"><pre id="file-body"></pre></div></div>`
	}

	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Nutshell Viewer — ` + escapeHTML(m.Task.Title) + `</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
<style>
:root {
  --white: #FFFFFF;
  --bg: #FBFBFA;
  --border: #E8E8E4;
  --text: #37352F;
  --text-secondary: #9B9A97;
  --blue: #165DFF;
  --blue-hover: #0D4DD6;
  --blue-subtle: rgba(22,93,255,0.06);
  --blue-dim: rgba(22,93,255,0.15);
  --font-body: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-mono: 'JetBrains Mono', Menlo, Consolas, monospace;
  --r-sm: 4px;
  --r-md: 8px;
  --r-lg: 12px;
}
*, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: var(--font-body); background: var(--bg); color: var(--text); line-height: 1.6; -webkit-font-smoothing: antialiased; }

/* ── Nav ── */
.nav { position: sticky; top: 0; z-index: 100; background: rgba(255,255,255,0.92); backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px); border-bottom: 1px solid var(--border); }
.nav-inner { max-width: 1080px; margin: 0 auto; padding: 0 2rem; height: 56px; display: flex; align-items: center; justify-content: space-between; }
.nav-brand { display: flex; align-items: center; gap: 0; font-size: 1.1rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; text-decoration: none; }
.nav-brand svg { margin-right: 6px; width: 24px; height: 24px; flex-shrink: 0; }
.nav-brand .logo-shell { color: var(--blue); }
.nav-meta { display: flex; align-items: center; gap: 12px; font-size: 0.8125rem; color: var(--text-secondary); }
.badge { display: inline-block; padding: 2px 10px; border-radius: 12px; font-size: 0.75rem; font-weight: 600; letter-spacing: 0.02em; text-transform: uppercase; }
.badge-muted { background: #E8E8E4; color: #9B9A97; }
.badge-ready { background: rgba(46,160,67,0.12); color: #1a7f37; }
.badge-warn { background: rgba(210,153,34,0.12); color: #9a6700; }
.badge-type { background: var(--blue-subtle); color: var(--blue); }

/* ── Hero ── */
.hero { background: linear-gradient(180deg, var(--white) 20%, #F4F7FF 100%); padding: 3rem 2rem 2rem; border-bottom: 1px solid var(--border); }
.hero-inner { max-width: 1080px; margin: 0 auto; }
.hero h1 { font-size: 1.75rem; font-weight: 700; letter-spacing: -0.02em; line-height: 1.2; color: var(--text); margin-bottom: 0.5rem; }
.hero-summary { color: var(--text-secondary); font-size: 0.95rem; line-height: 1.6; max-width: 720px; }
.hero-badges { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 1rem; }

/* ── Container ── */
.container { max-width: 1080px; margin: 0 auto; padding: 2rem; }

/* ── Grid ── */
.grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 1px; background: var(--border); border-radius: var(--r-lg); overflow: hidden; margin-bottom: 1.5rem; }
@media (max-width: 768px) { .grid { grid-template-columns: 1fr; } }

/* ── Card ── */
.card { background: var(--white); }
.card.full { border-radius: var(--r-lg); border: 1px solid var(--border); overflow: hidden; margin-bottom: 1.5rem; }
.card-header { display: flex; align-items: center; gap: 8px; padding: 0.875rem 1.25rem; border-bottom: 1px solid var(--border); background: var(--bg); font-size: 0.8125rem; font-weight: 600; color: var(--text); letter-spacing: -0.01em; }
.card-icon { width: 16px; height: 16px; color: var(--blue); flex-shrink: 0; }
.card-body { padding: 1.25rem; }
.card-accent { width: 24px; height: 2px; background: var(--blue); margin-bottom: 1rem; }
.card-title { font-size: 0.8125rem; font-weight: 600; color: var(--text); margin-bottom: 1rem; letter-spacing: -0.01em; display: flex; align-items: center; gap: 8px; }
.card-title svg { width: 16px; height: 16px; color: var(--blue); }

/* ── Fields ── */
.field { display: flex; justify-content: space-between; align-items: baseline; padding: 6px 0; border-bottom: 1px solid var(--border); font-size: 0.875rem; }
.field:last-child { border-bottom: none; }
.field-label { color: var(--text-secondary); font-size: 0.8125rem; }
.field-value { color: var(--text); font-weight: 500; }
.field-value.mono { font-family: var(--font-mono); font-size: 0.75rem; word-break: break-all; max-width: 60%; text-align: right; }

/* ── Tags ── */
.tag { display: inline-block; background: var(--blue-subtle); color: var(--blue); padding: 3px 10px; border-radius: 4px; font-size: 0.8125rem; font-weight: 500; margin: 2px 4px 2px 0; }
.tag-domain { background: rgba(46,160,67,0.08); color: #1a7f37; }
.muted { color: var(--text-secondary); font-size: 0.875rem; }

/* ── Checklist ── */
.checklist { list-style: none; padding: 0; }
.checklist li { padding: 8px 0; border-bottom: 1px solid var(--border); font-size: 0.875rem; display: flex; align-items: flex-start; gap: 10px; line-height: 1.5; }
.checklist li:last-child { border-bottom: none; }
.checklist li::before { content: ''; display: inline-block; width: 16px; height: 16px; min-width: 16px; border: 1.5px solid var(--border); border-radius: var(--r-sm); margin-top: 2px; }

/* ── Constraints ── */
.constraints-label { font-size: 0.8125rem; font-weight: 600; color: var(--text); margin-top: 0.75rem; margin-bottom: 0.5rem; }
.constraints { list-style: none; padding: 0; }
.constraints li { padding: 6px 0 6px 22px; border-bottom: 1px solid var(--border); font-size: 0.875rem; position: relative; color: var(--text); line-height: 1.5; }
.constraints li:last-child { border-bottom: none; }
.constraints li::before { content: ''; position: absolute; left: 0; top: 12px; width: 8px; height: 8px; border-radius: 2px; background: #d29922; }

/* ── Extension ── */
.ext-name { font-family: var(--font-mono); font-size: 0.8125rem; font-weight: 500; color: var(--blue); margin-bottom: 0.5rem; }

/* ── Files ── */
.file-list { max-height: 320px; overflow-y: auto; }
.file-item { padding: 8px 0; border-bottom: 1px solid var(--border); font-family: var(--font-mono); font-size: 0.8125rem; color: var(--text); cursor: pointer; display: flex; align-items: center; gap: 8px; transition: color 0.15s; }
.file-item:last-child { border-bottom: none; }
.file-item:hover { color: var(--blue); }
.file-item svg { width: 14px; height: 14px; color: var(--text-secondary); flex-shrink: 0; }
.file-item:hover svg { color: var(--blue); }

/* ── Pre/Code ── */
pre { font-family: var(--font-mono); font-size: 0.8125rem; line-height: 1.6; background: var(--bg); border: 1px solid var(--border); border-radius: var(--r-md); padding: 1rem 1.25rem; overflow-x: auto; max-height: 440px; overflow-y: auto; color: var(--text); }

/* ── Manifest toggle ── */
.manifest-toggle { display: flex; align-items: center; gap: 8px; padding: 0.875rem 1.25rem; background: var(--bg); border-bottom: 1px solid var(--border); font-size: 0.8125rem; font-weight: 600; color: var(--text); cursor: pointer; user-select: none; letter-spacing: -0.01em; }
.manifest-toggle:hover { background: #F4F4F2; }
.manifest-toggle svg { width: 16px; height: 16px; color: var(--blue); transition: transform 0.2s; }
.manifest-toggle.open svg { transform: rotate(90deg); }
.manifest-body { display: none; padding: 1.25rem; }
.manifest-body.open { display: block; }
</style>
</head>
<body>

<nav class="nav">
  <div class="nav-inner">
    <a class="nav-brand" href="#">` + nutshellIconSVG + `nut<span class="logo-shell">shell</span></a>
    <div class="nav-meta">
      <span class="badge badge-type">` + escapeHTML(m.BundleType) + `</span>
      <span>` + mode + `</span>
      ` + statusBadge + `
    </div>
  </div>
</nav>

<div class="hero">
  <div class="hero-inner">
    <h1>` + escapeHTML(m.Task.Title) + `</h1>` +
		func() string {
			if m.Task.Summary != "" {
				return `<p class="hero-summary">` + escapeHTML(m.Task.Summary) + `</p>`
			}
			return ""
		}() + `
    <div class="hero-badges">` + tagsHTML + `</div>
  </div>
</div>

<div class="container">

  <!-- Primary info grid -->
  <div class="grid">
    <div class="card">
      <div class="card-body">
        <div class="card-title"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>Task</div>
        <div class="field"><span class="field-label">Priority</span><span class="field-value">` + escapeHTML(m.Task.Priority) + `</span></div>
        <div class="field"><span class="field-label">Effort</span><span class="field-value">` + escapeHTML(m.Task.EstimatedEffort) + `</span></div>` +
		func() string {
			if m.ExpiresAt != "" {
				return `<div class="field"><span class="field-label">Expires</span><span class="field-value">` + escapeHTML(m.ExpiresAt) + `</span></div>`
			}
			return ""
		}() + `
      </div>
    </div>
    <div class="card">
      <div class="card-body">
        <div class="card-title"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>Bundle</div>
        <div class="field"><span class="field-label">Version</span><span class="field-value">` + escapeHTML(m.NutshellVersion) + `</span></div>
        <div class="field"><span class="field-label">ID</span><span class="field-value mono">` + m.ID + `</span></div>
        <div class="field"><span class="field-label">Created</span><span class="field-value">` + escapeHTML(m.CreatedAt) + `</span></div>` + parentHTML + `
      </div>
    </div>
  </div>

  <!-- Files -->
  <div class="card full">
    <div class="card-header"><svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg><span>Files</span><span style="margin-left:auto;font-weight:400;color:var(--text-secondary)">` + fmt.Sprintf("%d", len(entries)) + ` items</span></div>
    <div class="card-body">
      <div class="file-list" id="file-list"></div>
    </div>
  </div>

  ` + fileContentHTML + `

  ` + acceptanceHTML + `

  ` + harnessHTML + `

  ` + extensionsHTML + `

  <!-- Full Manifest (collapsible) -->
  <div class="card full">
    <div class="manifest-toggle" id="manifest-toggle" onclick="this.classList.toggle('open');document.getElementById('manifest-body').classList.toggle('open')">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="9 18 15 12 9 6"/></svg>
      <span>Full Manifest</span>
    </div>
    <div class="manifest-body" id="manifest-body">
      <pre>` + escapeHTML(string(manifestJSON)) + `</pre>
    </div>
  </div>

</div>

<script>
const files = ` + string(filesJSON) + `;
const isBundle = ` + fmt.Sprintf("%v", isBundle) + `;
const fileList = document.getElementById("file-list");
const fileIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>';
files.forEach(f => {
  const div = document.createElement("div");
  div.className = "file-item";
  div.innerHTML = fileIcon + '<span>' + f.replace(/</g,'&lt;') + '</span>';
  if (!isBundle) {
    div.onclick = () => {
      fetch("/api/file/" + encodeURIComponent(f))
        .then(r => r.text())
        .then(text => {
          const fc = document.getElementById("file-content");
          if (fc) {
            fc.style.display = "block";
            document.getElementById("file-name").textContent = f;
            document.getElementById("file-body").textContent = text;
            fc.scrollIntoView({behavior:'smooth',block:'start'});
          }
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
