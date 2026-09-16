package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-student/app/model"
	"api-student/app/repository"
	"api-student/helper"
)

type prestasiService struct {
	repo repository.PrestasiRepository
}

func NewPrestasiService(repo repository.PrestasiRepository) *prestasiService {
	return &prestasiService{repo: repo}
}

func translatePrestasiError(c *fiber.Ctx, err error, pesanUmum string) error {
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "prestasi tidak ditemukan")
	}
	return helper.Fail(c, fiber.StatusInternalServerError, pesanUmum)
}

func (s *prestasiService) List(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	q := helper.ParseListQuery(c)
	prestasis, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data prestasi")
	}

	totalPages := 0
	if q.Limit > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}
	return helper.OkList(c, "daftar prestasi berhasil diambil", prestasis, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (s *prestasiService) Get(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	prestasi, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translatePrestasiError(c, err, "gagal mengambil data prestasi")
	}

	return helper.Ok(c, "data prestasi berhasil diambil", prestasi)
}
func (s *prestasiService) Create(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	var req model.CreatePrestasiRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "request body tidak valid")
	}

	req.Title = strings.TrimSpace(req.Title)

	prestasi, err := s.repo.Create(ctx, req)
	if err != nil {
		return translatePrestasiError(c, err, "gagal membuat data prestasi")
	}

	return helper.Ok(c, "data prestasi berhasil dibuat", prestasi)
}
