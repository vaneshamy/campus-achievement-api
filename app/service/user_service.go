package service

import (
	"fmt"

	"go-fiber/app/model"
	"go-fiber/app/repository"
	"go-fiber/helper"

	"github.com/google/uuid"
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

    // Hash password
    hashed, err := helper.HashPassword(req.PasswordHash)
    if err != nil {
        return fmt.Errorf("failed to hash password")
    }
    req.PasswordHash = hashed

    // Ambil roleName berdasarkan roleId
    roleName, err := s.userRepo.GetRoleNameByID(req.RoleID)
    if err != nil {
        return fmt.Errorf("role not found")
    }
    req.RoleName = roleName

    // Simpan user
    err = s.userRepo.Create(req)
    if err != nil {
        return err
    }

    // AUTO CREATE STUDENT
    if req.RoleName == "Mahasiswa" {

        err = s.studentRepo.CreateStudent(&model.Student{
            ID:           uuid.New().String(),
            UserID:       req.ID,
            StudentID:    *req.StudentID,
            ProgramStudy: *req.ProgramStudy,
            AcademicYear: *req.AcademicYear,
        })
        if err != nil {
            return fmt.Errorf("failed to create student profile: %v", err)
        }
    }

    // AUTO CREATE LECTURER
    if req.RoleName == "Dosen Wali" {

        err = s.lecturerRepo.CreateLecturer(&model.Lecturer{
            ID:         uuid.New().String(),
            UserID:     req.ID,
            LecturerID: *req.LecturerID,
            Department: *req.Department,
        })
        if err != nil {
            return fmt.Errorf("failed to create lecturer profile: %v", err)
        }
    }

    return nil
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

