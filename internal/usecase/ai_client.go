package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"peteng-backend/internal/domain"
)

type AIServiceClient interface {
	ValidasiLaporan(ctx context.Context, req *domain.ValidasiAIRequest) (*domain.ValidasiAIResponse, error)
}

type HTTPAIServiceClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPAIServiceClient(baseURL string) *HTTPAIServiceClient {
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	return &HTTPAIServiceClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *HTTPAIServiceClient) ValidasiLaporan(ctx context.Context, req *domain.ValidasiAIRequest) (*domain.ValidasiAIResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/validasi", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(httpReq)
		if err != nil {
			lastErr = err
			time.Sleep(2 * time.Second)
			continue
		}

		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("ai service status %d", resp.StatusCode)
			time.Sleep(2 * time.Second)
			continue
		}

		var result domain.ValidasiAIResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, lastErr
}