package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"uts-044/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// activeUserQuery hanya mengembalikan user yang masih aktif: admin selalu lolos
// (tidak punya baris students), sedangkan mahasiswa yang datanya sudah di-soft-delete
// (students.deleted_at terisi) tidak boleh lagi login maupun diakses sesi /auth/me-nya.
const activeUserQuery = `
	SELECT u.id, u.email, u.password, u.role, u.created_at
	FROM users u
	LEFT JOIN students s ON s.user_id = u.id
	WHERE %s AND (s.id IS NULL OR s.deleted_at IS NULL)`

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	query := fmt.Sprintf(activeUserQuery, "u.email = $1")
	err := r.pool.QueryRow(ctx, query, email).
		Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	query := fmt.Sprintf(activeUserQuery, "u.id = $1")
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

var ErrNotFound = errors.New("data tidak ditemukan")
var ErrDuplicate = errors.New("data sudah ada")