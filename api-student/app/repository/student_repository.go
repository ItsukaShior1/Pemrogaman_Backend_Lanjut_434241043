// Package repository berisi lapisan akses data. Ia berbicara langsung
// dengan PostgreSQL lewat pgx dan menerjemahkan baris menjadi struct model.
//
// Catatan: package ini TIDAK mengimpor gofiber. Ia hanya butuh context,
// driver pgx, dan package model. Dengan begitu, repository mudah diuji
// dengan database asli (integration test) atau diganti ke driver lain.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-student/app/model"
)

// ErrNotFound dikembalikan saat data yang diminta tidak ada.
var ErrNotFound = errors.New("data tidak ditemukan")

// ErrDuplicate dikembalikan saat terjadi pelanggaran constraint unik
// (mis. NIM yang sudah dipakai mahasiswa lain).
var ErrDuplicate = errors.New("data sudah ada")

// StudentRepository adalah kontrak yang harus dipenuhi implementasi manapun.
// Service hanya bergantung pada interface ini, bukan pada struct konkret.
type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

// kolomUrut memetakan nama field yang diizinkan dari query string ke
// nama kolom sebenarnya di database. Whitelist ini mencegah SQL injection
// karena klien tidak boleh menentukan kolom urut secara bebas.
var kolomUrut = map[string]string{
	"id":         "id",
	"nim":        "nim",
	"name":       "name",
	"grade":      "grade",
	"created_at": "created_at",
}

// studentPostgresRepository adalah implementasi StudentRepository di atas
// PostgreSQL dengan pgxpool.
type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewStudentRepository membuat instance repository baru dari sebuah pool.
func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

// buildFilter menyusun fragmen WHERE dan argumen sesuai ListQuery.
// Dipakai oleh FindAll agar query COUNT dan query SELECT konsisten.
func buildFilter(q model.ListQuery) (string, []any) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	if q.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *q.IsActive)
	}

	if q.GradeMin != nil {
		where += fmt.Sprintf(" AND grade >= $%d", len(args)+1)
		args = append(args, *q.GradeMin)
	}

	if q.GradeMax != nil {
		where += fmt.Sprintf(" AND grade <= $%d", len(args)+1)
		args = append(args, *q.GradeMax)
	}

	return where, args
}

// FindAll mengambil daftar student dengan paginasi, pencarian, dan filter.
func (r *studentPostgresRepository) FindAll(
	ctx context.Context, q model.ListQuery,
) ([]model.Student, int, error) {
	where, args := buildFilter(q)

	var total int
	if err := r.pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM students"+where,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("menghitung student: %w", err)
	}

	arah := "ASC"
	if q.Order == "desc" {
		arah = "DESC"
	}

	sqlText := fmt.Sprintf(
		`SELECT id, nim, name, grade, is_active, created_at
		 FROM students%s
		 ORDER BY %s %s
		 LIMIT $%d OFFSET $%d`,
		where,
		kolomUrut[q.Sort],
		arah,
		len(args)+1,
		len(args)+2,
	)

	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar student: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(
			&s.ID, &s.NIM, &s.Name,
			&s.Grade, &s.IsActive, &s.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("membaca baris student: %w", err)
		}
		hasil = append(hasil, s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query: %w", err)
	}

	return hasil, total, nil
}

// FindByID mengambil satu student berdasarkan id.
func (r *studentPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.Student, error) {
	var s model.Student

	err := r.pool.QueryRow(
		ctx,
		`SELECT id, nim, name, grade, is_active, created_at
		 FROM students WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student: %w", err)
	}

	return s, nil
}

// Create menyimpan student baru. Mengembalikan ErrDuplicate bila NIM bentrok.
func (r *studentPostgresRepository) Create(
	ctx context.Context, s model.Student,
) (model.Student, error) {
	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO students (nim, name, grade, is_active)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive,
	).Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan student: %w", err)
	}

	return s, nil
}

// Update memperbarui seluruh field student. PUT dan PATCH keduanya memakai
// repository method ini; service yang memutuskan field mana yang berubah.
func (r *studentPostgresRepository) Update(
	ctx context.Context, s model.Student,
) (model.Student, error) {
	err := r.pool.QueryRow(
		ctx,
		`UPDATE students SET nim = $1, name = $2, grade = $3, is_active = $4
		 WHERE id = $5
		 RETURNING id, nim, name, grade, is_active, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive, s.ID,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("memperbarui student: %w", err)
	}

	return s, nil
}

// Delete menghapus student berdasarkan id. Mengembalikan ErrNotFound bila
// id tidak ada (tidak ada baris yang terpengaruh).
func (r *studentPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(
		ctx,
		`DELETE FROM students WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("menghapus student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// isUniqueViolation mengecek apakah error dari pgx adalah pelanggaran
// constraint unik. Kode "23505" adalah kode SQLSTATE untuk unique violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
