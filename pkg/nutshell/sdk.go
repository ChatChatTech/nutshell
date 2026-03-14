package nutshell

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Bundle provides programmatic access to a .nut bundle's contents.
type Bundle struct {
	path     string
	manifest *Manifest
	entries  map[string][]byte // rel path → content
}

// Open reads a .nut bundle and returns a Bundle for programmatic access.
func Open(nutPath string) (*Bundle, error) {
	f, err := os.Open(nutPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	magic := make([]byte, 4)
	if _, err := io.ReadFull(f, magic); err != nil {
		return nil, fmt.Errorf("reading magic bytes: %w", err)
	}
	if string(magic) != MagicBytes {
		return nil, fmt.Errorf("not a valid nutshell bundle (bad magic bytes)")
	}

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("invalid gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	b := &Bundle{
		path:    nutPath,
		entries: make(map[string][]byte),
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", header.Name, err)
		}
		b.entries[header.Name] = data

		if header.Name == "nutshell.json" {
			var m Manifest
			if err := jsonUnmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("invalid manifest: %w", err)
			}
			b.manifest = &m
		}
	}

	if b.manifest == nil {
		return nil, fmt.Errorf("no nutshell.json in bundle")
	}
	return b, nil
}

// Manifest returns the parsed bundle manifest.
func (b *Bundle) Manifest() *Manifest {
	return b.manifest
}

// ListFiles returns all file paths in the bundle.
func (b *Bundle) ListFiles() []string {
	out := make([]string, 0, len(b.entries))
	for name := range b.entries {
		out = append(out, name)
	}
	return out
}

// ReadFile returns the raw contents of a file inside the bundle.
func (b *Bundle) ReadFile(path string) ([]byte, error) {
	data, ok := b.entries[path]
	if !ok {
		return nil, fmt.Errorf("file not found in bundle: %s", path)
	}
	return data, nil
}

// ReadContext returns the content of the requirements file referenced in the manifest.
func (b *Bundle) ReadContext() ([]byte, error) {
	req := b.manifest.Context.Requirements
	if req == "" {
		return nil, fmt.Errorf("no context.requirements set in manifest")
	}
	return b.ReadFile(req)
}

// ReadFileString is a convenience wrapper that returns file content as string.
func (b *Bundle) ReadFileString(path string) (string, error) {
	data, err := b.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// HasFile checks whether a file exists in the bundle.
func (b *Bundle) HasFile(path string) bool {
	_, ok := b.entries[path]
	return ok
}

// FilesByPrefix returns all files under a given prefix (e.g. "context/").
func (b *Bundle) FilesByPrefix(prefix string) []string {
	var out []string
	for name := range b.entries {
		if strings.HasPrefix(name, prefix) {
			out = append(out, name)
		}
	}
	return out
}

// ManifestJSON returns the raw manifest JSON bytes.
func (b *Bundle) ManifestJSON() ([]byte, error) {
	return b.ReadFile("nutshell.json")
}

// Repack writes the bundle contents to a writer in .nut format.
func (b *Bundle) Repack(w io.Writer) error {
	if _, err := w.Write([]byte(MagicBytes)); err != nil {
		return err
	}
	gw := gzip.NewWriter(w)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Write manifest first
	if data, ok := b.entries["nutshell.json"]; ok {
		if err := tw.WriteHeader(&tar.Header{
			Name: "nutshell.json",
			Size: int64(len(data)),
			Mode: 0644,
		}); err != nil {
			return err
		}
		if _, err := tw.Write(data); err != nil {
			return err
		}
	}

	for name, data := range b.entries {
		if name == "nutshell.json" {
			continue
		}
		if err := tw.WriteHeader(&tar.Header{
			Name: name,
			Size: int64(len(data)),
			Mode: 0644,
		}); err != nil {
			return err
		}
		if _, err := tw.Write(data); err != nil {
			return err
		}
	}
	return nil
}

// jsonUnmarshal is a simple wrapper to avoid importing encoding/json again.
func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(bytes.TrimSpace(data), v)
}
