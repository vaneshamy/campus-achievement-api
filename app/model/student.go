package model

import "time"

type Student struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	StudentID     string    `json:"studentId"`
	ProgramStudy  string    `json:"programStudy"`
	AcademicYear  string    `json:"academicYear"`
	AdvisorID     string    `json:"advisorId"`
	CreatedAt     time.Time `json:"createdAt"`
}

type UpdateAdvisorRequest struct {
	AdvisorID string `json:"advisorId"`
}
