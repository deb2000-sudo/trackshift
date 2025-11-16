package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/deb2000-sudo/trackshift/pkg/models"
)

// OrchestratorClient is a small HTTP client for the orchestrator service.
type OrchestratorClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewOrchestratorClient creates a new client with reasonable defaults.
func NewOrchestratorClient(baseURL string) *OrchestratorClient {
	return &OrchestratorClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateSession creates a new transfer session.
func (c *OrchestratorClient) CreateSession(file models.FileMetadata) (*models.TransferSession, error) {
	body, err := json.Marshal(map[string]any{
		"file": file,
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/v1/session", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var sess models.TransferSession
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

// GetSession fetches a session by ID.
func (c *OrchestratorClient) GetSession(id string) (*models.TransferSession, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/v1/session/" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var sess models.TransferSession
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
		return nil, err
	}
	return &sess, nil
}


