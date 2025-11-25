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

// ================================
//  CREATE
// ================================
func (r *UserRepository) Create(u *model.User) error {

	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,true)
	`

	_, err := r.db.Exec(
		query,
		u.ID,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.FullName,
		u.RoleID,
	)

	return err
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

