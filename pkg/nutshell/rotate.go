package nutshell

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CredentialStatus describes the state of a single credential scope.
type CredentialStatus struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	ExpiresAt string `json:"expires_at,omitempty"`
	Status    string `json:"status"` // "valid", "expiring_soon", "expired", "no_expiry"
	DaysLeft  int    `json:"days_left,omitempty"`
}

// RotateResult holds the outcome of a credential rotation.
type RotateResult struct {
	Scope      string `json:"scope"`
	OldExpiry  string `json:"old_expires_at"`
	NewExpiry  string `json:"new_expires_at"`
}

// AuditCredentials checks expiration status of all credential scopes in a bundle.
func AuditCredentials(dir string) ([]CredentialStatus, error) {
	data, err := os.ReadFile(filepath.Join(dir, "nutshell.json"))
	if err != nil {
		return nil, fmt.Errorf("no nutshell.json in %s: %w", dir, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid nutshell.json: %w", err)
	}

	if m.Credentials == nil || len(m.Credentials.Scopes) == 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	var results []CredentialStatus

	for _, scope := range m.Credentials.Scopes {
		cs := CredentialStatus{
			Name: scope.Name,
			Type: scope.Type,
		}

		if scope.ExpiresAt == "" {
			cs.Status = "no_expiry"
			cs.ExpiresAt = ""
		} else {
			cs.ExpiresAt = scope.ExpiresAt
			expiry, err := time.Parse(time.RFC3339, scope.ExpiresAt)
			if err != nil {
				cs.Status = "valid" // can't parse, assume valid
			} else {
				remaining := expiry.Sub(now)
				cs.DaysLeft = int(remaining.Hours() / 24)
				if remaining <= 0 {
					cs.Status = "expired"
				} else if remaining <= 7*24*time.Hour {
					cs.Status = "expiring_soon"
				} else {
					cs.Status = "valid"
				}
			}
		}
		results = append(results, cs)
	}

	return results, nil
}

// RotateCredential updates the expiration of a named credential scope.
// If newExpiry is empty, it extends by 30 days from now.
func RotateCredential(dir, scopeName, newExpiry string) (*RotateResult, error) {
	mPath := filepath.Join(dir, "nutshell.json")
	data, err := os.ReadFile(mPath)
	if err != nil {
		return nil, fmt.Errorf("no nutshell.json in %s: %w", dir, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid nutshell.json: %w", err)
	}

	if m.Credentials == nil || len(m.Credentials.Scopes) == 0 {
		return nil, fmt.Errorf("no credentials defined in manifest")
	}

	// Find the scope
	idx := -1
	for i, s := range m.Credentials.Scopes {
		if s.Name == scopeName {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("credential scope %q not found", scopeName)
	}

	oldExpiry := m.Credentials.Scopes[idx].ExpiresAt
	if newExpiry == "" {
		newExpiry = time.Now().UTC().Add(30 * 24 * time.Hour).Format(time.RFC3339)
	}

	m.Credentials.Scopes[idx].ExpiresAt = newExpiry

	// Write back
	out, _ := json.MarshalIndent(&m, "", "  ")
	if err := os.WriteFile(mPath, out, 0644); err != nil {
		return nil, err
	}

	return &RotateResult{
		Scope:     scopeName,
		OldExpiry: oldExpiry,
		NewExpiry: newExpiry,
	}, nil
}
