package nutshell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupCredentialDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "Credential Test"
	m.Credentials = &Credentials{
		Vault:      "credentials/vault.age",
		Encryption: "age",
		Scopes: []CredentialScope{
			{
				Name:      "staging-db",
				Type:      "postgres",
				ExpiresAt: time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339), // expired
			},
			{
				Name:      "prod-api",
				Type:      "api_key",
				ExpiresAt: time.Now().Add(5 * 24 * time.Hour).UTC().Format(time.RFC3339), // expiring soon
			},
			{
				Name:      "long-lived",
				Type:      "token",
				ExpiresAt: time.Now().Add(365 * 24 * time.Hour).UTC().Format(time.RFC3339), // valid
			},
			{
				Name: "no-expiry",
				Type: "ssh_key",
				// no ExpiresAt
			},
		},
	}
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)
	return dir
}

func TestAuditCredentials(t *testing.T) {
	dir := setupCredentialDir(t)

	statuses, err := AuditCredentials(dir)
	if err != nil {
		t.Fatalf("AuditCredentials failed: %v", err)
	}
	if len(statuses) != 4 {
		t.Fatalf("expected 4 credential statuses, got %d", len(statuses))
	}

	statusMap := make(map[string]CredentialStatus)
	for _, s := range statuses {
		statusMap[s.Name] = s
	}

	if statusMap["staging-db"].Status != "expired" {
		t.Fatalf("staging-db should be expired, got %q", statusMap["staging-db"].Status)
	}
	if statusMap["prod-api"].Status != "expiring_soon" {
		t.Fatalf("prod-api should be expiring_soon, got %q", statusMap["prod-api"].Status)
	}
	if statusMap["long-lived"].Status != "valid" {
		t.Fatalf("long-lived should be valid, got %q", statusMap["long-lived"].Status)
	}
	if statusMap["no-expiry"].Status != "no_expiry" {
		t.Fatalf("no-expiry should be no_expiry, got %q", statusMap["no-expiry"].Status)
	}
}

func TestAuditCredentialsNoCredentials(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	m.Task.Title = "No Creds"
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)

	statuses, err := AuditCredentials(dir)
	if err != nil {
		t.Fatalf("AuditCredentials failed: %v", err)
	}
	if len(statuses) != 0 {
		t.Fatalf("expected 0 statuses, got %d", len(statuses))
	}
}

func TestRotateCredential(t *testing.T) {
	dir := setupCredentialDir(t)

	newExpiry := time.Now().Add(90 * 24 * time.Hour).UTC().Format(time.RFC3339)
	result, err := RotateCredential(dir, "staging-db", newExpiry)
	if err != nil {
		t.Fatalf("RotateCredential failed: %v", err)
	}
	if result.Scope != "staging-db" {
		t.Fatalf("expected scope 'staging-db', got %q", result.Scope)
	}
	if result.NewExpiry == "" {
		t.Fatal("expected new expiry")
	}

	// Re-audit to verify it's no longer expired
	statuses, err := AuditCredentials(dir)
	if err != nil {
		t.Fatalf("re-audit failed: %v", err)
	}
	for _, s := range statuses {
		if s.Name == "staging-db" && s.Status == "expired" {
			t.Fatal("staging-db should no longer be expired after rotation")
		}
	}
}

func TestRotateCredentialDefault(t *testing.T) {
	dir := setupCredentialDir(t)

	// Empty expiry should default to +30 days
	result, err := RotateCredential(dir, "staging-db", "")
	if err != nil {
		t.Fatalf("RotateCredential(default) failed: %v", err)
	}
	if result.NewExpiry == "" {
		t.Fatal("expected new expiry from default")
	}
}

func TestRotateCredentialNotFound(t *testing.T) {
	dir := setupCredentialDir(t)

	_, err := RotateCredential(dir, "nonexistent", "2026-01-01T00:00:00Z")
	if err == nil {
		t.Fatal("expected error for nonexistent scope")
	}
}
