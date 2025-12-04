package service

import (
	"fmt"

	"go-fiber/app/model"
	"go-fiber/app/repository"
	"go-fiber/helper"

	// "github.com/google/uuid"
)

type UserService struct {
	userRepo     *repository.UserRepository
	studentRepo  *repository.StudentRepository
	lecturerRepo *repository.LecturerRepository
}


func NewUserService(
	userRepo *repository.UserRepository,
	studentRepo *repository.StudentRepository,
	lecturerRepo *repository.LecturerRepository,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

func (s *UserService) CreateUser(req *model.User) error {
	return s.userRepo.Create(req)
}

func (s *UserService) GetAllUsers() ([]model.User, error) {
	return s.userRepo.FindAll()
}

func (s *UserService) GetUserByID(id string) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) UpdateUser(id string, req *model.User) error {

	if req.PasswordHash != "" {
		hashed, err := helper.HashPassword(req.PasswordHash)
		if err != nil {
			return fmt.Errorf("failed to hash password")
		}
		req.PasswordHash = hashed
	}

	return s.userRepo.Update(id, req)
}

func (s *UserService) DeleteUser(id string) error {
	return s.userRepo.Delete(id)
}

func (s *UserService) AssignRole(userID, roleID string) error {
	return s.userRepo.SetUserRole(userID, roleID)
}

