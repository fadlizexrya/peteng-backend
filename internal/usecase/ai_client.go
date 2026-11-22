package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"
	"time"

	"peteng-backend/internal/domain"
)

type AIServiceClient interface {
	ValidasiFoto(ctx context.Context, imageBytes []byte, filename, contentType string) (*domain.ValidasiAIResponse, error)
}

type HTTPAIServiceClient struct {
	baseURL      string
	serviceToken string
	client       *http.Client
}

func NewHTTPAIServiceClient(baseURL string) *HTTPAIServiceClient {
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	return &HTTPAIServiceClient{
		baseURL:      strings.TrimRight(baseURL, "/"),
		serviceToken: os.Getenv("AI_SERVICE_TOKEN"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *HTTPAIServiceClient) ValidasiFoto(
	ctx context.Context,
	imageBytes []byte,
	filename string,
	contentType string,
) (*domain.ValidasiAIResponse, error) {
	if len(imageBytes) == 0 {
		return nil, errors.New("foto laporan kosong")
	}

	const maxAttempts = 3
	requestID := fmt.Sprintf("ai-%d", time.Now().UnixNano())
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		if filename == "" {
			filename = "laporan.jpg"
		}
		if contentType == "" {
			contentType = "image/jpeg"
		}
		part, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Disposition": []string{fmt.Sprintf(`form-data; name="image"; filename="%s"`, filename)},
			"Content-Type":        []string{contentType},
		})
		if err != nil {
			return nil, err
		}
		if _, err = part.Write(imageBytes); err != nil {
			return nil, err
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}

		httpReq, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.baseURL+"/api/v1/ai/validate",
			&body,
		)
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Content-Type", writer.FormDataContentType())
		httpReq.Header.Set("X-Request-ID", requestID)
		if c.serviceToken != "" {
			httpReq.Header.Set("Authorization", "Bearer "+c.serviceToken)
		}
		_ = contentType // MIME is detected and validated by the AI service.

		resp, err := c.client.Do(httpReq)
		if err != nil {
			lastErr = err
		} else {
			responseBody, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				return nil, readErr
			}

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				var result domain.ValidasiAIResponse
				if err := json.Unmarshal(responseBody, &result); err != nil {
					return nil, err
				}
				result.Data = domain.ValidasiAIResponseData{
					Success:       result.Success,
					RequestID:     result.RequestID,
					AIStatus:      result.AIStatus,
					PhotoStatus:   result.PhotoStatus,
					LightingClass: result.LightingClass,
					Confidence:    result.Confidence,
					ModelVersion:  result.ModelVersion,
					Message:       result.Message,
				}
				return &result, nil
			}

			lastErr = fmt.Errorf("ai service status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
			if resp.StatusCode < 500 && resp.StatusCode != http.StatusRequestTimeout {
				return nil, lastErr
			}
		}

		if attempt < maxAttempts-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(1<<attempt) * time.Second):
			}
		}
	}
	return nil, lastErr
}

// DecodeStoredPhoto converts the existing base64 JSON field into image bytes.
// It supports both plain base64 and data URI values used by older clients.
func DecodeStoredPhoto(value string) ([]byte, string, error) {
	value = strings.TrimSpace(value)
	contentType := "image/jpeg"
	if comma := strings.Index(value, ","); strings.HasPrefix(value, "data:") && comma >= 0 {
		header := value[:comma]
		value = value[comma+1:]
		if semi := strings.Index(header, ";"); semi > len("data:") {
			contentType = header[len("data:"):semi]
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, "", fmt.Errorf("foto laporan bukan base64 valid: %w", err)
	}
	if len(decoded) == 0 {
		return nil, "", errors.New("foto laporan kosong")
	}
	return decoded, contentType, nil
}
