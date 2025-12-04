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
func (s *StudentService) GetAllStudentsService(c *fiber.Ctx) error {
	students, err := s.studentRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed to fetch students", err.Error()))
	}
	return c.JSON(model.SuccessResponse(students))
}

// GET /students/:id
func (s *StudentService) GetStudentDetailService(c *fiber.Ctx) error {
	id := c.Params("id")

	student, err := s.studentRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("student not found", nil))
	}

	return c.JSON(model.SuccessResponse(student))
}

// PUT /students/:id/advisor
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
