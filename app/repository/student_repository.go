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

func (r *StudentRepository) CreateStudent(req *model.CreateStudentRequest) error {
	var advisor interface{}
	if req.AdvisorID == nil || *req.AdvisorID == "" {
		advisor = nil
	} else {
		advisor = *req.AdvisorID
	}

	_, err := r.db.Exec(`
		INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
	`,
		req.UserID,
		req.StudentID,
		req.ProgramStudy,
		req.AcademicYear,
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
func (r *StudentRepository) UpdateAdvisor(studentID string, advisorID string) error {
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

func (r *StudentRepository) FindByAdvisorID(advisorID string) ([]model.Student, error) {
    rows, err := r.db.Query(`
        SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
        FROM students WHERE advisor_id = $1
    `, advisorID)
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
