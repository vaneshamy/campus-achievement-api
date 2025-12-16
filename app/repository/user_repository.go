package repository

import (
    "database/sql"
    "fmt"
    "go-fiber/helper"

	"github.com/google/uuid"
    "go-fiber/app/model"
	"strings"
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

func (r *UserRepository) Create(req *model.CreateUserRequest) error {
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

	// Ambil role
	roleName, err := r.GetRoleNameByID(req.RoleID)
	if err != nil {
		return fmt.Errorf("invalid role")
	}

	// Validasi role
	if roleName == "Mahasiswa" {
		if req.StudentID == nil || req.ProgramStudy == nil || req.AcademicYear == nil {
			return fmt.Errorf("studentId, programStudy, academicYear wajib untuk Mahasiswa")
		}
	}

	if roleName == "Dosen Wali" {
		if req.LecturerID == nil || req.Department == nil {
			return fmt.Errorf("lecturerId dan department wajib untuk Dosen Wali")
		}
	}

	// Hash password
	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return err
	}

	userID := uuid.New().String()

	// Insert users
	_, err = tx.Exec(`
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,true)
	`,
		userID,
		req.Username,
		req.Email,
		hashed,
		req.FullName,
		req.RoleID,
	)
	if err != nil {
		return err
	}

	// Insert mahasiswa
	if roleName == "Mahasiswa" {
		_, err = tx.Exec(`
			INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id)
			VALUES ($1,$2,$3,$4,$5,$6)
		`,
			uuid.New().String(),
			userID,
			*req.StudentID,
			*req.ProgramStudy,
			*req.AcademicYear,
			req.AdvisorID,
		)
		if err != nil {
			return err
		}
	}

	// Insert dosen wali
	if roleName == "Dosen Wali" {
		_, err = tx.Exec(`
			INSERT INTO lecturers (id, user_id, lecturer_id, department)
			VALUES ($1,$2,$3,$4)
		`,
			uuid.New().String(),
			userID,
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

func (r *UserRepository) UpdatePartial(id string, req *model.UpdateUserRequest) error {
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

    // ===== users table =====
    sets := []string{}
    args := []interface{}{}
    idx := 1

    if req.Username != nil {
        sets = append(sets, fmt.Sprintf("username=$%d", idx))
        args = append(args, *req.Username)
        idx++
    }
    if req.Email != nil {
        sets = append(sets, fmt.Sprintf("email=$%d", idx))
        args = append(args, *req.Email)
        idx++
    }
    if req.FullName != nil {
        sets = append(sets, fmt.Sprintf("full_name=$%d", idx))
        args = append(args, *req.FullName)
        idx++
    }
    if req.Password != nil {
        hashed, _ := helper.HashPassword(*req.Password)
        sets = append(sets, fmt.Sprintf("password_hash=$%d", idx))
        args = append(args, hashed)
        idx++
    }
    if req.RoleID != nil {
        sets = append(sets, fmt.Sprintf("role_id=$%d", idx))
        args = append(args, *req.RoleID)
        idx++
    }

    if len(sets) > 0 {
        query := fmt.Sprintf(
            "UPDATE users SET %s WHERE id=$%d",
            strings.Join(sets, ", "),
            idx,
        )
        args = append(args, id)
        if _, err = tx.Exec(query, args...); err != nil {
            return err
        }
    }

    // ===== ambil role aktif =====
    var roleName string
    err = tx.QueryRow(`
        SELECT r.name
        FROM users u JOIN roles r ON r.id=u.role_id
        WHERE u.id=$1
    `, id).Scan(&roleName)
    if err != nil {
        return err
    }

    // ===== students =====
    if roleName == "Mahasiswa" {
        sets := []string{}
        args := []interface{}{}
        idx := 1

        if req.StudentID != nil {
            sets = append(sets, fmt.Sprintf("student_id=$%d", idx))
            args = append(args, *req.StudentID)
            idx++
        }
        if req.ProgramStudy != nil {
            sets = append(sets, fmt.Sprintf("program_study=$%d", idx))
            args = append(args, *req.ProgramStudy)
            idx++
        }
        if req.AcademicYear != nil {
            sets = append(sets, fmt.Sprintf("academic_year=$%d", idx))
            args = append(args, *req.AcademicYear)
            idx++
        }
        if req.AdvisorID != nil {
            sets = append(sets, fmt.Sprintf("advisor_id=$%d", idx))
            args = append(args, *req.AdvisorID)
            idx++
        }

        if len(sets) > 0 {
            query := fmt.Sprintf(
                "UPDATE students SET %s WHERE user_id=$%d",
                strings.Join(sets, ", "),
                idx,
            )
            args = append(args, id)
            _, err = tx.Exec(query, args...)
            if err != nil {
                return err
            }
        }
    }

    // ===== lecturers =====
    if roleName == "Dosen Wali" {
        sets := []string{}
        args := []interface{}{}
        idx := 1

        if req.LecturerID != nil {
            sets = append(sets, fmt.Sprintf("lecturer_id=$%d", idx))
            args = append(args, *req.LecturerID)
            idx++
        }
        if req.Department != nil {
            sets = append(sets, fmt.Sprintf("department=$%d", idx))
            args = append(args, *req.Department)
            idx++
        }

        if len(sets) > 0 {
            query := fmt.Sprintf(
                "UPDATE lecturers SET %s WHERE user_id=$%d",
                strings.Join(sets, ", "),
                idx,
            )
            args = append(args, id)
            _, err = tx.Exec(query, args...)
            if err != nil {
                return err
            }
        }
    }

    return nil
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

func (r *UserRepository) AssignRole(userID string, req *model.AssignRoleRequest) error {
	_, err := r.db.Exec(`
		UPDATE users SET role_id=$1 WHERE id=$2
	`, req.RoleID, userID)
	return err
}

