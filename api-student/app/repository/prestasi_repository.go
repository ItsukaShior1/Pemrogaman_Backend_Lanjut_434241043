package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-student/app/model"
)

type PrestasiRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Prestasi, int, error)
	FindByID(ctx context.Context, id int) (model.Prestasi, error)
	Create(ctx context.Context, p model.CreatePrestasiRequest) (model.Prestasi, error)
	Update(ctx context.Context, id int, p model.CreatePrestasiRequest) (model.Prestasi, error)
	Delete(ctx context.Context, id int) error
}

var kolomUrutPrestasi = map[string]string{
	"id":          "id_prestasi",
	"id_prestasi": "id_prestasi",
	"student_id":  "student_id",
	"title":       "title",
	"champion":    "champion",
	"created_at":  "created_at",
}

type prestasiPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPrestasiRepository(pool *pgxpool.Pool) PrestasiRepository {
	return &prestasiPostgresRepository{
		pool: pool,
	}
}

func (r *prestasiPostgresRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Prestasi, int, error) {
	where, args := buildFilterPrestasi(q)
	query := fmt.Sprintf("SELECT * FROM prestasi %s ORDER BY %s LIMIT $%d OFFSET $%d", where, kolomUrutPrestasi[q.Sort], len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, query, append(args, q.Limit, q.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var prestasis []model.Prestasi
	for rows.Next() {
		var p model.Prestasi
		if err := rows.Scan(&p.IDPrestasi, &p.StudentID, &p.Title, &p.Champion, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		prestasis = append(prestasis, p)
	}
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM prestasi%s", where)
	err = r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	return prestasis, total, nil
}

func buildFilterPrestasi(q model.ListQuery) (string, []any) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND title ILIKE $%d", len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	return where, args
}

func (r *prestasiPostgresRepository) FindByID(ctx context.Context, id int) (model.Prestasi, error) {
	var p model.Prestasi
	err := r.pool.QueryRow(ctx, "SELECT id_prestasi, student_id, title, champion, created_at FROM prestasi WHERE id_prestasi = $1", id).Scan(&p.IDPrestasi, &p.StudentID, &p.Title, &p.Champion, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Prestasi{}, ErrNotFound
		}
		return model.Prestasi{}, fmt.Errorf("mengambil prestasi: %w", err)
	}
	return p, nil
}

func (r *prestasiPostgresRepository) Create(ctx context.Context, p model.CreatePrestasiRequest) (model.Prestasi, error) {
	var newPrestasi model.Prestasi
	err := r.pool.QueryRow(ctx, "INSERT INTO prestasi (student_id, title, champion) VALUES ($1, $2, $3) RETURNING id_prestasi, student_id, title, champion, created_at", p.StudentID, p.Title, p.Champion).Scan(&newPrestasi.IDPrestasi, &newPrestasi.StudentID, &newPrestasi.Title, &newPrestasi.Champion, &newPrestasi.CreatedAt)
	if err != nil {
		return model.Prestasi{}, err
	}
	return newPrestasi, nil
}

func (r *prestasiPostgresRepository) Update(ctx context.Context, id int, p model.CreatePrestasiRequest) (model.Prestasi, error) {
	var updatedPrestasi model.Prestasi
	err := r.pool.QueryRow(ctx, "UPDATE prestasi SET student_id = $1, title = $2, champion = $3 WHERE id_prestasi = $4 RETURNING id_prestasi, student_id, title, champion, created_at", p.StudentID, p.Title, p.Champion, id).Scan(&updatedPrestasi.IDPrestasi, &updatedPrestasi.StudentID, &updatedPrestasi.Title, &updatedPrestasi.Champion, &updatedPrestasi.CreatedAt)
	if err != nil {
		return model.Prestasi{}, err
	}
	return updatedPrestasi, nil
}

func (r *prestasiPostgresRepository) Delete(ctx context.Context, id int) error {
	cmdTag, err := r.pool.Exec(ctx, "DELETE FROM prestasi WHERE id_prestasi = $1", id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("prestasi not found")
	}
	return nil
}
