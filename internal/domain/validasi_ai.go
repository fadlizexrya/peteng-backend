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
