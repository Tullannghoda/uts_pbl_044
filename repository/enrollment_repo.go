package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"uts-044/model"
)

type EnrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{pool: pool}
}

func (r *EnrollmentRepository) Create(ctx context.Context, tx pgx.Tx, e model.Enrollment) (model.Enrollment, error) {
	err := tx.QueryRow(ctx,
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		 VALUES ($1, $2, $3) RETURNING id, created_at`,
		e.StudentID, e.CourseID, e.TahunAkademik,
	).Scan(&e.ID, &e.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return e, ErrDuplicate
		}
		return e, err
	}
	return e, nil
}

func (r *EnrollmentRepository) FindByID(ctx context.Context, id int) (model.Enrollment, error) {
	var e model.Enrollment
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, course_id, tahun_akademik, created_at
		 FROM enrollments WHERE id = $1`, id,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return e, ErrNotFound
	}
	return e, err
}

func (r *EnrollmentRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM enrollments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *EnrollmentRepository) FindByStudentID(ctx context.Context, studentID int) ([]model.EnrolledCourse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT e.id, c.kode_mk, c.nama_mk, c.sks, e.tahun_akademik
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1
		 ORDER BY e.created_at DESC`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.EnrolledCourse
	for rows.Next() {
		var ec model.EnrolledCourse
		if err := rows.Scan(&ec.EnrollmentID, &ec.KodeMK, &ec.NamaMK, &ec.SKS, &ec.TahunAkademik); err != nil {
			return nil, err
		}
		list = append(list, ec)
	}
	return list, rows.Err()
}

func (r *EnrollmentRepository) TotalSKS(ctx context.Context, studentID int, tahunAkademik string) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahunAkademik,
	).Scan(&total)
	return total, err
}

func (r *EnrollmentRepository) CourseQuotaWithLock(ctx context.Context, tx pgx.Tx, courseID int) (kuota int, terisi int, sks int, err error) {
	err = tx.QueryRow(ctx,
		`SELECT c.kuota, COALESCE(e.cnt, 0), c.sks
		 FROM courses c
		 LEFT JOIN (SELECT course_id, COUNT(*) AS cnt FROM enrollments WHERE course_id = $1 GROUP BY course_id) e
		 ON c.id = e.course_id
		 WHERE c.id = $1 FOR UPDATE OF c`, courseID,
	).Scan(&kuota, &terisi, &sks)
	return
}

func (r *EnrollmentRepository) Pool() *pgxpool.Pool {
	return r.pool
}
