package main

import "time"

// Student adalah entitas inti yang diangkat dari struct pada tugas pertemuan 1.
// Field NIM ditambahkan sebagai penanda unik mahasiswa (Nomor Induk Mahasiswa).
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateStudentRequest dipakai pada POST /api/v1/students.
// Seluruh field wajib dikirim dan tidak boleh kosong.
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// ReplaceStudentRequest dipakai pada PUT /api/v1/students/:id.
// PUT berarti mengganti SELURUH isi, sehingga semua field bertipe biasa
// dan semuanya wajib dikirim. Field yang tidak dikirim akan dianggap kosong.
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest dipakai pada PATCH /api/v1/students/:id.
// PATCH berarti mengubah sebagian isi. Karena itu semua field bertipe pointer,
// supaya server bisa membedakan "tidak dikirim" (nil) dengan "dikirim bernilai
// default" (misalnya 0 untuk Grade atau false untuk IsActive).
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// WebResponse adalah amplop baku untuk SEMUA respons, berhasil maupun gagal.
// Konsistensi bentuk ini membuat klien tidak perlu menebak struktur tiap endpoint.
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Meta memuat informasi paginasi yang dikirim bersama respons daftar.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListQuery menampung hasil parse query string pada endpoint daftar.
// Pointer untuk GradeMin/GradeMax dipakai supaya bisa membedakan
// "tidak dikirim" dengan "dikirim 0".
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