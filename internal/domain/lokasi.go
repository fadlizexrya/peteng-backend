package domain

type Lokasi struct {
	IDLokasi  int     `json:"id_lokasi"`
	Alamat    string  `json:"alamat"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}