package model

import "time"

// Portfolio adalah bentuk portofolio yang dikenal seluruh aplikasi.
//
// Sama seperti Review, tidak ada tipe driver MongoDB di package model.
// Field student_id hanya berupa integer: referensinya ke students.id di
// PostgreSQL checked oleh service, karena tidak ada foreign key lintas
// database.
type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// Project disarangkan di dalam dokumen portofolio: selalu dibaca bersama
// portofolionya, jumlahnya dibatasi, dan tidak dipakai dokumen lain.
type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
	Year         int      `json:"year"`
}

type Certificate struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Issuer string    `json:"issuer"`
	Date   time.Time `json:"date"`
}

// ReplacePortfolioRequest adalah isi PUT /students/:id/portfolio.
// Projects boleh dikosongkan; bila diisi, isinya menggantikan seluruh daftar
// proyek lama.
type ReplacePortfolioRequest struct {
	Headline     string               `json:"headline"`
	Bio          string               `json:"bio"`
	Projects     []ProjectRequest     `json:"projects"`
	Certificates []CertificateRequest `json:"certificates"`
}

type ProjectRequest struct {
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
	Year         int      `json:"year"`
}

type CertificateRequest struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Date   string `json:"date"`
}

// AddProjectRequest adalah isi POST /students/:id/portfolio/projects.
type AddProjectRequest struct {
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
	Year         int      `json:"year"`
}

type TechnologyStat struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}
