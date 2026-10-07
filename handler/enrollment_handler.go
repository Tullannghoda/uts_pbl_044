package handler

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"uts-044/helper"
	"uts-044/model"
	"uts-044/repository"
	"strconv"
)

type EnrollmentHandler struct {
	enrollmentRepo *repository.EnrollmentRepository
	studentRepo    *repository.StudentRepository
}

func NewEnrollmentHandler(er *repository.EnrollmentRepository, sr *repository.StudentRepository) *EnrollmentHandler {
	return &EnrollmentHandler{enrollmentRepo: er, studentRepo: sr}
}

func (h *EnrollmentHandler) Create(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(int)

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	student, err := h.studentRepo.FindByUserID(c.Context(), userID)
	if err != nil {
		return helper.Fail(c, fiber.StatusForbidden, "Data mahasiswa tidak ditemukan")
	}

	ctx := c.Context()
	tx, err := h.enrollmentRepo.Pool().Begin(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memulai transaksi")
	}
	defer tx.Rollback(ctx)

	kuota, terisi, sksMK, err := h.enrollmentRepo.CourseQuotaWithLock(ctx, tx, req.CourseID)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Mata kuliah tidak ditemukan")
	}

	if terisi >= kuota {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Kuota mata kuliah sudah penuh")
	}

	currentSKS, err := h.enrollmentRepo.TotalSKS(ctx, student.ID, req.TahunAkademik)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menghitung SKS")
	}

	batas := helper.BatasSKS(student.IPKTerakhir)
	if currentSKS+sksMK > batas {
		sisa := batas - currentSKS
		return helper.Fail(c, fiber.StatusUnprocessableEntity,
			fmt.Sprintf("Total SKS melebihi batas (%d). Sisa SKS yang dapat diambil: %d", batas, sisa))
	}

	enrollment, err := h.enrollmentRepo.Create(ctx, tx, model.Enrollment{
		StudentID:     student.ID,
		CourseID:      req.CourseID,
		TahunAkademik: req.TahunAkademik,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict, "Mata kuliah sudah pernah diambil pada tahun akademik ini")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menambahkan KRS")
	}

	if err := tx.Commit(ctx); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menyimpan transaksi")
	}

	return helper.Success(c, fiber.StatusCreated, "Mata kuliah berhasil diambil", enrollment)
}

func (h *EnrollmentHandler) Delete(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(int)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	enrollment, err := h.enrollmentRepo.FindByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Enrollment tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data")
	}

	student, err := h.studentRepo.FindByUserID(c.Context(), userID)
	if err != nil || enrollment.StudentID != student.ID {
		return helper.Fail(c, fiber.StatusForbidden, "Enrollment ini bukan milik Anda")
	}

	if err := h.enrollmentRepo.Delete(c.Context(), id); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menghapus enrollment")
	}

	return helper.NoContent(c)
}
