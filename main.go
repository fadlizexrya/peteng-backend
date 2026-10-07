package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"peteng-backend/config"
	"peteng-backend/internal/auth"
	deliveryHttp "peteng-backend/internal/delivery/http"
	petengMiddleware "peteng-backend/internal/middleware"
	"peteng-backend/internal/repository"
	"peteng-backend/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	// LOAD ENVIRONMENT VARIABLES
	err := godotenv.Load()

	if err != nil {
		log.Println(
			"Peringatan: Gagal membaca file .env atau file .env tidak ditemukan",
		)
	}

	// DATABASE CONNECTION
	db := config.ConnectDB()
	defer db.Close()

	// JWT MANAGER
	jwtManager := auth.NewJWTManager()

	// SETUP WARGA
	wargaRepo := repository.NewWargaRepository(db)

	wargaUsecase := usecase.NewWargaUsecase(
		wargaRepo,
		jwtManager,
	)

	wargaHandler := deliveryHttp.NewWargaHandler(
		wargaUsecase,
	)

	// SETUP LAPORAN
	laporanRepo := repository.NewLaporanRepository(db)
	progresRepo := repository.NewProgresLaporanRepository(db)
	laporanUsecase := usecase.NewLaporanUsecase(
		laporanRepo,
		progresRepo,
	)
	laporanHandler := deliveryHttp.NewLaporanHandler(
		laporanUsecase,
	)

	// SETUP DINAS PERHUBUNGAN
	dinasRepo := repository.NewDinasPerhubunganRepository(db)

	dinasUsecase := usecase.NewDinasPerhubunganUsecase(
		dinasRepo,
		jwtManager,
	)

	dinasHandler := deliveryHttp.NewDinasPerhubunganHandler(
		dinasUsecase,
	)

	// SETUP LOKASI
	lokasiRepo := repository.NewLokasiRepository(db)

	lokasiUsecase := usecase.NewLokasiUsecase(
		lokasiRepo,
	)

	lokasiHandler := deliveryHttp.NewLokasiHandler(
		lokasiUsecase,
	)

	// ROUTER
	r := chi.NewRouter()

	// Middleware bawaan Chi
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	r.Use(cors.Handler(cors.Options{

		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:3000",
		},

		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},

		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},

		AllowCredentials: true,

		MaxAge: 300,
	}))

	// =========================================================
	// 8. PUBLIC ROUTES
	// =========================================================
	// Route di bawah TIDAK membutuhkan JWT.
	// Karena user belum memiliki token ketika melakukan
	// register dan login.
	// =========================================================

	// WARGA
	r.Post(
		"/api/warga/register",
		wargaHandler.Register,
	)

	r.Post(
		"/api/warga/login",
		wargaHandler.Login,
	)

	// DINAS PERHUBUNGAN
	r.Post(
		"/api/dinas/register",
		dinasHandler.Register,
	)

	r.Post(
		"/api/dinas/login",
		dinasHandler.Login,
	)

	// LOKASI
	r.Get(
		"/api/lokasi",
		lokasiHandler.GetAll,
	)

	// =========================================================
	// 9. PROTECTED ROUTES - WARGA
	// =========================================================
	// Semua route di dalam group ini:
	// 1. Harus mempunyai JWT
	// 2. JWT harus valid
	// 3. Role harus "warga"
	// =========================================================
	r.Route("/api/warga", func(r chi.Router) {

		// Validasi JWT
		r.Use(
			petengMiddleware.JWTMiddleware(
				jwtManager,
			),
		)

		// Validasi role
		r.Use(
			petengMiddleware.RequireRole(
				"warga",
			),
		)

		// Endpoint laporan warga
		r.Post(
			"/laporan",
			laporanHandler.Create,
		)
	})

	// =========================================================
	// 10. PROTECTED ROUTES - DINAS PERHUBUNGAN
	// =========================================================
	// Semua route di dalam group ini:
	// 1. Harus mempunyai JWT
	// 2. JWT harus valid
	// 3. Role harus "dishub"
	// =========================================================
	r.Route("/api/dinas", func(r chi.Router) {

		// Validasi JWT
		r.Use(
			petengMiddleware.JWTMiddleware(
				jwtManager,
			),
		)

		// Validasi role
		r.Use(
			petengMiddleware.RequireRole(
				"dishub",
			),
		)

		// =====================================================
		// TEST ENDPOINT
		// =====================================================
		// Endpoint ini sementara digunakan untuk memastikan
		// JWT + role Dishub bekerja.
		// GET /api/dinas/profile
		// =====================================================
		r.Get(
			"/profile",
			func(w http.ResponseWriter, r *http.Request) {

				userID := r.Context().Value(
					petengMiddleware.UserIDKey,
				)

				email := r.Context().Value(
					petengMiddleware.EmailKey,
				)

				role := r.Context().Value(
					petengMiddleware.RoleKey,
				)

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				json.NewEncoder(w).Encode(
					map[string]interface{}{
						"message": "JWT valid",
						"user_id": userID,
						"email":   email,
						"role":    role,
					},
				)
			},
		)

		// =====================================================
		// NANTI ENDPOINT DINAS AKAN DITAMBAHKAN DI SINI
		// =====================================================
		// Contoh:
		// r.Get("/laporan", dinasHandler.GetLaporan)
		// r.Put("/laporan/{id}", dinasHandler.UpdateLaporan)
		// r.Post("/artikel", dinasHandler.CreateArtikel)
		//
	})

	// SERVER PORT
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Printf(
		"Server PETENG running di port %s...",
		port,
	)

	// START SERVER
	log.Fatal(
		http.ListenAndServe(
			":"+port,
			r,
		),
	)
}
