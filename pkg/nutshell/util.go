package nutshell

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GenerateID creates a new nutshell bundle ID.
func GenerateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	// Format as nut-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	return fmt.Sprintf("nut-%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// HashBundle computes SHA-256 of the packed .nut file contents (after magic bytes).
func HashBundle(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

// LoadIgnorePatterns reads .nutignore from dir and returns patterns.
func LoadIgnorePatterns(dir string) []string {
	path := filepath.Join(dir, ".nutignore")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

// IsIgnored checks if a relative path matches any ignore pattern.
func IsIgnored(relPath string, patterns []string) bool {
	for _, p := range patterns {
		if matched, _ := filepath.Match(p, relPath); matched {
			return true
		}
		if matched, _ := filepath.Match(p, filepath.Base(relPath)); matched {
			return true
		}
		// Support directory prefix patterns like "delivery/"
		clean := strings.TrimSuffix(p, "/")
		if strings.HasPrefix(relPath, clean+"/") || relPath == clean {
			return true
		}
	}
	return false
}
