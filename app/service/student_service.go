package service

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/model"
	"go-fiber/app/repository"
)

type StudentService struct {
	studentRepo *repository.StudentRepository
}

func NewStudentService(studentRepo *repository.StudentRepository) *StudentService {
	return &StudentService{studentRepo: studentRepo}
}

// GET /students

// GetAllStudents godoc
// @Summary Get all students
// @Description Mengambil seluruh data mahasiswa
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.APIResponse{data=[]model.Student}
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /students [get]
func (s *StudentService) GetAllStudentsService(c *fiber.Ctx) error {
	students, err := s.studentRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed to fetch students", err.Error()))
	}
	return c.JSON(model.SuccessResponse(students))
}

// GET /students/:id

// GetStudentDetail godoc
// @Summary Get student detail
// @Description Mengambil detail mahasiswa berdasarkan ID
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Param id path string true "Student ID"
// @Success 200 {object} model.APIResponse{data=model.Student}
// @Failure 404 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /students/{id} [get]
func (s *StudentService) GetStudentDetailService(c *fiber.Ctx) error {
	id := c.Params("id")

	student, err := s.studentRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("student not found", nil))
	}

	return c.JSON(model.SuccessResponse(student))
}

// PUT /students/:id/advisor

// UpdateAdvisor godoc
// @Summary Update student advisor
// @Description Mengubah dosen wali mahasiswa
// @Tags Students
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Student ID"
// @Param body body model.UpdateAdvisorRequest true "Update advisor request"
// @Success 200 {object} model.APIResponse
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /students/{id}/advisor [put]
func (s *StudentService) UpdateAdvisorService(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.UpdateAdvisorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
	}

	err := s.studentRepo.UpdateAdvisor(id, req.AdvisorID)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed to update advisor", err.Error()))
	}

	return c.JSON(model.SuccessResponse("advisor updated"))
}
