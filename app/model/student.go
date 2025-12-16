package model

import "time"

// ===== ENTITY (OUTPUT / DB RESULT) =====
type Student struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	StudentID    string    `json:"studentId"`
	ProgramStudy string    `json:"programStudy"`
	AcademicYear string    `json:"academicYear"`
	AdvisorID    *string   `json:"advisorId"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ===== REQUEST STRUCTS =====

// (internal) dipakai saat create user mahasiswa
type CreateStudentRequest struct {
	UserID       string  `json:"userId"`
	StudentID    string  `json:"studentId"`
	ProgramStudy string  `json:"programStudy"`
	AcademicYear string  `json:"academicYear"`
	AdvisorID    *string `json:"advisorId"`
}

// PUT /students/:id/advisor
type UpdateAdvisorRequest struct {
	AdvisorID string `json:"advisorId" validate:"required"`
}
