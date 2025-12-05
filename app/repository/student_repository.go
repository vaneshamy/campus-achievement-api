package repository

import (
	"database/sql"
	"go-fiber/app/model"
)

type StudentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db}
}

func (r *StudentRepository) CreateStudent(s *model.Student) error {

    var advisor interface{}
    if s.AdvisorID == nil || *s.AdvisorID == "" {
        advisor = nil 
    } else {
        advisor = s.AdvisorID
    }

    query := `
        INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
    _, err := r.db.Exec(
        query,
        s.ID,
        s.UserID,
        s.StudentID,
        s.ProgramStudy,
        s.AcademicYear,
        advisor,
    )
    return err
}

// GET ALL STUDENTS
func (r *StudentRepository) FindAll() ([]model.Student, error) {
	rows, err := r.db.Query(`SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at FROM students`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := []model.Student{}
	for rows.Next() {
		var s model.Student
		rows.Scan(&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt)
		students = append(students, s)
	}
	return students, nil
}

// GET STUDENT BY ID
func (r *StudentRepository) FindByID(id string) (*model.Student, error) {
    var s model.Student

    err := r.db.QueryRow(`
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
        FROM students WHERE id = $1
    `, id).Scan(
        &s.ID,
        &s.UserID,
        &s.StudentID,
        &s.ProgramStudy,
        &s.AcademicYear,
        &s.AdvisorID,   // <-- penting!
        &s.CreatedAt,
    )

    if err != nil {
        return nil, err
    }

    return &s, nil
}


// UPDATE ADVISOR
func (r *StudentRepository) UpdateAdvisor(studentID, advisorID string) error {
	_, err := r.db.Exec(`
        UPDATE students SET advisor_id = $1 WHERE id = $2
    `, advisorID, studentID)
	return err
}

// GET STUDENT BY USER ID
func (r *StudentRepository) FindByUserID(userID string) (*model.Student, error) {
    var s model.Student

    err := r.db.QueryRow(`
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
        FROM students WHERE user_id = $1
    `, userID).Scan(
        &s.ID,
        &s.UserID,
        &s.StudentID,
        &s.ProgramStudy,
        &s.AcademicYear,
        &s.AdvisorID,
        &s.CreatedAt,
    )

    if err != nil {
        return nil, err
    }

    return &s, nil
}
