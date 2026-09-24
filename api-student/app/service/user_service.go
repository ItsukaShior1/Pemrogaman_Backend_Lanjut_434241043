package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-student/app/model"
	"api-student/app/repository"
	"api-student/helper"
	"api-student/middleware"
)

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(repo repository.UserRepository, perms *helper.PermissionSet) *UserService {
	return &UserService{repo: repo, perms: perms}
}

func (s *UserService) List(c *fiber.Ctx) error {
	rows, err := s.repo.ListAll(c.UserContext())
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar user")
	}
	out := make([]model.UserResponse, 0, len(rows))
	for _, u := range rows {
		out = append(out, model.UserResponse{
			ID: u.ID, Username: u.Username, Email: u.Email,
			Role: u.Role, CreatedAt: u.CreatedAt,
		})
	}
	return helper.Ok(c, "daftar user", out)
}

func (s *UserService) Get(c *fiber.Ctx) error {
	caller, _ := middleware.GetUser(c)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus angka positif")
	}
	if !CanAccessUser(caller, id, s.perms, "user:read:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengakses data user ini")
	}
	u, err := s.repo.FindByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil user")
	}
	return helper.Ok(c, "user ditemukan", model.UserResponse{
		ID: u.ID, Username: u.Username, Email: u.Email,
		Role: u.Role, CreatedAt: u.CreatedAt,
	})
}

func (s *UserService) Create(c *fiber.Ctx) error {
	return helper.Fail(c, fiber.StatusNotImplemented, "belum diimplementasi")
}

func (s *UserService) Replace(c *fiber.Ctx) error {
	return helper.Fail(c, fiber.StatusNotImplemented, "belum diimplementasi")
}

func (s *UserService) Patch(c *fiber.Ctx) error {
	return helper.Fail(c, fiber.StatusNotImplemented, "belum diimplementasi")
}

func (s *UserService) Delete(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus angka positif")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus user")
	}
	return helper.NoContent(c)
}

func (s *UserService) AssignRole(c *fiber.Ctx) error {
	return helper.Fail(c, fiber.StatusNotImplemented, "belum diimplementasi")
}

var _ = strings.TrimSpace
