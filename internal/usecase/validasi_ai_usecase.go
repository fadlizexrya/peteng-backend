package usecase

import (
	"context"
	"errors"
	"fmt"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/repository"
)

type ValidasiAIUsecase struct {
	repo     *repository.ValidasiAIRepository
	aiClient AIServiceClient
}

func NewValidasiAIUsecase(
	repo *repository.ValidasiAIRepository,
	aiClient AIServiceClient,
) *ValidasiAIUsecase {
	return &ValidasiAIUsecase{repo: repo, aiClient: aiClient}
}

func (u *ValidasiAIUsecase) ProsesValidasiAI(
	ctx context.Context,
	idLaporan string,
	laporanRepo *repository.LaporanRepository,
) error {
	if idLaporan == "" {
		return errors.New("id laporan wajib diisi")
	}
	if u.aiClient == nil {
		return errors.New("ai client tidak terkonfigurasi")
	}

	if laporanRepo == nil {
		return errors.New("repository laporan tidak terkonfigurasi")
	}
	laporan, err := laporanRepo.GetByID(ctx, idLaporan)
	if err != nil {
		return err
	}
	imageBytes, contentType, err := DecodeStoredPhoto(laporan.Foto)
	if err != nil {
		return err
	}

	hasil, err := u.aiClient.ValidasiFoto(ctx, imageBytes, "laporan.jpg", contentType)
	if err != nil {
		return err
	}
	normalizeAIResponse(hasil)

	return u.repo.SimpanHasilAI(ctx, idLaporan, hasil)
}

func normalizeAIResponse(response *domain.ValidasiAIResponse) {
	if response == nil {
		return
	}
	level := int16(0)
	if response.LightingClass != nil {
		switch *response.LightingClass {
		case "terang":
			level = 1
		case "redup":
			level = 2
		case "gelap":
			level = 3
		}
	}
	valid := response.PhotoStatus != nil && *response.PhotoStatus == "layak_dinilai"
	if response.AIStatus == "perlu_tinjauan_manual" {
		valid = false
	}
	result := "tidak_valid"
	if valid {
		result = "valid"
	}
	response.Data = domain.ValidasiAIResponseData{
		Success:            response.Success,
		RequestID:          response.RequestID,
		AIStatus:           response.AIStatus,
		PhotoStatus:        response.PhotoStatus,
		LightingClass:      response.LightingClass,
		Confidence:         response.Confidence,
		ModelVersion:       response.ModelVersion,
		Message:            response.Message,
		HasilValidasi:      result,
		TingkatKegelapanAI: level,
		CNNConfidence:      response.Confidence,
		Catatan:            response.Message,
	}
	if response.Data.Catatan == "" {
		response.Data.Catatan = fmt.Sprintf("Status AI: %s", response.AIStatus)
	}
}

func (u *ValidasiAIUsecase) MulaiValidasi(
	ctx context.Context,
	idLaporan string,
) error {
	if idLaporan == "" {
		return errors.New("id laporan wajib diisi")
	}

	return u.repo.MulaiValidasi(ctx, idLaporan)
}

// Simulasi ini hanya untuk pengembangan dan pengujian.
func (u *ValidasiAIUsecase) SimpanHasilMock(
	ctx context.Context,
	idLaporan string,
	req *domain.ValidasiAIMockRequest,
) error {
	if idLaporan == "" {
		return errors.New("id laporan wajib diisi")
	}

	if req.HasilValidasi != "valid" &&
		req.HasilValidasi != "tidak_valid" {
		return errors.New(
			"hasil_validasi harus valid atau tidak_valid",
		)
	}

	if req.TingkatKegelapanAI < 1 ||
		req.TingkatKegelapanAI > 3 {
		return errors.New(
			"tingkat_kegelapan_ai harus bernilai 1, 2, atau 3",
		)
	}

	if req.Catatan == "" {
		return errors.New("catatan wajib diisi")
	}

	return u.repo.SimpanHasilMock(ctx, idLaporan, req)
}

func (u *ValidasiAIUsecase) GetTerbaru(
	ctx context.Context,
	idLaporan string,
) (*domain.ValidasiAI, error) {
	if idLaporan == "" {
		return nil, errors.New("id laporan wajib diisi")
	}

	return u.repo.GetTerbaru(ctx, idLaporan)
}
