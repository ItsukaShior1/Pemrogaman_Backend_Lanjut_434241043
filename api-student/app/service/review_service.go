package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-student/app/model"
	"api-student/app/repository"
	"api-student/helper"
	"api-student/middleware"
)

// ReviewService memakai DUA database sekaligus: data kelas dan KRS dibaca
// dari PostgreSQL, ulasan disimpan di MongoDB. Service-lah satu-satunya
// tempat yang mengenal keduanya.
type ReviewService struct {
	reviews repository.ReviewRepository
	courses repository.CourseRepository
	perms   *helper.PermissionSet
}

func NewReviewService(
	reviews repository.ReviewRepository,
	courses repository.CourseRepository,
	perms *helper.PermissionSet,
) *ReviewService {
	return &ReviewService{reviews: reviews, courses: courses, perms: perms}
}

// ---------- POST /courses/:id/reviews ----------

func (s *ReviewService) Create(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	current, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	courseID, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id kelas harus berupa angka positif")
	}

	var req model.CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.Comment = strings.TrimSpace(req.Comment)
	req.Tags = NormalizeTags(req.Tags)

	errs := map[string]string{}
	if msg, ok := validateRating(req.Rating); !ok {
		errs["rating"] = msg
	}
	if msg, ok := validateComment(req.Comment); !ok {
		errs["comment"] = msg
	}
	if msg, ok := validateTags(req.Tags); !ok {
		errs["tags"] = msg
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// Tidak ada FOREIGN KEY yang menghubungkan MongoDB dengan PostgreSQL.
	// Pemeriksaan "kelasnya ada" dan "mahasiswa ini peserta kelasnya" harus
	// dilakukan sendiri oleh service, dengan bertanya ke PostgreSQL.
	if _, err := s.courses.FindByID(ctx, courseID); err != nil {
		status, pesan := translateReviewError(err)
		return helper.Fail(c, status, pesan)
	}

	enrolled, err := s.courses.IsEnrolled(ctx, courseID, current.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa pendaftaran")
	}
	if !enrolled {
		return helper.Fail(c, fiber.StatusForbidden, "hanya peserta kelas yang boleh menulis ulasan")
	}

	created, err := s.reviews.Create(ctx, model.Review{
		CourseID: courseID, UserID: current.UserID, Username: current.Username,
		Rating: req.Rating, Comment: req.Comment, Tags: req.Tags,
	})
	if err != nil {
		status, pesan := translateReviewError(err)
		return helper.Fail(c, status, pesan)
	}
	return helper.Created(c, "ulasan berhasil disimpan", created, "/api/v1/reviews/"+created.ID)
}

// ---------- GET /courses/:id/reviews ----------

func (s *ReviewService) ListByCourse(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	courseID, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id kelas harus berupa angka positif")
	}

	q := model.ReviewQuery{
		CourseID: courseID,
		Tag:      strings.ToLower(strings.TrimSpace(c.Query("tag"))),
		Before:   c.Query("before"),
		Limit:    c.QueryInt("limit", 10),
	}
	if q.Limit < 1 || q.Limit > 50 {
		q.Limit = 10
	}

	// Minta satu lebih banyak untuk mengetahui apakah masih ada halaman
	// berikutnya.
	want := q.Limit
	q.Limit = want + 1

	reviews, err := s.reviews.FindByCourse(ctx, q)
	if err != nil {
		status, pesan := translateReviewError(err)
		return helper.Fail(c, status, pesan)
	}

	hasMore := len(reviews) > want
	if hasMore {
		reviews = reviews[:want]
	}
	next := ""
	if hasMore {
		next = reviews[len(reviews)-1].ID
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "daftar ulasan berhasil diambil",
		"data":    reviews,
		"meta": fiber.Map{
			"limit": want, "has_more": hasMore, "next_before": next,
		},
	})
}

// ---------- GET /courses/:id/reviews/summary ----------

func (s *ReviewService) Summary(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	courseID, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id kelas harus berupa angka positif")
	}
	if _, err := s.courses.FindByID(ctx, courseID); err != nil {
		status, pesan := translateReviewError(err)
		return helper.Fail(c, status, pesan)
	}

	summary, err := s.reviews.Summary(ctx, courseID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal meringkas ulasan")
	}
	return helper.Ok(c, "ringkasan ulasan berhasil diambil", summary)
}

// ---------- GET /reviews/:id ----------

func (s *ReviewService) Get(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	review, err := s.reviews.FindByID(ctx, c.Params("id"))
	if err != nil {
		status, pesan := translateReviewError(err)
		return helper.Fail(c, status, pesan)
	}
	return helper.Ok(c, "ulasan ditemukan", review)
}

// ---------- PATCH /reviews/:id ----------

func (s *ReviewService) Update(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	current, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	var req model.UpdateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if req.Rating == nil && req.Comment == nil && req.Tags == nil {
		return helper.Fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}
	if req.Comment != nil {
		trimmed := strings.TrimSpace(*req.Comment)
		req.Comment = &trimmed
	}
	if req.Tags != nil {
		normalized := NormalizeTags(*req.Tags)
		req.Tags = &normalized
	}

	errs := map[string]string{}
	if req.Rating != nil {
		if msg, ok := validateRating(*req.Rating); !ok {
			errs["rating"] = msg
		}
	}
	if req.Comment != nil {
		if msg, ok := validateComment(*req.Comment); !ok {
			errs["comment"] = msg
		}
	}
	if req.Tags != nil {
		if msg, ok := validateTags(*req.Tags); !ok {
			errs["tags"] = msg
		}
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	updated, err := s.reviews.UpdateOwn(ctx, c.Params("id"), current.UserID, req)
	if err != nil {
		status, pesan := translateReviewError(err)
		return helper.Fail(c, status, pesan)
	}
	return helper.Ok(c, "ulasan berhasil diubah", updated)
}

// ---------- DELETE /reviews/:id ----------
// Pemilik boleh menghapus ulasannya sendiri; pemegang permission
// review:moderate boleh menghapus ulasan siapa pun.

func (s *ReviewService) Delete(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	current, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	anyOwner := s.perms.Can(current.Role, "review:moderate")
	if err := s.reviews.Delete(ctx, c.Params("id"), current.UserID, anyOwner); err != nil {
		status, pesan := translateReviewError(err)
		return helper.Fail(c, status, pesan)
	}
	return helper.NoContent(c)
}

// ---------- POST /reviews/:id/replies ----------

func (s *ReviewService) Reply(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	current, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	var req model.ReplyRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.Body = strings.TrimSpace(req.Body)
	if msg, ok := validateReplyBody(req.Body); !ok {
		return helper.FailValidation(c, map[string]string{"body": msg})
	}

	updated, err := s.reviews.AddReply(ctx, c.Params("id"), model.Reply{
		UserID: current.UserID, Username: current.Username, Body: req.Body,
	})
	if err != nil {
		status, pesan := translateReviewError(err)
		return helper.Fail(c, status, pesan)
	}
	return helper.Ok(c, "balasan berhasil ditambahkan", updated)
}

// ---------- POST /reviews/:id/helpful ----------

func (s *ReviewService) MarkHelpful(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	current, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	updated, err := s.reviews.MarkHelpful(ctx, c.Params("id"), current.UserID)
	if err != nil {
		status, pesan := translateReviewError(err)
		return helper.Fail(c, status, pesan)
	}
	return helper.Ok(c, "terima kasih atas penilaiannya", updated)
}

func translateReviewError(err error) (int, string) {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return fiber.StatusBadRequest, "id ulasan tidak valid"
	case errors.Is(err, repository.ErrNotFound):
		return fiber.StatusNotFound, "data tidak ditemukan"
	case errors.Is(err, repository.ErrDuplicate):
		return fiber.StatusConflict,
			"Anda sudah menulis ulasan untuk kelas ini, ubah ulasan yang ada"
	case errors.Is(err, repository.ErrNotOwner):
		return fiber.StatusForbidden, "hanya pemilik ulasan yang boleh melakukan ini"
	case errors.Is(err, repository.ErrOwnReview):
		return fiber.StatusConflict, "tidak dapat menilai ulasan sendiri"
	case errors.Is(err, repository.ErrAlreadyVoted):
		return fiber.StatusConflict, "Anda sudah menilai ulasan ini"
	case errors.Is(err, repository.ErrReplyLimit):
		return fiber.StatusConflict, "jumlah balasan sudah mencapai batas"
	default:
		return fiber.StatusInternalServerError, "terjadi kesalahan pada server"
	}
}
