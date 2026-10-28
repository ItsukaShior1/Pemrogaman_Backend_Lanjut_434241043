package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-student/app/model"
)

// CourseRepository membaca data kelas dan KRS dari PostgreSQL. Data kelas
// tetap di database relasional karena isinya berstruktur tetap dan perlu
// constraint.
type CourseRepository interface {
	FindAll(ctx context.Context) ([]model.Course, error)
	FindByID(ctx context.Context, id int) (model.Course, error)
	IsEnrolled(ctx context.Context, courseID, userID int) (bool, error)
}

type coursePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &coursePostgresRepository{pool: pool}
}

func (r *coursePostgresRepository) FindAll(ctx context.Context) ([]model.Course, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, code, name, semester, created_at
		 FROM courses ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar kelas: %w", err)
	}
	defer rows.Close()

	result := []model.Course{}
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(
			&c.ID, &c.Code, &c.Name, &c.Semester, &c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("membaca baris kelas: %w", err)
		}
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query kelas: %w", err)
	}
	return result, nil
}

func (r *coursePostgresRepository) FindByID(ctx context.Context, id int) (model.Course, error) {
	var c model.Course
	err := r.pool.QueryRow(ctx,
		`SELECT id, code, name, semester, created_at
		 FROM courses WHERE id = $1`, id,
	).Scan(&c.ID, &c.Code, &c.Name, &c.Semester, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("mengambil kelas: %w", err)
	}
	return c, nil
}

// IsEnrolled dipakai layer lain -- misalnya layanan ulasan di MongoDB -- yang
// perlu tahu apakah seorang mahasiswa terdaftar di sebuah kelas.
//
// MongoDB tidak punya foreign key ke PostgreSQL, jadi pemeriksaan ini wajib
// dilakukan sendiri sebelum menulis ke MongoDB.
func (r *coursePostgresRepository) IsEnrolled(
	ctx context.Context, courseID, userID int,
) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM enrollments WHERE course_id = $1 AND user_id = $2
		)`,
		courseID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("memeriksa pendaftaran: %w", err)
	}
	return exists, nil
}
