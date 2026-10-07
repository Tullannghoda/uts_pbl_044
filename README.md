# UTS SIAKAD Mini API

Proyek Ujian Tengah Semester (UTS) mata kuliah Backend Lanjut. API ini merupakan miniatur Sistem Informasi Akademik (SIAKAD) yang melayani kebutuhan dasar pencatatan mahasiswa, mata kuliah, dan pengisian Kartu Rencana Studi (KRS).

## Fitur Utama
- **Autentikasi JWT**: Login dan Role-based Access Control (Admin & Mahasiswa).
- **Manajemen Mahasiswa (Admin)**: CRUD data mahasiswa dengan validasi ketat (NIM unik, format email, dll).
- **Manajemen Mata Kuliah**: Melihat daftar mata kuliah beserta sisa kuota yang tersedia secara *real-time*.
- **Pengisian KRS (Mahasiswa)**: Proses enrollment mata kuliah dengan *business rules* kompleks (batas SKS berdasarkan IPK, pencegahan duplikasi, dan validasi kuota menggunakan row-level locking).
- **Rate Limiting**: Perlindungan *brute force* pada endpoint login (maksimal 5x percobaan per menit per IP).
- **Soft Delete**: Data mahasiswa tidak dihapus permanen, melainkan disembunyikan menggunakan `deleted_at`.

## Teknologi
- **Bahasa**: Golang (Go 1.21+)
- **Framework**: Go Fiber v2
- **Database**: PostgreSQL 16
- **Driver Database**: pgxpool (jackc/pgx/v5)
- **Validasi**: go-playground/validator/v10
- **Keamanan**: bcrypt, golang-jwt

## Menjalankan Server
1. Sesuaikan DSN Database di `config/database.go`
2. Lakukan inisialisasi tabel: `go run . migrate`
3. Masukkan data dummy awal: `go run . seed`
4. Jalankan server API: `go run .`

*Dikerjakan oleh: Glennovian*
