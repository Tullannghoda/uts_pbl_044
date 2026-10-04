package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB() *pgxpool.Pool {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:@localhost:5432/uts_044?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("gagal ping database: %v", err)
	}

	fmt.Println("database terhubung")
	return pool
}
