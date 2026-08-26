package main

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// students adalah penyimpanan di memori. Setiap kali server dimatikan,
// seluruh data hilang. Penyimpanan persisten (basis data) baru dibahas
// pada pertemuan 3.
var students []Student
var nextStudentID = 1

// findStudentIndex mencari posisi mahasiswa berdasarkan ID.
// Mengembalikan -1 bila tidak ditemukan.
func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

// findStudentByNIM memeriksa apakah NIM sudah dipakai oleh mahasiswa lain.
// Abaikan pengecualian bila excludeID > 0 (untuk PUT/PATCH mahasiswa itu sendiri).
func findStudentByNIM(nim string, excludeID int) bool {
	for _, s := range students {
		if strings.EqualFold(s.NIM, nim) && s.ID != excludeID {
			return true
		}
	}
	return false
}

// cocokPencarian memeriksa apakah kata kunci muncul pada Name (case-insensitive).
func cocokPencarian(s Student, kata string) bool {
	return strings.Contains(strings.ToLower(s.Name), strings.ToLower(kata))
}

// paramID membaca parameter :id dari path dan memastikan nilainya valid.
func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// listStudents menampilkan daftar mahasiswa dengan paginasi, pencarian,
// penyaringan, dan pengurutan. Query string dijelaskan pada README.
func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)

	// 1) Saring (filter + search)
	hasil := []Student{}
	for _, s := range students {
		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue
		}
		if q.GradeMin != nil && s.Grade < *q.GradeMin {
			continue
		}
		if q.GradeMax != nil && s.Grade > *q.GradeMax {
			continue
		}
		if q.Search != "" && !cocokPencarian(s, q.Search) {
			continue
		}
		hasil = append(hasil, s)
	}

	// 2) Urutkan (sort + order)
	sort.SliceStable(hasil, func(i, j int) bool {
		var lebihKecil bool
		switch q.Sort {
		case "nim":
			lebihKecil = hasil[i].NIM < hasil[j].NIM
		case "name":
			lebihKecil = hasil[i].Name < hasil[j].Name
		case "grade":
			lebihKecil = hasil[i].Grade < hasil[j].Grade
		case "created_at":
			lebihKecil = hasil[i].CreatedAt.Before(hasil[j].CreatedAt)
		default:
			lebihKecil = hasil[i].ID < hasil[j].ID
		}
		if q.Order == "desc" {
			return !lebihKecil
		}
		return lebihKecil
	})

	// 3) Potong sesuai halaman
	total := len(hasil)
	totalPages := (total + q.Limit - 1) / q.Limit
	mulai := (q.Page - 1) * q.Limit
	if mulai > total {
		mulai = total
	}
	akhir := mulai + q.Limit
	if akhir > total {
		akhir = total
	}

	return okList(c, "daftar student berhasil diambil",
		hasil[mulai:akhir],
		&Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPages: totalPages,
		})
}

// getStudent mengambil satu mahasiswa berdasarkan ID, atau 404 bila tidak ada.
func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	return ok(c, "student ditemukan", students[i])
}

// createStudent membuat mahasiswa baru. Status 201 dengan header Location.
// 422 bila validasi isi gagal, 409 bila NIM bentrok, 415 bila Content-Type salah.
func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}
	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus bernilai antara 0 dan 100"
	}
	if len(errs) == 0 && findStudentByNIM(req.NIM, 0) {
		return fail(c, fiber.StatusConflict, "NIM sudah dipakai")
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	baru := Student{
		ID:        nextStudentID,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
	}
	students = append(students, baru)
	nextStudentID++

	return created(c, "student berhasil dibuat", baru,
		"/api/v1/students/"+strconv.Itoa(baru.ID))
}

// replaceStudent adalah handler PUT. Mengganti SELURUH isi mahasiswa.
// Field yang tidak dikirim akan dianggap kosong (lihat ReplaceStudentRequest).
func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if req.Name == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus bernilai antara 0 dan 100 pada PUT"
	}
	if len(errs) == 0 && findStudentByNIM(req.NIM, id) {
		return fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	students[i].NIM = req.NIM
	students[i].Name = req.Name
	students[i].Grade = req.Grade
	students[i].IsActive = req.IsActive

	return ok(c, "student berhasil diganti seluruhnya", students[i])
}

// patchStudent adalah handler PATCH. Hanya mengubah field yang benar-benar
// dikirim. Field bertipe pointer supaya "tidak dikirim" (nil) dapat dibedakan
// dari "dikirim bernilai default" (misalnya Grade=0 atau IsActive=false).
func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)
		if nim == "" {
			return failValidation(c, map[string]string{"nim": "tidak boleh kosong"})
		}
		if findStudentByNIM(nim, id) {
			return fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
		}
		students[i].NIM = nim
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return failValidation(c, map[string]string{"name": "tidak boleh kosong"})
		}
		students[i].Name = name
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			return failValidation(c, map[string]string{"grade": "harus bernilai antara 0 dan 100"})
		}
		students[i].Grade = *req.Grade
	}
	if req.IsActive != nil {
		students[i].IsActive = *req.IsActive
	}

	return ok(c, "student berhasil diperbarui sebagian", students[i])
}

// deleteStudent menghapus mahasiswa berdasarkan ID. Status 204 tanpa body.
func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	students = append(students[:i], students[i+1:]...)

	return noContent(c)
}