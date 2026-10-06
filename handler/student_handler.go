package handler

import (
	"errors"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"uts-044/helper"
	"uts-044/model"
	"uts-044/repository"
)

type StudentHandler struct {
	studentRepo    *repository.StudentRepository
	enrollmentRepo *repository.EnrollmentRepository
}

func NewStudentHandler(sr *repository.StudentRepository, er *repository.EnrollmentRepository) *StudentHandler {
	return &StudentHandler{studentRepo: sr, enrollmentRepo: er}
}

func (h *StudentHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	if page < 1 { page = 1 }
	if perPage < 1 { perPage = 10 }
	if perPage > 50 { perPage = 50 }

	prodi := c.Query("prodi")
	angkatan, _ := strconv.Atoi(c.Query("angkatan"))
	search := c.Query("search")
	sort := c.Query("sort")

	students, total, err := h.studentRepo.FindAll(c.Context(), page, perPage, prodi, angkatan, search, sort)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data mahasiswa")
	}

	if students == nil {
		students = []model.Student{}
	}

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 { lastPage = 1 }

	return helper.SuccessWithMeta(c, fiber.StatusOK, "Data mahasiswa berhasil diambil", students, model.PaginationMeta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	})
}

func (h *StudentHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("user_id").(int)

	student, err := h.studentRepo.FindByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data")
	}

	if role == "mahasiswa" && student.UserID != userID {
		return helper.Fail(c, fiber.StatusForbidden, "Tidak dapat mengakses data mahasiswa lain")
	}

	courses, _ := h.enrollmentRepo.FindByStudentID(c.Context(), student.ID)
	if courses == nil { courses = []model.EnrolledCourse{} }

	totalSKS := 0
	for _, ec := range courses {
		totalSKS += ec.SKS
	}

	return helper.Success(c, fiber.StatusOK, "Detail mahasiswa", model.StudentDetail{
		Student:    student,
		MataKuliah: courses,
		TotalSKS:   totalSKS,
		BatasSKS:   helper.BatasSKS(student.IPKTerakhir),
	})
}

func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	hashedPw, err := bcrypt.GenerateFromPassword([]byte(req.NIM), 10)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memproses password")
	}

	ctx := c.Context()
	tx, err := h.studentRepo.Pool().Begin(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memulai transaksi")
	}
	defer tx.Rollback(ctx)

	userID, err := h.studentRepo.CreateUser(ctx, tx, req.Email, string(hashedPw))
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.FailValidation(c, map[string]string{"email": "Email sudah terdaftar"})
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal membuat akun user")
	}

	ipk := 0.0
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	student, err := h.studentRepo.Create(ctx, tx, model.Student{
		UserID:      userID,
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: ipk,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.FailValidation(c, map[string]string{"nim": "NIM sudah terdaftar"})
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal membuat data mahasiswa")
	}

	if err := tx.Commit(ctx); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menyimpan transaksi")
	}

	return helper.Success(c, fiber.StatusCreated, "Mahasiswa berhasil ditambahkan", student)
}

func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	existing, err := h.studentRepo.FindByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	ipk := existing.IPKTerakhir
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	result, err := h.studentRepo.Update(c.Context(), model.Student{
		ID:          id,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: ipk,
	})
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memperbarui data")
	}

	return helper.Success(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", result)
}

func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	if err := h.studentRepo.SoftDelete(c.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menghapus data")
	}

	return helper.NoContent(c)
}
