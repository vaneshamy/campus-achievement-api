package model

import "time"

type AchievementReference struct {
    ID               string     `json:"id"`
    StudentID        string     `json:"studentId"`
    MongoAchievementID string   `json:"mongoAchievementId"`
    Status           string     `json:"status"` // draft, submitted, verified, rejected
    SubmittedAt      *time.Time `json:"submittedAt"`
    VerifiedAt       *time.Time `json:"verifiedAt"`
    VerifiedBy       *string    `json:"verifiedBy"`
    RejectionNote    *string    `json:"rejectionNote"`
    CreatedAt        time.Time  `json:"createdAt"`
    UpdatedAt        time.Time  `json:"updatedAt"`
}

