package service

import (
	"go-fiber/app/model"
	"go-fiber/app/repository"
)

type LecturerService struct {
	lecturerRepo *repository.LecturerRepository
}

func NewLecturerService(lecturerRepo *repository.LecturerRepository) *LecturerService {
	return &LecturerService{lecturerRepo}
}

func (s *LecturerService) GetLecturers() ([]model.Lecturer, error) {
	return s.lecturerRepo.FindAll()
}

func (s *LecturerService) GetAdvisees(lecturerID string) ([]model.Student, error) {
	return s.lecturerRepo.FindAdvisees(lecturerID)
}
