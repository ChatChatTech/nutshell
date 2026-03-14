package nutshell

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CompressionLevel controls how aggressively to compress.
type CompressionLevel int

const (
	CompressNone    CompressionLevel = 0
	CompressFast    CompressionLevel = 1
	CompressDefault CompressionLevel = 6
	CompressBest    CompressionLevel = 9
)

// FileStrategy describes the compression decision for a single file.
type FileStrategy struct {
	Path           string `json:"path"`
	OriginalSize   int64  `json:"original_size"`
	Category       string `json:"category"`       // "text", "binary", "precompressed", "media"
	GzipLevel      int    `json:"gzip_level"`      // effective gzip level for this file
	Recommendation string `json:"recommendation"`  // "compress", "store", "skip"
}

// CompressionPlan is the analysis result for a directory.
type CompressionPlan struct {
	Files           []FileStrategy `json:"files"`
	TotalOriginal   int64          `json:"total_original"`
	TextBytes       int64          `json:"text_bytes"`
	PrecompBytes    int64          `json:"precompressed_bytes"`
	EstimatedTokens int            `json:"estimated_tokens"` // rough token count for text files
}

// precompressed extensions — these are already compressed, gzip adds no benefit.
var precompressedExts = map[string]bool{
	".gz": true, ".zip": true, ".zst": true, ".xz": true, ".bz2": true,
	".7z": true, ".rar": true, ".lz4": true, ".br": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".mp3": true, ".mp4": true, ".m4a": true, ".ogg": true, ".flac": true,
	".avi": true, ".mov": true, ".mkv": true, ".webm": true,
	".woff2": true, ".woff": true, ".ttf": true,
	".nut": true, // don't re-compress nutshell bundles
	".pdf": true,
}

// text extensions — these compress very well and contain context tokens.
var textExts = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
	".md": true, ".txt": true, ".json": true, ".yaml": true, ".yml": true, ".toml": true,
	".xml": true, ".html": true, ".css": true, ".scss": true, ".less": true,
	".rs": true, ".c": true, ".cpp": true, ".h": true, ".hpp": true,
	".java": true, ".kt": true, ".scala": true, ".swift": true,
	".rb": true, ".php": true, ".lua": true, ".sh": true, ".bash": true,
	".sql": true, ".graphql": true, ".proto": true, ".csv": true,
	".env": true, ".ini": true, ".cfg": true, ".conf": true,
	".r": true, ".jl": true, ".ex": true, ".exs": true, ".erl": true,
	".zig": true, ".nim": true, ".v": true, ".dart": true, ".vue": true,
	".svelte": true, ".astro": true,
	".log": true, ".diff": true, ".patch": true,
}

// tokensPerByte is a rough estimate: ~4 chars per token for English/code.
const tokensPerByte = 0.25

// AnalyzeCompression builds a context-aware compression plan for a directory.
func AnalyzeCompression(dir string) (*CompressionPlan, error) {
	plan := &CompressionPlan{}
	ignorePatterns := LoadIgnorePatterns(dir)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != dir {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		if rel == ".nutignore" || IsIgnored(rel, ignorePatterns) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}

		fs := classifyFile(rel, info.Size())
		plan.Files = append(plan.Files, fs)
		plan.TotalOriginal += info.Size()

		switch fs.Category {
		case "text":
			plan.TextBytes += info.Size()
			plan.EstimatedTokens += int(float64(info.Size()) * tokensPerByte)
		case "precompressed", "media":
			plan.PrecompBytes += info.Size()
		}
		return nil
	})
	return plan, err
}

func classifyFile(relPath string, size int64) FileStrategy {
	ext := strings.ToLower(filepath.Ext(relPath))
	name := strings.ToLower(filepath.Base(relPath))

	fs := FileStrategy{
		Path:         relPath,
		OriginalSize: size,
	}

	// Detect by extension
	if precompressedExts[ext] {
		fs.Category = "precompressed"
		fs.GzipLevel = int(CompressFast) // store-like; gzip at level 1 for tar compat
		fs.Recommendation = "store"
		return fs
	}

	if textExts[ext] || name == "nutshell.json" || name == "makefile" || name == "dockerfile" || name == "license" {
		fs.Category = "text"
		fs.GzipLevel = int(CompressBest) // text compresses excellently
		fs.Recommendation = "compress"
		return fs
	}

	// Default: generic binary
	fs.Category = "binary"
	fs.GzipLevel = int(CompressDefault)
	fs.Recommendation = "compress"
	return fs
}

// PackWithCompression creates a .nut bundle with context-aware compression strategy.
// level controls the overall compression aggressiveness.
func PackWithCompression(srcDir, output string, level CompressionLevel) (*Manifest, *CompressionPlan, error) {
	manifestPath := filepath.Join(srcDir, "nutshell.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, nil, fmt.Errorf("no nutshell.json found in %s: %w", srcDir, err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, nil, fmt.Errorf("invalid nutshell.json: %w", err)
	}

	// Analyze files
	plan, err := AnalyzeCompression(srcDir)
	if err != nil {
		return nil, nil, fmt.Errorf("analyzing compression: %w", err)
	}

	// Collect files
	ignorePatterns := LoadIgnorePatterns(srcDir)
	type fileInfo struct {
		path string
		rel  string
		size int64
	}
	var files []fileInfo
	var totalSize int64
	err = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != srcDir {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(srcDir, path)
		if rel == ".nutignore" || IsIgnored(rel, ignorePatterns) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		files = append(files, fileInfo{path: path, rel: rel, size: info.Size()})
		totalSize += info.Size()
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walking directory: %w", err)
	}

	// Update manifest
	manifest.Files.TotalCount = len(files)
	manifest.Files.TotalSizeBytes = totalSize
	manifest.Files.Tree = nil
	for _, f := range files {
		hash, err := hashFile(f.path)
		if err != nil {
			return nil, nil, fmt.Errorf("hashing %s: %w", f.rel, err)
		}
		manifest.Files.Tree = append(manifest.Files.Tree, FileEntry{
			Path: f.rel, Size: f.size, Hash: hash,
		})
	}
	if manifest.Compression == nil {
		manifest.Compression = &Compression{}
	}
	manifest.Compression.Algorithm = "gzip"
	manifest.Compression.OriginalSizeBytes = totalSize
	manifest.Compression.ContextTokensEstimate = plan.EstimatedTokens

	// Write bundle with selected gzip level
	outFile, err := os.Create(output)
	if err != nil {
		return nil, nil, fmt.Errorf("creating output file: %w", err)
	}
	defer outFile.Close()

	if _, err := outFile.Write([]byte(MagicBytes)); err != nil {
		return nil, nil, err
	}

	gzLevel := int(level)
	if gzLevel == 0 {
		gzLevel = gzip.NoCompression
	}
	gw, err := gzip.NewWriterLevel(outFile, gzLevel)
	if err != nil {
		return nil, nil, err
	}
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	if err := tw.WriteHeader(&tar.Header{
		Name: "nutshell.json",
		Size: int64(len(manifestBytes)),
		Mode: 0644,
	}); err != nil {
		return nil, nil, err
	}
	if _, err := tw.Write(manifestBytes); err != nil {
		return nil, nil, err
	}

	for _, f := range files {
		if f.rel == "nutshell.json" {
			continue
		}
		if err := addFileToTar(tw, f.path, f.rel); err != nil {
			return nil, nil, fmt.Errorf("adding %s: %w", f.rel, err)
		}
	}

	return &manifest, plan, nil
}
