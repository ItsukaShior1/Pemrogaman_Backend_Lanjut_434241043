// Package model berisi struct entitas Student dan DTO request/response.
//
// Catatan: package ini tidak mengimpor package apapun dari proyek ini sendiri
// maupun framework HTTP. Tujuannya agar layer Entities (paling dalam pada
// Clean Architecture) benar-benar murni dan dapat diuji tanpa dependency luar.
package model

import "time"

// Student adalah entitas inti yang merepresentasikan satu baris pada tabel
// students. Struct ini dipakai sebagai representasi data, bukan sebagai
// representasi permintaan HTTP.
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateStudentRequest adalah DTO untuk POST /students. Karena POST bermakna
// "buat baru", seluruh field wajib dikirim dan tipe datanya non-pointer.
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// ReplaceStudentRequest adalah DTO untuk PUT /students/:id. Karena PUT
// bermakna "ganti seluruh", seluruh field wajib dikirim.
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest adalah DTO untuk PATCH /students/:id. Karena PATCH
// bermakna "ubah sebagian", setiap field bertipe pointer supaya bisa
// membedakan "tidak dikirim" (nil) dari "dikirim bernilai default".
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// WebResponse adalah amplop respons seragam yang dipakai seluruh endpoint.
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Meta berisi informasi paginasi untuk respons daftar.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListQuery menampung seluruh parameter query string pada GET /students.
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
	GradeMin *float64
	GradeMax *float64
}

// Offset menghitung offset SQL berdasarkan halaman dan limit.
func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
