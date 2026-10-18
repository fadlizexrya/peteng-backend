package domain

import "time"

type ValidasiAI struct {
	IDValidasi         string    `json:"id_validasi"`
	IDLaporan          string    `json:"id_laporan"`
	HasilValidasi      string    `json:"hasil_validasi"`
	TingkatKegelapanAI int16     `json:"tingkat_kegelapan_ai"`
	CNNConfidence      *float64  `json:"cnn_confidence"`
	ObjekTerdeteksi    any       `json:"objek_terdeteksi"`
	YOLOConfidence     *float64  `json:"yolo_confidence"`
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

// Request internal ke AI service.
type ValidasiAIRequest struct {
	IDLaporan              string `json:"id_laporan"`
	FotoURL                string `json:"foto_url"`
	Kategori               string `json:"kategori"`
	TingkatKegelapanManual int16  `json:"tingkat_kegelapan_manual"`
}

// Response dari AI service.
type ValidasiAIResponseData struct {
	HasilValidasi      string   `json:"hasil_validasi"`
	TingkatKegelapanAI int16    `json:"tingkat_kegelapan_ai"`
	ObjekTerdeteksi    any      `json:"objek_terdeteksi"`
	CNNConfidence      *float64 `json:"cnn_confidence"`
	YOLOConfidence     *float64 `json:"yolo_confidence"`
	ModelVersion       string   `json:"model_version"`
	Catatan            string   `json:"catatan"`
}

type ValidasiAIResponse struct {
	Status string                 `json:"status"`
	Data   ValidasiAIResponseData `json:"data"`
}
