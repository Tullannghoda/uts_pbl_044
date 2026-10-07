package route

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"uts-044/handler"
	"uts-044/middleware"
	"uts-044/repository"
)

func Setup(app *fiber.App, pool *pgxpool.Pool) {
	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)

	authHandler := handler.NewAuthHandler(userRepo, studentRepo)
	studentHandler := handler.NewStudentHandler(studentRepo, enrollmentRepo)
	courseHandler := handler.NewCourseHandler(courseRepo)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentRepo, studentRepo)

	api := app.Group("/api/v1")

	api.Post("/auth/login", middleware.RateLimit(5, 1*time.Minute), authHandler.Login)
	api.Get("/auth/me", middleware.Auth(), authHandler.Me)

	students := api.Group("/students", middleware.Auth())
	students.Get("/", middleware.AdminOnly(), studentHandler.List)
	students.Post("/", middleware.AdminOnly(), studentHandler.Create)
	students.Get("/:id", studentHandler.Get)
	students.Put("/:id", middleware.AdminOnly(), studentHandler.Update)
	students.Delete("/:id", middleware.AdminOnly(), studentHandler.Delete)

	api.Get("/courses", middleware.Auth(), courseHandler.List)

	enrollments := api.Group("/enrollments", middleware.Auth(), middleware.MahasiswaOnly())
	enrollments.Post("/", enrollmentHandler.Create)
	enrollments.Delete("/:id", enrollmentHandler.Delete)
}
