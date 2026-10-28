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

// PortfolioService memakai DUA database sekaligus: identitas mahasiswa dibaca
// dari PostgreSQL, portofolionya disimpan di MongoDB.
//
// MongoDB tidak punya foreign key ke PostgreSQL, jadi pemeriksaan "mahasiswa
// ini ada" tetap dilakukan di sini, bukan di database.
type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
	perms      *helper.PermissionSet
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
	perms *helper.PermissionSet,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students, perms: perms}
}

// ---------- GET /students/:id/portfolio ----------

func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	caller, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	student, status, pesan := s.requireStudent(c, ctx)
	if status != fiber.StatusOK {
		return helper.Fail(c, status, pesan)
	}

	if !CanAccessStudent(caller, student.OwnerID, s.perms, "student:read:any") {
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak berhak melihat portofolio ini")
	}

	portfolio, err := s.portfolios.FindByStudentID(ctx, student.ID)
	if err != nil {
		status, pesan := translatePortfolioError(err)
		return helper.Fail(c, status, pesan)
	}

	logStudent(c, fiber.StatusOK, "portfolio-get", "student_id="+strconv.Itoa(student.ID))
	return helper.Ok(c, "portofolio ditemukan", portfolio)
}

// ---------- PUT /students/:id/portfolio ----------

func (s *PortfolioService) Replace(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	caller, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	student, status, pesan := s.requireStudent(c, ctx)
	if status != fiber.StatusOK {
		return helper.Fail(c, status, pesan)
	}

	if !CanAccessStudent(caller, student.OwnerID, s.perms, "student:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak boleh mengubah portofolio ini")
	}

	var req model.ReplacePortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Headline = strings.TrimSpace(req.Headline)
	req.Bio = strings.TrimSpace(req.Bio)

	projects := make([]model.Project, 0, len(req.Projects))
	for _, p := range req.Projects {
		projects = append(projects, model.Project{
			Title:        strings.TrimSpace(p.Title),
			Technologies: NormalizeTechnologies(p.Technologies),
			URL:          strings.TrimSpace(p.URL),
			Year:         p.Year,
		})
	}

	certificates := make([]model.Certificate, 0, len(req.Certificates))
	for _, cert := range req.Certificates {
		tanggal, _ := time.Parse("2006-01-02", strings.TrimSpace(cert.Date))
		certificates = append(certificates, model.Certificate{
			Name:   strings.TrimSpace(cert.Name),
			Issuer: strings.TrimSpace(cert.Issuer),
			Date:   tanggal,
		})
	}

	if errs := ValidateReplacePortfolio(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	portfolio, err := s.portfolios.Replace(ctx, model.Portfolio{
		StudentID:    student.ID,
		Headline:     req.Headline,
		Bio:          req.Bio,
		Projects:     projects,
		Certificates: certificates,
	})
	if err != nil {
		status, pesan := translatePortfolioError(err)
		return helper.Fail(c, status, pesan)
	}

	logStudent(c, fiber.StatusOK, "portfolio-replace", "student_id="+strconv.Itoa(student.ID))
	return helper.Ok(c, "portofolio berhasil disimpan", portfolio)
}

// ---------- POST /students/:id/portfolio/projects ----------

func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	caller, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	student, status, pesan := s.requireStudent(c, ctx)
	if status != fiber.StatusOK {
		return helper.Fail(c, status, pesan)
	}

	if !CanAccessStudent(caller, student.OwnerID, s.perms, "student:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak boleh mengubah portofolio ini")
	}

	var req model.AddProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Title = strings.TrimSpace(req.Title)
	req.URL = strings.TrimSpace(req.URL)
	req.Technologies = NormalizeTechnologies(req.Technologies)

	if errs := ValidateAddProject(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	portfolio, err := s.portfolios.AddProject(ctx, student.ID, model.Project{
		Title:        req.Title,
		Technologies: req.Technologies,
		URL:          req.URL,
		Year:         req.Year,
	})
	if err != nil {
		status, pesan := translatePortfolioError(err)
		return helper.Fail(c, status, pesan)
	}

	logStudent(c, fiber.StatusCreated, "portfolio-project-add", "student_id="+strconv.Itoa(student.ID))
	return helper.Created(c, "proyek berhasil ditambahkan", portfolio,
		"/api/v1/students/"+strconv.Itoa(student.ID)+"/portfolio")
}

// ---------- DELETE /students/:id/portfolio/projects/:projectId ----------

func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	caller, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	student, status, pesan := s.requireStudent(c, ctx)
	if status != fiber.StatusOK {
		return helper.Fail(c, status, pesan)
	}

	if !CanAccessStudent(caller, student.OwnerID, s.perms, "student:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak boleh mengubah portofolio ini")
	}

	if err := s.portfolios.DeleteProject(ctx, student.ID, c.Params("projectId")); err != nil {
		status, pesan := translatePortfolioError(err)
		return helper.Fail(c, status, pesan)
	}

	logStudent(c, fiber.StatusNoContent, "portfolio-project-delete",
		"student_id="+strconv.Itoa(student.ID))
	return helper.NoContent(c)
}

// ---------- GET /portfolios/stats/technologies ----------

func (s *PortfolioService) TechnologyStats(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	stats, err := s.portfolios.TechnologyStats(ctx, 10)
	if err != nil {
		status, pesan := translatePortfolioError(err)
		return helper.Fail(c, status, pesan)
	}

	return helper.Ok(c, "statistik teknologi berhasil diambil", stats)
}

// requireStudent membaca mahasiswa dari PostgreSQL. Pemanggil harus sudah
// memeriksa hak akses sebelum menggunakan hasilnya.
func (s *PortfolioService) requireStudent(
	c *fiber.Ctx, ctx context.Context,
) (model.Student, int, string) {
	id, valid := paramID(c)
	if !valid {
		return model.Student{}, fiber.StatusBadRequest, "id harus berupa angka positif"
	}

	student, err := s.students.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Student{}, fiber.StatusNotFound, "student tidak ditemukan"
		}
		return model.Student{}, fiber.StatusInternalServerError, "gagal mengambil data student"
	}
	return student, fiber.StatusOK, ""
}

func translatePortfolioError(err error) (int, string) {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return fiber.StatusBadRequest, "id proyek tidak valid"
	case errors.Is(err, repository.ErrNotFound):
		return fiber.StatusNotFound, "student tidak ditemukan"
	case errors.Is(err, repository.ErrPortfolioNotFound):
		return fiber.StatusNotFound, "portofolio belum dibuat"
	case errors.Is(err, repository.ErrProjectNotFound):
		return fiber.StatusNotFound, "proyek tidak ditemukan"
	case errors.Is(err, repository.ErrProjectLimit):
		return fiber.StatusConflict, "jumlah proyek sudah mencapai batas 20"
	case errors.Is(err, repository.ErrDuplicate):
		return fiber.StatusConflict, "portofolio sudah ada"
	default:
		return fiber.StatusInternalServerError, "terjadi kesalahan pada server"
	}
}
