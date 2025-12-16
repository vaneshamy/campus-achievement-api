package model

import "time"

type AchievementReference struct {
    ID               string     `json:"id"`
    StudentID        string     `json:"studentId"`
    MongoAchievementID string   `json:"mongoAchievementId"`
    Status           string     `json:"status"`
    SubmittedAt      *time.Time `json:"submittedAt"`
    VerifiedAt       *time.Time `json:"verifiedAt"`
    VerifiedBy       *string    `json:"verifiedBy"`
    RejectionNote    *string    `json:"rejectionNote"`
    CreatedAt        time.Time  `json:"createdAt"`
    UpdatedAt        time.Time  `json:"updatedAt"`
}

type CreateAchievementRequest struct {
	AchievementType string                 `json:"achievementType"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`
	Tags            []string               `json:"tags"`
	Points int `json:"points"` 
}

type UpdateAchievementRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	Tags        []string               `json:"tags"`
}

type RejectAchievementRequest struct {
	Note string `json:"note"`
}

type FilterAchievementRequest struct {
	StudentID string `query:"studentId"`
}