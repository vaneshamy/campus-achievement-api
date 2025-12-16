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

func (r *LecturerRepository) CreateLecturer(req *model.CreateLecturerRequest) error {
	_, err := r.db.Exec(`
		INSERT INTO lecturers (id, user_id, lecturer_id, department, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
	`,
		req.UserID,
		req.LecturerID,
		req.Department,
	)

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

func (r *LecturerRepository) FindByID(id string) (*model.Lecturer, error) {
    var lec model.Lecturer

    err := r.db.QueryRow(`
        SELECT id, user_id, lecturer_id, department, created_at
        FROM lecturers
        WHERE id = $1
    `, id).Scan(
        &lec.ID,
        &lec.UserID,
        &lec.LecturerID,
        &lec.Department,
        &lec.CreatedAt,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, err
        }
        return nil, err
    }

    return &lec, nil
}
