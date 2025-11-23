package repository

import (
    "database/sql"
    "fmt"
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
