package repository

import (
	"fmt"
	"database/sql"
	"go-fiber/app/model"
)

type LecturerRepository struct {
	db *sql.DB
}

func NewLecturerRepository(db *sql.DB) *LecturerRepository {
	return &LecturerRepository{db}
}

func (r *LecturerRepository) CreateLecturer(l *model.Lecturer) error {
    query := `
        INSERT INTO lecturers (id, user_id, lecturer_id, department, created_at)
        VALUES ($1, $2, $3, $4, NOW())
    `
    _, err := r.db.Exec(query, l.ID, l.UserID, l.LecturerID, l.Department)
    if err != nil {
        return fmt.Errorf("failed to create lecturer: %v", err)
    }
    return nil
}

func (r *LecturerRepository) FindAll() ([]model.Lecturer, error) {
	rows, err := r.db.Query(`SELECT id, user_id, lecturer_id, department, created_at FROM lecturers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lecturers := []model.Lecturer{}
	for rows.Next() {
		var l model.Lecturer
		rows.Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt)
		lecturers = append(lecturers, l)
	}
	return lecturers, nil
}

func (r *LecturerRepository) FindAdvisees(lecturerID string) ([]model.Student, error) {
	rows, err := r.db.Query(`
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at 
        FROM students WHERE advisor_id = $1
    `, lecturerID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Student
	for rows.Next() {
		var s model.Student
		rows.Scan(&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt)
		list = append(list, s)
	}

	return list, nil
}
