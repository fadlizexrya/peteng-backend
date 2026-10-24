package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func ConnectDB() *pgxpool.Pool {
	// Memuat konfigurasi dari file .env.
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL belum diatur di file .env")
	}

	// Membaca konfigurasi koneksi PostgreSQL.
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Gagal membaca konfigurasi database: %v", err)
	}

	// Menghindari cache prepared statement bernama.
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	// Membuat connection pool.
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Gagal membuat koneksi ke Supabase: %v", err)
	}

	// Memastikan koneksi database berfungsi.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		log.Fatalf("Database tidak merespons: %v", err)
	}

	fmt.Println("Berhasil terhubung ke Database Supabase!")
	return pool
}
