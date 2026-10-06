package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"uts-044/helper"
	"uts-044/model"
	"uts-044/repository"
)

type CourseHandler struct {
	courseRepo *repository.CourseRepository
}

func NewCourseHandler(cr *repository.CourseRepository) *CourseHandler {
	return &CourseHandler{courseRepo: cr}
}

func (h *CourseHandler) List(c *fiber.Ctx) error {
	semester, _ := strconv.Atoi(c.Query("semester"))
	search := c.Query("search")
	available := c.Query("available") == "true"

	courses, err := h.courseRepo.FindAll(c.Context(), semester, search, available)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data mata kuliah")
	}

	if courses == nil {
		courses = []model.Course{}
	}

	return helper.Success(c, fiber.StatusOK, "Daftar mata kuliah berhasil diambil", courses)
}
