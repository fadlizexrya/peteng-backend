package usecase

import (
	"context"
	"testing"

	"peteng-backend/internal/domain"
)

func TestNormalizeAIResponseMapsValidLighting(t *testing.T) {
	for _, test := range []struct {
		lighting string
		want     int16
	}{{"terang", 1}, {"redup", 2}, {"gelap", 3}} {
		response := &domain.ValidasiAIResponse{
			Success: true, AIStatus: "selesai",
			PhotoStatus:   stringPointer("layak_dinilai"),
			LightingClass: stringPointer(test.lighting),
		}
		if err := normalizeAIResponse(response); err != nil {
			t.Fatalf("normalize %s: %v", test.lighting, err)
		}
		if response.Data.HasilValidasi != "valid" || response.Data.TingkatKegelapanAI == nil || *response.Data.TingkatKegelapanAI != test.want {
			t.Errorf("unexpected normalized result for %s: %+v", test.lighting, response.Data)
		}
	}
}

func TestNormalizeAIResponseMapsUnusablePhotoToNilLevel(t *testing.T) {
	response := &domain.ValidasiAIResponse{
		Success: true, AIStatus: "selesai",
		PhotoStatus: stringPointer("tidak_layak"),
		Message:     "Foto tidak menampilkan kondisi jalan. Harap unggah ulang foto yang sesuai.",
	}
	if err := normalizeAIResponse(response); err != nil {
		t.Fatal(err)
	}
	if response.Data.HasilValidasi != "tidak_valid" || response.Data.TingkatKegelapanAI != nil {
		t.Fatalf("expected tidak_valid with nil level, got %+v", response.Data)
	}
}

func TestRetryValidasiRequiresConfiguredDependencies(t *testing.T) {
	usecase := &ValidasiAIUsecase{}
	err := usecase.RetryValidasi(context.Background(), "report-1", nil)
	if err == nil || err.Error() != "layanan validasi AI tidak terkonfigurasi" {
		t.Fatalf("expected configuration error, got %v", err)
	}
}

func TestDecodeStoredPhotoDataURI(t *testing.T) {
	got, contentType, err := DecodeStoredPhoto("data:image/png;base64,aGVsbG8=")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" || contentType != "image/png" {
		t.Fatalf("unexpected decoded photo: bytes=%q contentType=%q", got, contentType)
	}
}

func stringPointer(value string) *string { return &value }
