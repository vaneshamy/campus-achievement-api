package service

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/model"
	"go-fiber/app/repository"
)

type LecturerService struct {
	lecturerRepo *repository.LecturerRepository
}

func NewLecturerService(lecturerRepo *repository.LecturerRepository) *LecturerService {
	return &LecturerService{lecturerRepo}
}

func (s *LecturerService) GetLecturers(c *fiber.Ctx) error {
	data, err := s.lecturerRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("Failed get lecturers", err.Error()))
	}
	return c.JSON(model.SuccessResponse(data))
}

func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	lecturerID := c.Params("id")

	data, err := s.lecturerRepo.FindAdvisees(lecturerID)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("Lecturer not found", nil))
	}
	return c.JSON(model.SuccessResponse(data))
}
