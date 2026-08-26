package main

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ok mengirim respons berhasil tunggal (200 OK).
func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// okList mengirim respons daftar (200 OK) beserta informasi paginasi.
func okList(c *fiber.Ctx, message string, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// created mengirim respons sumber daya baru dibuat (201 Created)
// dan menyertakan header Location agar klien tahu alamat resource baru.
func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// noContent mengirim 204 No Content. Dipakai untuk DELETE yang berhasil
// karena memang tidak ada body yang perlu dikembalikan.
func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// fail mengirim respons gagal umum dengan status yang dipilih pemanggil.
func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
	})
}

// failValidation mengirim respons 422 dengan rincian kesalahan per field,
// sehingga aplikasi klien dapat menandai kolom yang salah pada formulir.
func failValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errs,
	})
}

// allowedSort adalah DAFTAR PUTIH field yang boleh dipakai untuk mengurutkan.
// Daftar putih (bukan hitam) adalah kebiasaan aman karena menerima nama field
// apa pun dari klien dan menempelkannya ke query bisa berujung injeksi.
var allowedSort = map[string]bool{
	"id":         true,
	"nim":        true,
	"name":       true,
	"grade":      true,
	"created_at": true,
}

// allowedOrder membatasi arah pengurutan hanya pada asc atau desc.
var allowedOrder = map[string]bool{
	"asc":  true,
	"desc": true,
}

// parseListQuery membaca query string dari permintaan daftar dan
// memberikan nilai bawaan yang aman. Masukan klien tidak pernah dipercaya.
func parseListQuery(c *fiber.Ctx) ListQuery {
	q := ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}
	if !allowedOrder[q.Order] {
		q.Order = "asc"
	}

	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}

	if raw := c.Query("grade_min"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			q.GradeMin = &v
		}
	}
	if raw := c.Query("grade_max"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			q.GradeMax = &v
		}
	}

	return q
}