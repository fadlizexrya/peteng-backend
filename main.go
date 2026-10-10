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
	ruasJalanRepo := repository.NewRuasJalanRepository(db)

	lokasiUsecase := usecase.NewLokasiUsecase(
		lokasiRepo,
	)

	lokasiHandler := deliveryHttp.NewLokasiHandler(
		lokasiUsecase,
	)

	// SETUP VALIDASI AI
	validasiAIRepo := repository.NewValidasiAIRepository(db)
	validasiAIUsecase := usecase.NewValidasiAIUsecase(validasiAIRepo)
	validasiAIHandler := deliveryHttp.NewValidasiAIHandler(validasiAIUsecase)

	// SETUP NAVIGASI
	navigasiUsecase := usecase.NewNavigasiUsecase(ruasJalanRepo)
	navigasiHandler := deliveryHttp.NewNavigasiHandler(navigasiUsecase)

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
			"Patch",
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

	// PUBLIC ROUTES
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

	r.Get(
		"/api/navigasi/rute",
		navigasiHandler.CariRute,
	)

	// PROTECTED ROUTES WARGA
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

		// Endpoint untuk mengambil laporan warga
		r.Get(
			"/laporan",
			laporanHandler.GetMyReports,
		)

		// Endpoint untuk mengambil detail laporan warga
		r.Get(
			"/laporan/{id_laporan}",
			laporanHandler.GetDetail,
		)
	})

	// PROTECTED ROUTES DINAS PERHUBUNGAN
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

		r.Get(
			"/laporan",
			laporanHandler.GetAllForDishub,
		)

		r.Patch(
			"/laporan/{id_laporan}/status",
			laporanHandler.UpdateStatusDishub,
		)

		// Endpoint validasi AI untuk pengembangan
		r.Post(
			"/laporan/{id_laporan}/validasi-ai/mulai",
			validasiAIHandler.Mulai,
		)

		r.Post(
			"/laporan/{id_laporan}/validasi-ai/mock",
			validasiAIHandler.SimpanHasilMock,
		)

		r.Get(
			"/laporan/{id_laporan}/validasi-ai",
			validasiAIHandler.GetTerbaru,
		)

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
