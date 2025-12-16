package service

import (
	"go-fiber/app/model"
	"go-fiber/app/repository"

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

// ==================== GET ALL USERS ====================

// GetAllUsers godoc
// @Summary Get all users
// @Description Mengambil semua data user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.APIResponse{data=[]model.User}
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /users [get]
func (s *UserService) GetAllUsers(c *fiber.Ctx) error {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(
			model.ErrorResponse("failed to fetch users", err.Error()),
		)
	}
	return c.JSON(model.SuccessResponse(users))
}

// ==================== GET USER BY ID ====================

// GetUserByID godoc
// @Summary Get user by ID
// @Description Mengambil detail user berdasarkan ID
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.APIResponse{data=model.User}
// @Failure 404 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /users/{id} [get]
func (s *UserService) GetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(
			model.ErrorResponse("user not found", nil),
		)
	}

	return c.JSON(model.SuccessResponse(user))
}

// ==================== CREATE USER ====================

// CreateUser godoc
// @Summary Create new user
// @Description Membuat user baru (Admin only)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body model.CreateUserRequest true "Create user request"
// @Success 201 {object} model.APIResponse
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /users [post]
func (s *UserService) CreateUser(c *fiber.Ctx) error {
	var req model.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(
			model.ErrorResponse("invalid request body", err.Error()),
		)
	}

	if err := s.userRepo.Create(&req); err != nil {
		return c.Status(500).JSON(
			model.ErrorResponse("failed to create user", err.Error()),
		)
	}

	return c.Status(201).JSON(
		model.SuccessResponse("User created successfully"),
	)
}

// ==================== UPDATE USER ====================

// UpdateUser godoc
// @Summary Update user
// @Description Update sebagian data user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body model.UpdateUserRequest true "Update user request"
// @Success 200 {object} model.APIResponse
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /users/{id} [put]
func (s *UserService) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(
			model.ErrorResponse("invalid request body", err.Error()),
		)
	}

	if err := s.userRepo.UpdatePartial(id, &req); err != nil {
		return c.Status(500).JSON(
			model.ErrorResponse("failed to update user", err.Error()),
		)
	}

	return c.JSON(model.SuccessResponse("User updated successfully"))
}

// ==================== DELETE USER ====================

// DeleteUser godoc
// @Summary Delete user
// @Description Menghapus user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /users/{id} [delete]
func (s *UserService) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.userRepo.Delete(id); err != nil {
		return c.Status(500).JSON(
			model.ErrorResponse("failed to delete user", err.Error()),
		)
	}

	return c.JSON(model.SuccessResponse("User deleted successfully"))
}

// ==================== ASSIGN ROLE ====================

// AssignRole godoc
// @Summary Assign role to user
// @Description Mengubah role user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body model.AssignRoleRequest true "Assign role request"
// @Success 200 {object} model.APIResponse
// @Failure 400 {object} model.APIResponse
// @Failure 401 {object} model.APIResponse
// @Failure 403 {object} model.APIResponse
// @Router /users/{id}/role [put]
func (s *UserService) AssignRole(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(
			model.ErrorResponse("invalid request body", err.Error()),
		)
	}

	if err := s.userRepo.AssignRole(id, &req); err != nil {
		return c.Status(500).JSON(
			model.ErrorResponse("failed to assign role", err.Error()),
		)
	}

	return c.JSON(model.SuccessResponse("Role updated successfully"))
}
