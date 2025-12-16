package model

import (
	"time"
)

// User model untuk tabel users
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` 
	FullName     string    `json:"fullName"`
	RoleID       string    `json:"roleId"`
	RoleName     string    `json:"role"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`

	StudentID    *string   `json:"studentId"`
	ProgramStudy *string   `json:"programStudy"`
	AcademicYear *string   `json:"academicYear"`
	AdvisorID     *string    `json:"advisorId"`

	LecturerID *string     `json:"lecturerId"`
	Department *string     `json:"department"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"fullName" validate:"required"`
	RoleID   string `json:"roleId" validate:"required"`

	// Mahasiswa
	StudentID     *string `json:"studentId,omitempty"`
	ProgramStudy *string `json:"programStudy,omitempty"`
	AcademicYear *string `json:"academicYear,omitempty"`
	AdvisorID    *string `json:"advisorId,omitempty"`

	// Dosen
	LecturerID *string `json:"lecturerId,omitempty"`
	Department *string `json:"department,omitempty"`
}

type UpdateUserRequest struct {
    Username      *string `json:"username"`
    Email         *string `json:"email"`
    Password      *string `json:"password"`
    FullName      *string `json:"full_name"`
    RoleID        *string `json:"role_id"`

    // Mahasiswa
    StudentID     *string `json:"student_id"`
    ProgramStudy *string `json:"program_study"`
    AcademicYear *string `json:"academic_year"`
    AdvisorID    *string `json:"advisor_id"`

    // Dosen
    LecturerID *string `json:"lecturer_id"`
    Department *string `json:"department"`
}

type AssignRoleRequest struct {
	RoleID string `json:"roleId" validate:"required"`
}


// Role model untuk tabel roles
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Permission model untuk tabel permissions
type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type UserResponse struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"fullName"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}
