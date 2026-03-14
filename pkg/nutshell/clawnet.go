package nutshell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultClawNetAddr = "http://localhost:3998"
	clawnetTimeout     = 15 * time.Second
)

// ClawNetClient talks to a local ClawNet daemon HTTP API.
// Nutshell has zero compile-time dependency on ClawNet —
// all interaction is through the REST API.
type ClawNetClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClawNetClient creates a client for the local ClawNet daemon.
func NewClawNetClient(addr string) *ClawNetClient {
	if addr == "" {
		addr = DefaultClawNetAddr
	}
	return &ClawNetClient{
		BaseURL:    strings.TrimRight(addr, "/"),
		HTTPClient: &http.Client{Timeout: clawnetTimeout},
	}
}

// ClawNetStatus represents the ClawNet daemon status.
type ClawNetStatus struct {
	PeerID    string `json:"peer_id"`
	PeerCount int    `json:"peer_count"`
	AgentName string `json:"agent_name"`
}

// ClawNetTask represents a task in the ClawNet Task Bazaar.
type ClawNetTask struct {
	ID          string  `json:"id"`
	AuthorID    string  `json:"author_id"`
	AuthorName  string  `json:"author_name"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Tags        string  `json:"tags"`
	Deadline    string  `json:"deadline"`
	Reward      float64 `json:"reward"`
	Status      string  `json:"status"`
	AssignedTo  string  `json:"assigned_to"`
	Result      string  `json:"result"`
	// Nutshell extension fields (Phase 5)
	NutshellHash string `json:"nutshell_hash,omitempty"`
	NutshellID   string `json:"nutshell_id,omitempty"`
	BundleType   string `json:"bundle_type,omitempty"`
}

// Ping checks if the ClawNet daemon is reachable.
func (c *ClawNetClient) Ping() (*ClawNetStatus, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/status")
	if err != nil {
		return nil, fmt.Errorf("clawnet daemon not reachable at %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()
	var status ClawNetStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("invalid status response: %w", err)
	}
	return &status, nil
}

// PublishTask creates a task in ClawNet's Task Bazaar from a nutshell manifest.
func (c *ClawNetClient) PublishTask(m *Manifest, nutHash string) (*ClawNetTask, error) {
	// Map nutshell manifest → ClawNet task fields
	tags, _ := json.Marshal(m.Tags.SkillsRequired)

	body := map[string]interface{}{
		"title":       m.Task.Title,
		"description": m.Task.Summary,
		"tags":        m.Tags.SkillsRequired,
	}
	if m.ExpiresAt != "" {
		body["deadline"] = m.ExpiresAt
	}

	// Nutshell extension fields
	if nutHash != "" {
		body["nutshell_hash"] = nutHash
	}
	body["nutshell_id"] = m.ID
	body["bundle_type"] = m.BundleType

	_ = tags // used in the map above as array

	data, _ := json.Marshal(body)
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/tasks", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating task: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("task creation failed (%d): %s", resp.StatusCode, string(b))
	}
	var task ClawNetTask
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("decoding task response: %w", err)
	}
	return &task, nil
}

// UploadBundle uploads a .nut file to ClawNet as a task bundle attachment.
func (c *ClawNetClient) UploadBundle(taskID, nutPath string) error {
	f, err := os.Open(nutPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("bundle", filepath.Base(nutPath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, f); err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.BaseURL+"/api/tasks/"+taskID+"/bundle", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("uploading bundle: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

// DownloadBundle downloads the .nut bundle attached to a ClawNet task.
func (c *ClawNetClient) DownloadBundle(taskID, outPath string) error {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/tasks/" + taskID + "/bundle")
	if err != nil {
		return fmt.Errorf("downloading bundle: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return fmt.Errorf("no bundle attached to task %s", taskID)
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed (%d): %s", resp.StatusCode, string(b))
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// SubmitDelivery submits a delivery result for a task, optionally with a .nut bundle hash.
func (c *ClawNetClient) SubmitDelivery(taskID, result, deliveryHash string) error {
	body := map[string]string{"result": result}
	if deliveryHash != "" {
		body["nutshell_hash"] = deliveryHash
	}
	data, _ := json.Marshal(body)
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/tasks/"+taskID+"/submit", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("submitting delivery: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("submit failed (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

// GetTask fetches task details from ClawNet.
func (c *ClawNetClient) GetTask(taskID string) (*ClawNetTask, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/tasks/" + taskID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("task not found")
	}
	var task ClawNetTask
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}
	return &task, nil
}

// ListOpenTasks lists open tasks from ClawNet.
func (c *ClawNetClient) ListOpenTasks() ([]*ClawNetTask, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/tasks?status=open")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tasks []*ClawNetTask
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// ManifestToClawNetExtension returns the extensions.clawnet JSON for embedding in a nutshell manifest.
func ManifestToClawNetExtension(peerID, taskID string, reward float64) json.RawMessage {
	ext := map[string]interface{}{
		"peer_id": peerID,
		"task_id": taskID,
		"reward":  map[string]interface{}{"amount": reward, "currency": "energy"},
	}
	data, _ := json.Marshal(ext)
	return json.RawMessage(data)
}
