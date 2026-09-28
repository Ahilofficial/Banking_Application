package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AuditLogPayload struct {
	UserID       uint   `json:"user_id"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	OldValue     string `json:"old_value,omitempty"`
	NewValue     string `json:"new_value,omitempty"`
	IPAddress    string `json:"ip_address,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
}

type AuditClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuditClient(baseURL string) *AuditClient {
	return &AuditClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Log sends an audit log entry asynchronously to the audit microservice
func (c *AuditClient) Log(payload AuditLogPayload) {
	go func() {
		body, err := json.Marshal(payload)
		if err != nil {
			return
		}

		req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/audit/logs", bytes.NewBuffer(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if payload.RequestID != "" {
			req.Header.Set("X-Request-ID", payload.RequestID)
		}

		resp, err := c.httpClient.Do(req)
		if err == nil && resp != nil {
			_ = resp.Body.Close()
		}
	}()
}

// LogSync sends an audit log entry synchronously
func (c *AuditClient) LogSync(payload AuditLogPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/audit/logs", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if payload.RequestID != "" {
		req.Header.Set("X-Request-ID", payload.RequestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("audit service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("audit service returned error status: %d", resp.StatusCode)
	}
	return nil
}
