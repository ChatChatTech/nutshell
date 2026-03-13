package nutshell

import (
	"encoding/json"
	"time"
)

const (
	SpecVersion = "0.2.0"
	MagicBytes  = "NUT\x01"
)

// Manifest is the nutshell.json schema.
type Manifest struct {
	NutshellVersion string       `json:"nutshell_version"`
	BundleType      string       `json:"bundle_type"` // "request" | "delivery"
	ID              string       `json:"id"`
	CreatedAt       string       `json:"created_at"`
	ExpiresAt       string       `json:"expires_at,omitempty"`
	Task            Task         `json:"task"`
	Tags            Tags         `json:"tags,omitempty"`
	Publisher       Publisher    `json:"publisher,omitempty"`
	Context         Context      `json:"context,omitempty"`
	Files           FileManifest `json:"files,omitempty"`
	APIs            *APIConfig   `json:"apis,omitempty"`
	Credentials     *Credentials `json:"credentials,omitempty"`
	Acceptance      *Acceptance  `json:"acceptance,omitempty"`
	Harness         *Harness     `json:"harness,omitempty"`
	Resources       *Resources   `json:"resources,omitempty"`
	Completeness    *Completeness `json:"completeness,omitempty"`
	ParentID        string       `json:"parent_id,omitempty"`
	Compression     *Compression `json:"compression,omitempty"`
	Extensions      map[string]json.RawMessage `json:"extensions,omitempty"`
}

type Task struct {
	Title           string `json:"title"`
	Summary         string `json:"summary,omitempty"`
	Priority        string `json:"priority,omitempty"`
	EstimatedEffort string `json:"estimated_effort,omitempty"`
}

type Tags struct {
	SkillsRequired []string               `json:"skills_required,omitempty"`
	Domains        []string               `json:"domains,omitempty"`
	DataSources    []string               `json:"data_sources,omitempty"`
	Custom         map[string]interface{} `json:"custom,omitempty"`
}

type Publisher struct {
	Name    string `json:"name,omitempty"`
	Contact string `json:"contact,omitempty"`
	Tool    string `json:"tool,omitempty"`
}

type Context struct {
	Requirements string   `json:"requirements,omitempty"`
	Architecture string   `json:"architecture,omitempty"`
	References   string   `json:"references,omitempty"`
	Additional   []string `json:"additional,omitempty"`
}

type FileManifest struct {
	TotalCount     int        `json:"total_count"`
	TotalSizeBytes int64      `json:"total_size_bytes"`
	Tree           []FileEntry `json:"tree,omitempty"`
}

type FileEntry struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Role string `json:"role,omitempty"` // "scaffold" | "reference" | "specification"
}

type APIConfig struct {
	EndpointsSpec string            `json:"endpoints_spec,omitempty"`
	BaseURLs      map[string]string `json:"base_urls,omitempty"`
	AuthMethod    string            `json:"auth_method,omitempty"`
	CredentialRef string            `json:"credential_ref,omitempty"`
}

type Credentials struct {
	Vault      string           `json:"vault,omitempty"`
	Encryption string           `json:"encryption,omitempty"` // "age" | "sops" | "vault" | "none"
	Scopes     []CredentialScope `json:"scopes,omitempty"`
}

type CredentialScope struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	AccessLevel string `json:"access_level,omitempty"`
	RateLimit   string `json:"rate_limit,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

type Acceptance struct {
	CriteriaFile        string   `json:"criteria_file,omitempty"`
	TestScripts         []string `json:"test_scripts,omitempty"`
	AutoVerifiable      bool     `json:"auto_verifiable,omitempty"`
	HumanReviewRequired bool     `json:"human_review_required,omitempty"`
	Checklist           []string `json:"checklist,omitempty"`
}

type Harness struct {
	AgentTypeHint     string   `json:"agent_type_hint,omitempty"`
	ContextBudgetHint float64  `json:"context_budget_hint,omitempty"`
	ExecutionStrategy string   `json:"execution_strategy,omitempty"`
	Checkpoints       bool     `json:"checkpoints,omitempty"`
	Constraints       []string `json:"constraints,omitempty"`
}

type Resources struct {
	Repos  []RepoRef  `json:"repos,omitempty"`
	Docs   []DocRef   `json:"docs,omitempty"`
	Images []ImageRef `json:"images,omitempty"`
	Links  []LinkRef  `json:"links,omitempty"`
}

type RepoRef struct {
	URL       string `json:"url"`
	Branch    string `json:"branch,omitempty"`
	Relevance string `json:"relevance,omitempty"`
}

type DocRef struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

type ImageRef struct {
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

type LinkRef struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

type Completeness struct {
	Status   string   `json:"status"` // "draft" | "incomplete" | "ready"
	Missing  []string `json:"missing,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// BundleHash holds the content-address of a packed bundle.
type BundleHash struct {
	Algorithm string `json:"algorithm,omitempty"` // "sha256"
	Digest    string `json:"digest,omitempty"`
}

type Compression struct {
	Algorithm            string `json:"algorithm,omitempty"`
	OriginalSizeBytes    int64  `json:"original_size_bytes,omitempty"`
	CompressedSizeBytes  int64  `json:"compressed_size_bytes,omitempty"`
	ContextTokensEstimate int   `json:"context_tokens_estimate,omitempty"`
}

// NewManifest creates a manifest with defaults.
func NewManifest() *Manifest {
	return &Manifest{
		NutshellVersion: SpecVersion,
		BundleType:      "request",
		ID:              GenerateID(),
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		Task:            Task{Priority: "medium"},
		Tags:            Tags{},
		Context: Context{
			Requirements: "context/requirements.md",
		},
		Harness: &Harness{
			AgentTypeHint:     "execution",
			ContextBudgetHint: 0.35,
			ExecutionStrategy: "incremental",
			Checkpoints:       true,
		},
		Completeness: &Completeness{
			Status: "draft",
		},
		Compression: &Compression{
			Algorithm: "gzip",
		},
	}
}
