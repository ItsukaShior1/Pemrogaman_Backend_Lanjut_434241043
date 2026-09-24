package service

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-student/app/model"
	"api-student/app/repository"
	"api-student/helper"
	"api-student/middleware"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

func reqCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	return id, err == nil && id > 0
}

func translateError(c *fiber.Ctx, err error, pesanUmum string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, "NIM sudah dipakai")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, pesanUmum)
	}
}

func logStudent(c *fiber.Ctx, status int, pesan string, details string) {
	caller, _ := middleware.GetUser(c)
	log.Printf(
		"user_id=%d role=%s method=%s path=%s status=%d msg=%s %s",
		caller.UserID, caller.Role, c.Method(), c.Path(), status, pesan, details,
	)
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	q := helper.ParseListQuery(c)

	students, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data student")
	}

	totalPages := 0
	if q.Limit > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	logStudent(c, fiber.StatusOK, "list",
		"page="+strconv.Itoa(q.Page)+" limit="+strconv.Itoa(q.Limit)+" total="+strconv.Itoa(total))

	return helper.OkList(c, "daftar student berhasil diambil", students, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	caller, _ := middleware.GetUser(c)

	ctx, cancel := reqCtx(c)
	defer cancel()

	id, ok := paramID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data student")
	}

	if !CanAccessStudent(caller, student.OwnerID, s.perms, "student:read:any") {
		logStudent(c, fiber.StatusForbidden, "get-ditolak",
			"id="+strconv.Itoa(id)+" owner="+strconv.Itoa(student.OwnerID))
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak berhak melihat student ini")
	}

	logStudent(c, fiber.StatusOK, "get", "id="+strconv.Itoa(id))
	return helper.Ok(c, "student ditemukan", student)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	caller, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	ctx, cancel := reqCtx(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := ValidateCreate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	baru, err := s.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
		OwnerID:  caller.UserID,
	})
	if err != nil {
		return translateError(c, err, "gagal menyimpan student")
	}

	logStudent(c, fiber.StatusCreated, "create",
		"id="+strconv.Itoa(baru.ID)+" owner="+strconv.Itoa(baru.OwnerID))

	return helper.Created(c, "student berhasil dibuat", baru, "/api/v1/students/"+strconv.Itoa(baru.ID))
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	caller, _ := middleware.GetUser(c)

	ctx, cancel := reqCtx(c)
	defer cancel()

	id, ok := paramID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	saatIni, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data student")
	}

	if !CanAccessStudent(caller, saatIni.OwnerID, s.perms, "student:update:any") {
		logStudent(c, fiber.StatusForbidden, "replace-ditolak",
			"id="+strconv.Itoa(id)+" owner="+strconv.Itoa(saatIni.OwnerID))
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak boleh mengubah student ini")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := ValidateReplace(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	hasil, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
		OwnerID:  saatIni.OwnerID,
	})
	if err != nil {
		return translateError(c, err, "gagal memperbarui student")
	}

	logStudent(c, fiber.StatusOK, "replace", "id="+strconv.Itoa(id))
	return helper.Ok(c, "student berhasil diganti seluruhnya", hasil)
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	caller, _ := middleware.GetUser(c)

	ctx, cancel := reqCtx(c)
	defer cancel()

	id, ok := paramID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	saatIni, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data student")
	}

	if !CanAccessStudent(caller, saatIni.OwnerID, s.perms, "student:update:any") {
		logStudent(c, fiber.StatusForbidden, "patch-ditolak",
			"id="+strconv.Itoa(id)+" owner="+strconv.Itoa(saatIni.OwnerID))
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak boleh mengubah student ini")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	mutasi, errs := ApplyPatch(saatIni, req)
	if len(errs) > 0 {
		if errs["_"] == "tidak ada field yang diubah" {
			return helper.Fail(c, fiber.StatusBadRequest, errs["_"])
		}
		return helper.FailValidation(c, errs)
	}

	mutasi.OwnerID = saatIni.OwnerID

	hasil, err := s.repo.Update(ctx, mutasi)
	if err != nil {
		return translateError(c, err, "gagal memperbarui student")
	}

	logStudent(c, fiber.StatusOK, "patch", "id="+strconv.Itoa(id))
	return helper.Ok(c, "student berhasil diperbarui sebagian", hasil)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, ok := paramID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(c, err, "gagal menghapus student")
	}

	logStudent(c, fiber.StatusNoContent, "delete", "id="+strconv.Itoa(id))
	return helper.NoContent(c)
}