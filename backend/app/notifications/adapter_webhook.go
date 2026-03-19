package notifications

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WebhookAdapter struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Secret  string            `json:"secret,omitempty"`
}

func (a *WebhookAdapter) Type() string { return "webhook" }

func (a *WebhookAdapter) Validate() error {
	if a.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if a.Method == "" {
		a.Method = "POST"
	}
	return nil
}

func (a *WebhookAdapter) Send(ctx context.Context, msg Message) error {
	payload := map[string]string{
		"subject":   msg.Subject,
		"body":      msg.Body,
		"severity":  string(msg.Severity),
		"ruleType":  msg.RuleType,
		"ruleName":  msg.RuleName,
		"url":       msg.URL,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	method := a.Method
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequestWithContext(ctx, method, a.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range a.Headers {
		req.Header.Set(k, v)
	}

	if a.Secret != "" {
		mac := hmac.New(sha256.New, []byte(a.Secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Traceway-Signature", "sha256="+sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
