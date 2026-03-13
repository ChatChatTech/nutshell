package nutshell

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Pack creates a .nut bundle from a directory.
func Pack(srcDir, output string) (*Manifest, error) {
	manifestPath := filepath.Join(srcDir, "nutshell.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("no nutshell.json found in %s: %w", srcDir, err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid nutshell.json: %w", err)
	}

	// Collect files
	type fileInfo struct {
		path string
		rel  string
		size int64
	}
	// Load ignore patterns
	ignorePatterns := LoadIgnorePatterns(srcDir)

	var files []fileInfo
	var totalSize int64

	err = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(d.Name(), ".") && path != srcDir {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(srcDir, path)
		// Skip .nutignore itself and ignored files
		if rel == ".nutignore" {
			return nil
		}
		if IsIgnored(rel, ignorePatterns) {
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
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	// Update manifest
	manifest.Files.TotalCount = len(files)
	manifest.Files.TotalSizeBytes = totalSize
	if manifest.Compression == nil {
		manifest.Compression = &Compression{}
	}
	manifest.Compression.OriginalSizeBytes = totalSize

	// Write bundle
	outFile, err := os.Create(output)
	if err != nil {
		return nil, fmt.Errorf("creating output file: %w", err)
	}
	defer outFile.Close()

	// Write magic bytes
	if _, err := outFile.Write([]byte(MagicBytes)); err != nil {
		return nil, err
	}

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Write manifest first (with updated counts)
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{
		Name: "nutshell.json",
		Size: int64(len(manifestBytes)),
		Mode: 0644,
	}); err != nil {
		return nil, err
	}
	if _, err := tw.Write(manifestBytes); err != nil {
		return nil, err
	}

	// Write all other files
	for _, f := range files {
		if f.rel == "nutshell.json" {
			continue
		}
		if err := addFileToTar(tw, f.path, f.rel); err != nil {
			return nil, fmt.Errorf("adding %s: %w", f.rel, err)
		}
	}

	return &manifest, nil
}

func addFileToTar(tw *tar.Writer, absPath, relPath string) error {
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name: relPath,
		Size: info.Size(),
		Mode: 0644,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err = io.Copy(tw, file)
	return err
}

// Unpack extracts a .nut bundle to a directory.
func Unpack(nutPath, outDir string) (*Manifest, error) {
	f, err := os.Open(nutPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := checkMagic(f); err != nil {
		return nil, err
	}

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("invalid gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, err
	}

	var manifest *Manifest

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Security: prevent path traversal
		clean := filepath.Clean(header.Name)
		if filepath.IsAbs(clean) || strings.Contains(clean, "..") {
			return nil, fmt.Errorf("unsafe path in bundle: %s", header.Name)
		}

		target := filepath.Join(outDir, clean)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return nil, err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return nil, err
			}
			outFile.Close()

			// Parse manifest if found
			if clean == "nutshell.json" {
				data, err := os.ReadFile(target)
				if err == nil {
					var m Manifest
					if json.Unmarshal(data, &m) == nil {
						manifest = &m
					}
				}
			}
		}
	}

	return manifest, nil
}

// Inspect reads the manifest from a .nut bundle without extracting.
func Inspect(nutPath string) (*Manifest, []string, error) {
	f, err := os.Open(nutPath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return InspectReader(f)
}

// InspectReader reads the manifest from a .nut stream (file or stdin).
func InspectReader(r io.Reader) (*Manifest, []string, error) {
	// Read and validate magic bytes
	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, nil, fmt.Errorf("reading magic bytes: %w", err)
	}
	if string(magic) != MagicBytes {
		return nil, nil, fmt.Errorf("not a valid nutshell bundle (bad magic bytes)")
	}

	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	var manifest *Manifest
	var entries []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		entries = append(entries, header.Name)

		if header.Name == "nutshell.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, nil, err
			}
			var m Manifest
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, nil, fmt.Errorf("invalid manifest: %w", err)
			}
			manifest = &m
		}
	}

	if manifest == nil {
		return nil, nil, fmt.Errorf("no nutshell.json in bundle")
	}
	return manifest, entries, nil
}

func checkMagic(f *os.File) error {
	magic := make([]byte, 4)
	if _, err := io.ReadFull(f, magic); err != nil {
		return fmt.Errorf("reading magic bytes: %w", err)
	}
	if string(magic) != MagicBytes {
		return fmt.Errorf("not a valid nutshell bundle (bad magic bytes)")
	}
	return nil
}
