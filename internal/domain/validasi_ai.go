package domain

import (
	"encoding/json"
	"time"
)

type ValidasiAI struct {
	IDValidasi         string    `json:"id_validasi"`
	IDLaporan          string    `json:"id_laporan"`
	HasilValidasi      string    `json:"hasil_validasi"`
	TingkatKegelapanAI *int16    `json:"tingkat_kegelapan_ai"`
	CNNConfidence      *float64  `json:"cnn_confidence"`
	ObjekTerdeteksi    any       `json:"objek_terdeteksi"`
	Catatan            *string   `json:"catatan"`
	ModelVersion       *string   `json:"model_version"`
	TanggalValidasi    time.Time `json:"tanggal_validasi"`
}

// Request untuk simulasi hasil validasi.
// Nilai ini dikirim secara eksplisit untuk pengujian,
// bukan hasil prediksi AI sungguhan.
type ValidasiAIMockRequest struct {
	HasilValidasi      string `json:"hasil_validasi"`
	TingkatKegelapanAI int16  `json:"tingkat_kegelapan_ai"`
	Catatan            string `json:"catatan"`
}

// Response internal dari AI PETENG.
type ValidasiAIResponseData struct {
	// Field berikut adalah kontrak AI PETENG.
	Success       bool     `json:"success"`
	RequestID     string   `json:"request_id"`
	AIStatus      string   `json:"ai_status"`
	PhotoStatus   *string  `json:"photo_status"`
	LightingClass *string  `json:"lighting_class"`
	Confidence    *float64 `json:"confidence"`
	ModelVersion  *string  `json:"model_version"`
	Message       string   `json:"message"`

	// Field legacy dipertahankan agar repository dan API lama tetap kompatibel.
	HasilValidasi      string   `json:"hasil_validasi"`
	TingkatKegelapanAI *int16   `json:"tingkat_kegelapan_ai"`
	ObjekTerdeteksi    any      `json:"objek_terdeteksi"`
	CNNConfidence      *float64 `json:"cnn_confidence"`
	Catatan            string   `json:"catatan"`

	// Internal processing metadata for persistence.
	HasilJSON json.RawMessage `json:"-"`
}

type ValidasiAIResponse struct {
	// Kontrak AI PETENG dikembalikan flat, sedangkan Data digunakan sebagai
	// model internal setelah normalisasi di usecase.
	Success       bool     `json:"success"`
	RequestID     string   `json:"request_id"`
	AIStatus      string   `json:"ai_status"`
	PhotoStatus   *string  `json:"photo_status"`
	LightingClass *string  `json:"lighting_class"`
	Confidence    *float64 `json:"confidence"`
	ModelVersion  *string  `json:"model_version"`
	Message       string   `json:"message"`

	Data ValidasiAIResponseData `json:"-"`
}
