package seeder

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func Migrate(pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sql, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		log.Fatalf("gagal baca file migrasi: %v", err)
	}

	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		log.Fatalf("gagal migrasi: %v", err)
	}
	fmt.Println("migrasi berhasil")
}

func Seed(pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hash := func(pw string) string {
		h, _ := bcrypt.GenerateFromPassword([]byte(pw), 10)
		return string(h)
	}

	// 1 admin
	adminPw := hash("admin123")
	pool.Exec(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, $2, 'admin') ON CONFLICT (email) DO NOTHING`,
		"admin@siakad.ac.id", adminPw)
	fmt.Println("admin seeded")

	// 20 mahasiswa
	prodis := []string{"Sistem Informasi", "Teknik Informatika", "Manajemen Informatika", "Bisnis Digital"}
	for i := 1; i <= 20; i++ {
		nim := fmt.Sprintf("1872210000%02d", i)
		email := fmt.Sprintf("mhs%02d@siakad.ac.id", i)
		nama := fmt.Sprintf("Mahasiswa %02d", i)
		prodi := prodis[i%len(prodis)]
		angkatan := 2022 + (i % 3)
		ipk := 2.50 + float64(i%6)*0.25

		pw := hash(nim)
		var userID int
		err := pool.QueryRow(ctx,
			`INSERT INTO users (email, password, role) VALUES ($1, $2, 'mahasiswa')
			 ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email RETURNING id`,
			email, pw).Scan(&userID)
		if err != nil {
			log.Printf("skip mahasiswa %d: %v", i, err)
			continue
		}

		_, err = pool.Exec(ctx,
			`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			 VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (nim) DO NOTHING`,
			userID, nim, nama, prodi, angkatan, ipk)
		if err != nil {
			log.Printf("skip student %d: %v", i, err)
		}
	}
	fmt.Println("20 mahasiswa seeded")

	// 10 mata kuliah
	matkul := []struct{ kode, nama string; sks, sem, kuota int }{
		{"IF101", "Algoritma dan Pemrograman", 3, 1, 40},
		{"IF102", "Basis Data", 3, 2, 35},
		{"IF201", "Struktur Data", 3, 3, 30},
		{"IF202", "Pemrograman Web", 3, 4, 40},
		{"IF301", "Rekayasa Perangkat Lunak", 3, 5, 35},
		{"IF302", "Jaringan Komputer", 3, 5, 30},
		{"IF303", "Kecerdasan Buatan", 3, 6, 25},
		{"IF401", "Proyek Akhir", 6, 7, 20},
		{"IF402", "Etika Profesi", 2, 7, 50},
		{"IF403", "Cloud Computing", 3, 6, 30},
	}
	for _, m := range matkul {
		pool.Exec(ctx,
			`INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota)
			 VALUES ($1, $2, $3, $4, $5) ON CONFLICT (kode_mk) DO NOTHING`,
			m.kode, m.nama, m.sks, m.sem, m.kuota)
	}
	fmt.Println("10 mata kuliah seeded")
	fmt.Println("seeding selesai")
}
