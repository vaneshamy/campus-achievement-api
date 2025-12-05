package repository

import (
	"database/sql"
	"time"

	"go-fiber/app/model"
)

type AchievementRepository struct {
	db *sql.DB
}

func NewAchievementRepository(db *sql.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

func (r *AchievementRepository) CreateReference(ref *model.AchievementReference) error {
	query := `
		INSERT INTO achievement_references
		(id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`
	_, err := r.db.Exec(
		query,
		ref.ID,
		ref.StudentID,
		ref.MongoAchievementID,
		ref.Status,
		ref.SubmittedAt,
		ref.VerifiedAt,
		ref.VerifiedBy,
		ref.RejectionNote,
		ref.CreatedAt,
		ref.UpdatedAt,
	)
	return err
}

func (r *AchievementRepository) UpdateReferenceStatus(id, status string, submittedAt, verifiedAt *time.Time, verifiedBy, rejectionNote *string) error {
	query := `
		UPDATE achievement_references
		SET status=$1, submitted_at=$2, verified_at=$3, verified_by=$4, rejection_note=$5, updated_at=$6
		WHERE id=$7
	`
	_, err := r.db.Exec(query, status, submittedAt, verifiedAt, verifiedBy, rejectionNote, time.Now(), id)
	return err
}

func (r *AchievementRepository) FindReferenceByID(id string) (*model.AchievementReference, error) {
	var ref model.AchievementReference
	err := r.db.QueryRow(`
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references WHERE id=$1
	`, id).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&ref.SubmittedAt,
		&ref.VerifiedAt,
		&ref.VerifiedBy,
		&ref.RejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

func (r *AchievementRepository) FindByStudentID(studentID string) ([]model.AchievementReference, error) {
	rows, err := r.db.Query(`
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references WHERE student_id=$1
	`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.AchievementReference
	for rows.Next() {
		var ref model.AchievementReference
		rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.SubmittedAt,
			&ref.VerifiedAt,
			&ref.VerifiedBy,
			&ref.RejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		list = append(list, ref)
	}
	return list, nil
}

// For admin or lecturer: get all (simple version, add pagination/filter later)
func (r *AchievementRepository) FindAll() ([]model.AchievementReference, error) {
	rows, err := r.db.Query(`
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.AchievementReference
	for rows.Next() {
		var ref model.AchievementReference
		rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.SubmittedAt,
			&ref.VerifiedAt,
			&ref.VerifiedBy,
			&ref.RejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		list = append(list, ref)
	}
	return list, nil
}
