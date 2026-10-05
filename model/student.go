package model

import "time"

type Student struct {
	ID           int        `json:"id"`
	UserID       int        `json:"user_id"`
	NIM          string     `json:"nim"`
	Nama         string     `json:"nama"`
	Prodi        string     `json:"prodi"`
	Angkatan     int        `json:"angkatan"`
	IPKTerakhir  float64    `json:"ipk_terakhir"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type StudentDetail struct {
	Student
	MataKuliah []EnrolledCourse `json:"mata_kuliah"`
	TotalSKS   int              `json:"total_sks"`
	BatasSKS   int              `json:"batas_sks"`
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim" validate:"required,len=12,numeric"`
	Nama        string   `json:"nama" validate:"required"`
	Email       string   `json:"email" validate:"required,email"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required,min=2000,max=2026"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,min=0,max=4"`
}

type UpdateStudentRequest struct {
	Nama        string   `json:"nama" validate:"required"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required,min=2000,max=2026"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,min=0,max=4"`
}
