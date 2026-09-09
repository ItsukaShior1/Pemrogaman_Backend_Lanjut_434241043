// Package helper berisi utilitas amplop respons dan parser query string.
// Ia dipakai oleh handler service untuk menulis respons konsisten dan
// membaca query string tanpa duplikasi logika di tiap endpoint.
package helper

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-student/app/model"
)

// Ok menulis respons 200 dengan data tunggal.
func Ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// OkList menulis respons 200 dengan data daftar dan metadata paginasi.
func OkList(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created menulis respons 201 lengkap dengan header Location.
func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent menulis respons 204 tanpa body.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Fail menulis respons gagal dengan status tertentu.
func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: false,
		Message: message,
	})
}

// FailValidation menulis respons 422 berisi map error per field.
func FailValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(model.WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errs,
	})
}

// helperFail / OkList / dll diimpor oleh service lewat nama "helperFail"
// untuk konsistensi; kita sediakan alias singkat di package helper.

var (
	allowedSort = map[string]bool{
		"id":         true,
		"nim":        true,
		"name":       true,
		"grade":      true,
		"created_at": true,
	}

	allowedOrder = map[string]bool{
		"asc":  true,
		"desc": true,
	}
)

// ParseListQuery membaca parameter query string GET /students dan
// mengembalikan ListQuery yang sudah dinormalisasi (default value +
// clamp + whitelist).
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
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
