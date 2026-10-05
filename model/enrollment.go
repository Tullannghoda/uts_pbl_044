package model

import "time"

type Enrollment struct {
	ID             int       `json:"id"`
	StudentID      int       `json:"student_id"`
	CourseID       int       `json:"course_id"`
	TahunAkademik  string    `json:"tahun_akademik"`
	CreatedAt      time.Time `json:"created_at"`
}

type EnrolledCourse struct {
	EnrollmentID  int    `json:"enrollment_id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	TahunAkademik string `json:"tahun_akademik"`
}

type CreateEnrollmentRequest struct {
	CourseID      int    `json:"course_id" validate:"required"`
	TahunAkademik string `json:"tahun_akademik" validate:"required"`
}
