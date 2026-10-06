package handler

import (
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"uts-044/config"
	"uts-044/helper"
	"uts-044/model"
	"uts-044/repository"
)

type AuthHandler struct {
	userRepo    *repository.UserRepository
	studentRepo *repository.StudentRepository
}

func NewAuthHandler(ur *repository.UserRepository, sr *repository.StudentRepository) *AuthHandler {
	return &AuthHandler{userRepo: ur, studentRepo: sr}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	user, err := h.userRepo.FindByEmail(c.Context(), req.Email)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "Email atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "Email atau password salah")
	}

	token, err := helper.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal membuat token")
	}

	return helper.Success(c, fiber.StatusOK, "Login berhasil", model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(config.JWTExpiration.Seconds()),
		User: model.UserBrief{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(int)
	role, _ := c.Locals("role").(string)

	user, err := h.userRepo.FindByID(c.Context(), userID)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "User tidak ditemukan")
	}

	result := fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}

	if role == "mahasiswa" {
		student, err := h.studentRepo.FindByUserID(c.Context(), userID)
		if err == nil {
			result["student"] = fiber.Map{
				"nim":      student.NIM,
				"nama":     student.Nama,
				"prodi":    student.Prodi,
				"angkatan": student.Angkatan,
			}
		}
	}

	return helper.Success(c, fiber.StatusOK, "Profil pengguna", result)
}
