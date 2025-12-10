package service

import (
	"go-fiber/app/model"
	"go-fiber/app/repository"
	"go-fiber/helper"

	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
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

func (s *UserService) GetAllUsers(c *fiber.Ctx) error {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed to fetch users", err.Error()))
	}
	return c.JSON(model.SuccessResponse(users))
}

func (s *UserService) GetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("user not found", nil))
	}
	return c.JSON(model.SuccessResponse(user))
}

func (s *UserService) CreateUser(c *fiber.Ctx) error {
	var req model.User

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
	}

	req.ID = uuid.New().String()

	if req.PasswordHash != "" {
		hashed, err := helper.HashPassword(req.PasswordHash)
		if err != nil {
			return c.Status(500).JSON(model.ErrorResponse("failed to hash password", err.Error()))
		}
		req.PasswordHash = hashed
	}

	if err := s.userRepo.Create(&req); err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed to create user", err.Error()))
	}

	return c.JSON(model.SuccessResponse(fiber.Map{
		"message": "User created successfully",
		"id":      req.ID,
	}))
}

func (s *UserService) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var req model.User

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
	}

	if req.PasswordHash != "" {
		hashed, err := helper.HashPassword(req.PasswordHash)
		if err != nil {
			return c.Status(500).JSON(model.ErrorResponse("failed to hash password", err.Error()))
		}
		req.PasswordHash = hashed
	}

	if err := s.userRepo.Update(id, &req); err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed to update user", err.Error()))
	}

	return c.JSON(model.SuccessResponse("User updated successfully"))
}

func (s *UserService) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.userRepo.Delete(id); err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed to delete user", err.Error()))
	}

	return c.JSON(model.SuccessResponse("User deleted successfully"))
}

func (s *UserService) AssignRole(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		RoleID string `json:"roleId"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
	}

	if body.RoleID == "" {
		return c.Status(400).JSON(model.ErrorResponse("roleId is required", nil))
	}

	if err := s.userRepo.SetUserRole(id, body.RoleID); err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed to update role", err.Error()))
	}

	return c.JSON(model.SuccessResponse("Role updated successfully"))
}
