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

// GetLecturers godoc
// @Summary Get all lecturers
// @Description Mengambil seluruh data dosen
// @Tags Lecturers
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.APIResponse{data=[]model.Lecturer}
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /lecturers [get]
func (s *LecturerService) GetLecturers(c *fiber.Ctx) error {
	data, err := s.lecturerRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("Failed get lecturers", err.Error()))
	}
	return c.JSON(model.SuccessResponse(data))
}

// GetAdvisees godoc
// @Summary Get lecturer advisees
// @Description Mengambil daftar mahasiswa bimbingan dosen
// @Tags Lecturers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Lecturer ID"
// @Success 200 {object} model.APIResponse{data=[]model.Student}
// @Failure 404 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /lecturers/{id}/advisees [get]
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	lecturerID := c.Params("id")

	data, err := s.lecturerRepo.FindAdvisees(lecturerID)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("Lecturer not found", nil))
	}
	return c.JSON(model.SuccessResponse(data))
}
