package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func ConnectDB() *pgxpool.Pool {
	// 1. Muat file .env secara paksa di sini
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL belum diatur di file .env")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Gagal membaca konfigurasi database: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Gagal terhubung ke Supabase: %v", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database tidak merespon: %v", err)
	}

	fmt.Println("Berhasil terhubung ke Database Supabase!")
	return pool
}
