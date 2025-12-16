package model

import "time"

type Lecturer struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	LecturerID string    `json:"lecturerId"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"createdAt"`
}

type CreateLecturerRequest struct {
	UserID     string `json:"userId"`
	LecturerID string `json:"lecturerId"`
	Department string `json:"department"`
}