package repository

import (
    "database/sql"
    "fmt"
    "go-fiber/helper"

	"github.com/google/uuid"
    "go-fiber/app/model"
)

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Cari user berdasarkan username atau email
func (r *UserRepository) FindByUsernameOrEmail(identifier string) (*model.User, error) {
    query := `
        SELECT u.id, u.username, u.email, u.password_hash, u.full_name,
               u.role_id, r.name as role_name, u.is_active
        FROM users u
        JOIN roles r ON r.id = u.role_id
        WHERE u.username = $1 OR u.email = $1
    `

    user := &model.User{}
    err := r.db.QueryRow(query, identifier).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
        &user.PasswordHash,
        &user.FullName,
        &user.RoleID,
        &user.RoleName,
        &user.IsActive,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    return user, err
}

// Cari user by ID
func (r *UserRepository) FindByID(id string) (*model.User, error) {
    query := `
        SELECT u.id, u.username, u.email, u.password_hash, u.full_name,
               u.role_id, r.name as role_name, u.is_active
        FROM users u
        JOIN roles r ON r.id = u.role_id
        WHERE u.id = $1
    `

    user := &model.User{}
    err := r.db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
        &user.PasswordHash,
        &user.FullName,
        &user.RoleID,
        &user.RoleName,
        &user.IsActive,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    return user, err
}

// Ambil semua permission dari role
func (r *UserRepository) GetUserPermissions(roleID string) ([]string, error) {
    query := `
        SELECT p.name
        FROM permissions p
        JOIN role_permissions rp ON p.id = rp.permission_id
        WHERE rp.role_id = $1
    `

    rows, err := r.db.Query(query, roleID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var perms []string
    for rows.Next() {
        var perm string
        rows.Scan(&perm)
        perms = append(perms, perm)
    }

    return perms, nil
}

func (r *UserRepository) Create(req *model.User) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Ambil nama role
	roleName, err := r.GetRoleNameByID(req.RoleID)
	if err != nil {
		return fmt.Errorf("invalid role")
	}
	req.RoleName = roleName

	// Validasi role Mahasiswa
	if roleName == "Mahasiswa" {
		if req.StudentID == nil || req.ProgramStudy == nil || req.AcademicYear == nil {
			return fmt.Errorf("studentId, programStudy and academicYear are required for Mahasiswa")
		}
	}

	// Validasi role Dosen Wali
	if roleName == "Dosen Wali" {
		if req.LecturerID == nil || req.Department == nil {
			return fmt.Errorf("lecturerId and department are required for Dosen Wali")
		}
	}

	// Hash password
	hashed, err := helper.HashPassword(req.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to hash password")
	}
	req.PasswordHash = hashed

	// Insert ke users
	_, err = tx.Exec(`
        INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active)
        VALUES ($1,$2,$3,$4,$5,$6,true)
    `,
		req.ID,
		req.Username,
		req.Email,
		req.PasswordHash,
		req.FullName,
		req.RoleID,
	)
	if err != nil {
		return err
	}

	// Insert student jika role mahasiswa
	if roleName == "Mahasiswa" {
		_, err = tx.Exec(`
            INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id)
			VALUES ($1,$2,$3,$4,$5,$6)
		`,
			uuid.New().String(),
			req.ID,
			*req.StudentID,
			*req.ProgramStudy,
			*req.AcademicYear,
			*req.AdvisorID,
		)
		if err != nil {
			return err
		}
	}

	// Insert lecturer jika role dosen wali
	if roleName == "Dosen Wali" {
		_, err = tx.Exec(`
            INSERT INTO lecturers (id, user_id, lecturer_id, department)
			VALUES ($1,$2,$3,$4)
        `,
			uuid.New().String(),
			req.ID,
			*req.LecturerID,
			*req.Department,
		)
		if err != nil {
			return err
		}
	}

	return nil
}


func (r *UserRepository) FindAll() ([]model.User, error) {

	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name,
			   u.role_id, r.name as role_name, u.is_active
		FROM users u
		JOIN roles r ON r.id = u.role_id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User

	for rows.Next() {
		var u model.User
		rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.PasswordHash,
			&u.FullName,
			&u.RoleID,
			&u.RoleName,
			&u.IsActive,
		)
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) Update(id string, u *model.User) error {

	query := `
		UPDATE users
		SET username=$1, email=$2, password_hash=$3, full_name=$4, role_id=$5
		WHERE id=$6
	`

	_, err := r.db.Exec(
		query,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.FullName,
		u.RoleID,
		id,
	)

	return err
}

func (r *UserRepository) Delete(id string) error {

	_, err := r.db.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}

func (r *UserRepository) SetUserRole(userID, roleID string) error {
	_, err := r.db.Exec(`UPDATE users SET role_id=$1 WHERE id=$2`, roleID, userID)
	return err
}

func (r *UserRepository) GetRoleNameByID(roleID string) (string, error) {
    var name string
    err := r.db.QueryRow(`
        SELECT name FROM roles WHERE id = $1
    `, roleID).Scan(&name)

    if err == sql.ErrNoRows {
        return "", fmt.Errorf("role not found")
    }

    return name, err
}
