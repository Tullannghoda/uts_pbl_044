package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"uts-044/model"
)

type StudentRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{pool: pool}
}

func (r *StudentRepository) FindAll(ctx context.Context, page, perPage int, prodi string, angkatan int, search, sort string) ([]model.Student, int, error) {
	where := " WHERE s.deleted_at IS NULL"
	args := []any{}
	idx := 1

	if prodi != "" {
		where += fmt.Sprintf(" AND s.prodi = $%d", idx)
		args = append(args, prodi)
		idx++
	}
	if angkatan > 0 {
		where += fmt.Sprintf(" AND s.angkatan = $%d", idx)
		args = append(args, angkatan)
		idx++
	}
	if search != "" {
		where += fmt.Sprintf(" AND (s.nim ILIKE $%d OR s.nama ILIKE $%d)", idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students s"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	orderBy := " ORDER BY s.id ASC"
	if sort != "" {
		if strings.HasPrefix(sort, "-") {
			col := strings.TrimPrefix(sort, "-")
			orderBy = fmt.Sprintf(" ORDER BY s.%s DESC", col)
		} else {
			orderBy = fmt.Sprintf(" ORDER BY s.%s ASC", sort)
		}
	}

	offset := (page - 1) * perPage
	args = append(args, perPage, offset)
	query := fmt.Sprintf(
		`SELECT s.id, s.user_id, s.nim, s.nama, s.prodi, s.angkatan, s.ipk_terakhir, s.created_at
		 FROM students s%s%s LIMIT $%d OFFSET $%d`, where, orderBy, idx, idx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		students = append(students, s)
	}
	return students, total, rows.Err()
}

func (r *StudentRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at
		 FROM students WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrNotFound
	}
	return s, err
}

func (r *StudentRepository) FindByUserID(ctx context.Context, userID int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at
		 FROM students WHERE user_id = $1 AND deleted_at IS NULL`, userID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrNotFound
	}
	return s, err
}

func (r *StudentRepository) Create(ctx context.Context, tx pgx.Tx, s model.Student) (model.Student, error) {
	err := tx.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		s.UserID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir,
	).Scan(&s.ID, &s.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return s, ErrDuplicate
		}
		return s, err
	}
	return s, nil
}

func (r *StudentRepository) CreateUser(ctx context.Context, tx pgx.Tx, email, hashedPassword string) (int, error) {
	var id int
	err := tx.QueryRow(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, $2, 'mahasiswa') RETURNING id`,
		email, hashedPassword,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, ErrDuplicate
		}
		return 0, err
	}
	return id, nil
}

func (r *StudentRepository) Update(ctx context.Context, s model.Student) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE students SET nama=$1, prodi=$2, angkatan=$3, ipk_terakhir=$4
		 WHERE id=$5 AND deleted_at IS NULL
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at`,
		s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir, s.ID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrNotFound
	}
	return s, err
}

func (r *StudentRepository) SoftDelete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *StudentRepository) Pool() *pgxpool.Pool {
	return r.pool
}
