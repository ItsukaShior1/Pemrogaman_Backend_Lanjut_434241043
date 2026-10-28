package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"api-student/app/repository"
	"api-student/helper"
)

// CourseService membaca data kelas dari PostgreSQL. Ulasannya sendiri
// disimpan di MongoDB dan ditangani ReviewService.
type CourseService struct {
	courses repository.CourseRepository
}

func NewCourseService(courses repository.CourseRepository) *CourseService {
	return &CourseService{courses: courses}
}

func (s *CourseService) List(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	courses, err := s.courses.FindAll(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar kelas")
	}
	return helper.Ok(c, "daftar kelas berhasil diambil", courses)
}

func (s *CourseService) Get(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	course, err := s.courses.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "kelas tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data kelas")
	}
	return helper.Ok(c, "kelas ditemukan", course)
}
