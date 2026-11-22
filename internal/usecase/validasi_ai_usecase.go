package usecase

import (
	"context"
	"encoding/json"
	"errors"

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
	if u.repo == nil {
		return errors.New("repository validasi AI tidak terkonfigurasi")
	}
	if u.aiClient == nil {
		return u.gagalValidasi(ctx, idLaporan, errors.New("ai client tidak terkonfigurasi"))
	}
	if laporanRepo == nil {
		return u.gagalValidasi(ctx, idLaporan, errors.New("repository laporan tidak terkonfigurasi"))
	}
	laporan, err := laporanRepo.GetByID(ctx, idLaporan)
	if err != nil {
		return err
	}
	if err := u.repo.MulaiValidasi(ctx, idLaporan); err != nil {
		return err
	}
	proses := func() error {
		imageBytes, contentType, err := DecodeStoredPhoto(laporan.Foto)
		if err != nil {
			return err
		}
		hasil, err := u.aiClient.ValidasiFoto(ctx, imageBytes, "laporan.jpg", contentType)
		if err != nil {
			return err
		}
		if hasil == nil {
			return errors.New("layanan AI mengembalikan respons kosong")
		}
		responseJSON, err := json.Marshal(hasil)
		if err != nil {
			return err
		}
		if err := normalizeAIResponse(hasil); err != nil {
			return err
		}
		hasil.Data.HasilJSON = responseJSON
		return u.repo.SimpanHasilAI(ctx, idLaporan, hasil)
	}
	if err := proses(); err != nil {
		return u.gagalValidasi(ctx, idLaporan, err)
	}
	return nil
}

func (u *ValidasiAIUsecase) gagalValidasi(ctx context.Context, idLaporan string, cause error) error {
	if u.repo == nil {
		return cause
	}
	if err := u.repo.TandaiGagal(ctx, idLaporan, cause.Error()); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func normalizeAIResponse(response *domain.ValidasiAIResponse) error {
	if response == nil {
		return errors.New("respons AI kosong")
	}
	if !response.Success {
		return errors.New("AI tidak berhasil memvalidasi foto")
	}
	if response.AIStatus != "selesai" {
		return errors.New("ai_status AI tidak dikenal atau belum selesai")
	}
	if response.PhotoStatus == nil {
		return errors.New("photo_status AI wajib diisi")
	}
	if response.Confidence == nil || *response.Confidence < 0 || *response.Confidence > 1 {
		return errors.New("confidence AI wajib berada dalam rentang 0 sampai 1")
	}
	if response.ModelVersion == nil || *response.ModelVersion == "" {
		return errors.New("model_version AI wajib diisi")
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
		default:
			return errors.New("lighting_class AI tidak dikenal")
		}
	}

	photoLayak := *response.PhotoStatus == "layak_dinilai"
	if !photoLayak && *response.PhotoStatus != "tidak_layak" {
		return errors.New("photo_status AI tidak didukung")
	}
	if photoLayak && level == 0 {
		return errors.New("AI menyatakan foto layak tetapi tidak memberikan lighting_class")
	}
	if photoLayak && response.AIStatus != "selesai" {
		return errors.New("status AI tidak konsisten untuk foto layak")
	}
	result := "tidak_valid"
	var levelPtr *int16
	if photoLayak {
		result = "valid"
		levelPtr = &level
	}

	response.Data = domain.ValidasiAIResponseData{
		Success: response.Success, RequestID: response.RequestID,
		AIStatus: response.AIStatus, PhotoStatus: response.PhotoStatus,
		LightingClass: response.LightingClass, Confidence: response.Confidence,
		ModelVersion: response.ModelVersion, Message: response.Message,
		HasilValidasi: result, TingkatKegelapanAI: levelPtr,
		CNNConfidence: response.Confidence, Catatan: response.Message,
	}
	return nil
}

func (u *ValidasiAIUsecase) RetryValidasi(ctx context.Context, idLaporan string, laporanRepo *repository.LaporanRepository) error {
	if idLaporan == "" {
		return errors.New("id laporan wajib diisi")
	}
	if u.repo == nil || u.aiClient == nil {
		return errors.New("layanan validasi AI tidak terkonfigurasi")
	}
	status, err := u.repo.StatusLaporan(ctx, idLaporan)
	if err != nil {
		return err
	}
	if status != "gagal_validasi" {
		return errors.New("retry hanya diizinkan untuk laporan berstatus gagal_validasi")
	}
	if err := u.ProsesValidasiAI(ctx, idLaporan, laporanRepo); err != nil {
		return err
	}
	return nil
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
	if req == nil {
		return errors.New("request mock wajib diisi")
	}
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

func (u *ValidasiAIUsecase) TandaiGagal(ctx context.Context, idLaporan, alasan string) error {
	return u.repo.TandaiGagal(ctx, idLaporan, alasan)
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
