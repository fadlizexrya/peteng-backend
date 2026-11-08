//go:build integration

package usecase

import (
	"context"
	"os"
	"testing"
)

func TestManualAIIntegration(t *testing.T) {
	image, err := os.ReadFile(`D:\peteng-project\ai_peteng\dataset\test\terang\pub_terang_008.jpg`)
	if err != nil {
		t.Fatal(err)
	}

	client := NewHTTPAIServiceClient("http://127.0.0.1:8000")
	result, err := client.ValidasiFoto(context.Background(), image, "laporan.jpg", "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || result.AIStatus == "" {
		t.Fatalf("response AI tidak valid: %#v", result)
	}
	t.Logf("AI response: success=%v request_id=%s ai_status=%s photo_status=%v lighting_class=%v confidence=%v model_version=%v message=%q", result.Success, result.RequestID, result.AIStatus, result.PhotoStatus, result.LightingClass, result.Confidence, result.ModelVersion, result.Message)
}
