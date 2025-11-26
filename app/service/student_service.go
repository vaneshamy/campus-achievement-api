package service

import (
	"go-fiber/app/model"
	"go-fiber/app/repository"
)

type StudentService struct {
	studentRepo  *repository.StudentRepository
}

func NewStudentService(studentRepo *repository.StudentRepository) *StudentService {
	return &StudentService{studentRepo}
}

func (s *StudentService) GetAllStudents() ([]model.Student, error) {
	return s.studentRepo.FindAll()
}

func (s *StudentService) GetStudentByID(id string) (*model.Student, error) {
	return s.studentRepo.FindByID(id)
}

func (s *StudentService) UpdateAdvisor(studentID, advisorID string) error {
	return s.studentRepo.UpdateAdvisor(studentID, advisorID)
}
