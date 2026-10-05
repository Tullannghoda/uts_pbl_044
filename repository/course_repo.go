package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"uts-044/model"
)

type CourseRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{pool: pool}
}

func (r *CourseRepository) FindAll(ctx context.Context, semester int, search string, availableOnly bool) ([]model.Course, error) {
	where := " WHERE 1=1"
	args := []any{}
	idx := 1

	if semester > 0 {
		where += fmt.Sprintf(" AND c.semester = $%d", idx)
		args = append(args, semester)
		idx++
	}
	if search != "" {
		where += fmt.Sprintf(" AND (c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	if availableOnly {
		where += " AND c.kuota - COALESCE(e.terisi, 0) > 0"
	}

	query := fmt.Sprintf(
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		        COALESCE(e.terisi, 0) AS terisi,
		        c.kuota - COALESCE(e.terisi, 0) AS sisa_kuota
		 FROM courses c
		 LEFT JOIN (SELECT course_id, COUNT(*) AS terisi FROM enrollments GROUP BY course_id) e
		 ON c.id = e.course_id
		 %s ORDER BY c.id ASC`, where)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []model.Course
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

func (r *CourseRepository) FindByID(ctx context.Context, id int) (model.Course, error) {
	var c model.Course
	err := r.pool.QueryRow(ctx,
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		        COALESCE(e.terisi, 0), c.kuota - COALESCE(e.terisi, 0)
		 FROM courses c
		 LEFT JOIN (SELECT course_id, COUNT(*) AS terisi FROM enrollments GROUP BY course_id) e
		 ON c.id = e.course_id
		 WHERE c.id = $1`, id,
	).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota)
	return c, err
}
