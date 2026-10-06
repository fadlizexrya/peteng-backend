package main

import (
	"log"
	"net/http"
	"os"

	"peteng-backend/config"
	deliveryHttp "peteng-backend/internal/delivery/http"
	"peteng-backend/internal/repository"
	"peteng-backend/internal/usecase"
	"peteng-backend/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	// 1. PANGGIL godotenv TERLEBIH DAHULU!
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: Gagal membaca file .env atau file .env tidak ditemukan")
	}

	// 2. BARU PANGGIL KONEKSI DATABASE
	db := config.ConnectDB()
	defer db.Close()

	// 3. Setup Dependency & Router
	wargaRepo := repository.NewWargaRepository(db)
	jwtManager := auth.NewJWTManager()
	wargaUsecase := usecase.NewWargaUsecase(wargaRepo, jwtManager)
	wargaHandler := deliveryHttp.NewWargaHandler(wargaUsecase)

	// 4. Setup Dinas Perhubungan Dependencies
	dinasRepo := repository.NewDinasPerhubunganRepository(db)
	dinasUsecase := usecase.NewDinasPerhubunganUsecase(dinasRepo, jwtManager)
	dinasHandler := deliveryHttp.NewDinasPerhubunganHandler(dinasUsecase)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Post("/api/warga/register", wargaHandler.Register)
	r.Post("/api/warga/login", wargaHandler.Login)
	r.Post("/api/warga/laporan", wargaHandler.LaporkanJalanGelap)

	// Route API Dinas Perhubungan (Baru)
	r.Post("/api/dinas/register", dinasHandler.Register)
	r.Post("/api/dinas/login", dinasHandler.Login)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server PETENG running di port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
