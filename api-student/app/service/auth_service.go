package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"api-student/app/model"
	"api-student/app/repository"
	"api-student/helper"
)

const (
	BcryptCost       = 10
	AccessTTLSeconds  = 15 * 60
	RefreshTTLSeconds = 7 * 24 * 60 * 60
)

type AuthService struct {
	users       repository.UserRepository
	refresh     repository.RefreshTokenRepository
	tokens      *helper.TokenIssuer
	rateLimiter *LoginRateLimiter
}

func NewAuthService(
	users repository.UserRepository,
	refresh repository.RefreshTokenRepository,
	tokens *helper.TokenIssuer,
	rateLimiter *LoginRateLimiter,
) *AuthService {
	return &AuthService{
		users:       users,
		refresh:     refresh,
		tokens:      tokens,
		rateLimiter: rateLimiter,
	}
}

func toUserResponse(u model.User) model.UserResponse {
	return model.UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

func (s *AuthService) issueTokens(ctx context.Context, u model.User) (model.AuthResponse, error) {
	access, exp, err := s.tokens.Issue(u.ID, u.Username, string(u.Role))
	if err != nil {
		return model.AuthResponse{}, err
	}

	raw, hashed, err := helper.GenerateRefreshToken()
	if err != nil {
		return model.AuthResponse{}, err
	}

	rt, err := s.refresh.Create(ctx, model.RefreshToken{
		UserID:    u.ID,
		TokenHash: hashed,
		ExpiresAt: time.Now().Add(time.Duration(RefreshTTLSeconds) * time.Second),
	})
	if err != nil {
		return model.AuthResponse{}, err
	}

	_ = rt

	return model.AuthResponse{
		User:         toUserResponse(u),
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int(time.Until(exp).Seconds()),
	}, nil
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	errs := map[string]string{}
	for k, v := range ValidateUsername(req.Username) {
		errs["username_"+k] = v
	}
	for k, v := range ValidateEmail(req.Email) {
		errs["email_"+k] = v
	}
	for k, v := range ValidatePassword(req.Password, DefaultPasswordPolicy) {
		errs["password_"+k] = v
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), BcryptCost)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
	}

	u, err := s.users.Create(ctx, model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         model.RoleUser,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict, "username atau email sudah dipakai")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan user")
	}

	resp, err := s.issueTokens(ctx, u)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menerbitkan token")
	}

	return helper.Created(c, "registrasi berhasil", resp, "/api/v1/auth/me")
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)

	if req.Username == "" || req.Password == "" {
		return helper.Fail(c, fiber.StatusUnauthorized, "username atau password salah")
	}

	key := c.IP() + "|" + req.Username

	allowed, retryAfter := s.rateLimiter.Allow(key)
	if !allowed {
		c.Set("Retry-After", intToStr(retryAfter))
		return helper.Fail(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan login, coba lagi nanti")
	}

	u, err := s.users.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			_, ra := s.rateLimiter.RegisterFailure(key)
			if ra > 0 {
				c.Set("Retry-After", intToStr(ra))
				return helper.Fail(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan login, coba lagi nanti")
			}
			return helper.Fail(c, fiber.StatusUnauthorized, "username atau password salah")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		_, ra := s.rateLimiter.RegisterFailure(key)
		if ra > 0 {
			c.Set("Retry-After", intToStr(ra))
			return helper.Fail(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan login, coba lagi nanti")
		}
		return helper.Fail(c, fiber.StatusUnauthorized, "username atau password salah")
	}

	s.rateLimiter.RegisterSuccess(key)

	resp, err := s.issueTokens(ctx, u)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menerbitkan token")
	}

	return helper.Ok(c, "login berhasil", resp)
}

func (s *AuthService) Refresh(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.RefreshToken == "" {
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid")
	}

	hashed := helper.HashRefreshToken(req.RefreshToken)

	rt, err := s.refresh.FindByHash(ctx, hashed)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memvalidasi refresh token")
	}

	if rt.RevokedAt != nil {
		_ = s.refresh.RevokeAllForUser(ctx, rt.UserID)
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak aktif")
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = s.refresh.Revoke(ctx, rt.ID, nil)
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak aktif")
	}

	u, err := s.users.FindByID(ctx, rt.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak aktif")
	}

	raw, newHash, err := helper.GenerateRefreshToken()
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menerbitkan token")
	}

	newRT, err := s.refresh.Create(ctx, model.RefreshToken{
		UserID:    u.ID,
		TokenHash: newHash,
		ExpiresAt: time.Now().Add(time.Duration(RefreshTTLSeconds) * time.Second),
	})
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menerbitkan token")
	}

	if err := s.refresh.Revoke(ctx, rt.ID, &newRT.ID); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memutar refresh token")
	}

	access, exp, err := s.tokens.Issue(u.ID, u.Username, string(u.Role))
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menerbitkan token")
	}

	return helper.Ok(c, "token berhasil diperbarui", model.AuthResponse{
		User:         toUserResponse(u),
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int(time.Until(exp).Seconds()),
	})
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	var req model.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.RefreshToken == "" {
		return helper.Ok(c, "logout berhasil", nil)
	}

	hashed := helper.HashRefreshToken(req.RefreshToken)
	rt, err := s.refresh.FindByHash(ctx, hashed)
	if err == nil && rt.RevokedAt == nil {
		_ = s.refresh.Revoke(ctx, rt.ID, nil)
	}

	return helper.Ok(c, "logout berhasil", nil)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	uidVal := c.Locals("user_id")
	uid, ok := uidVal.(int)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "token tidak valid")
	}

	u, err := s.users.FindByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusUnauthorized, "user tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil user")
	}

	return helper.Ok(c, "profil user", toUserResponse(u))
}
