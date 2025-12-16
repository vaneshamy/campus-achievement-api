package repository_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"go-fiber/app/repository"
)

// Test FindByUsernameOrEmail - Success with Username
func TestFindByUsernameOrEmail_SuccessWithUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "full_name",
		"role_id", "role_name", "is_active",
	}).AddRow(
		"user-123", "testuser", "test@example.com", "hashedpassword",
		"Test User", "role-1", "Admin", true,
	)

	mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.password_hash, u.full_name`).
		WithArgs("testuser").
		WillReturnRows(rows)

	user, err := repo.FindByUsernameOrEmail("testuser")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Admin", user.RoleName)
	assert.True(t, user.IsActive)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test FindByUsernameOrEmail - Success with Email
func TestFindByUsernameOrEmail_SuccessWithEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "full_name",
		"role_id", "role_name", "is_active",
	}).AddRow(
		"user-123", "testuser", "test@example.com", "hashedpassword",
		"Test User", "role-1", "Admin", true,
	)

	mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.password_hash, u.full_name`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	user, err := repo.FindByUsernameOrEmail("test@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test FindByUsernameOrEmail - User Not Found
func TestFindByUsernameOrEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.password_hash, u.full_name`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.FindByUsernameOrEmail("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test FindByUsernameOrEmail - Database Error
func TestFindByUsernameOrEmail_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.password_hash, u.full_name`).
		WithArgs("testuser").
		WillReturnError(sql.ErrConnDone)

	user, err := repo.FindByUsernameOrEmail("testuser")

	assert.Error(t, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test FindByID - Success
func TestFindByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "full_name",
		"role_id", "role_name", "is_active",
	}).AddRow(
		"user-123", "testuser", "test@example.com", "hashedpassword",
		"Test User", "role-1", "Admin", true,
	)

	mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.password_hash, u.full_name`).
		WithArgs("user-123").
		WillReturnRows(rows)

	user, err := repo.FindByID("user-123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.FullName)
	assert.Equal(t, "role-1", user.RoleID)
	assert.Equal(t, "Admin", user.RoleName)
	assert.True(t, user.IsActive)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test FindByID - User Not Found
func TestFindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.password_hash, u.full_name`).
		WithArgs("nonexistent-id").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.FindByID("nonexistent-id")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test FindByID - Database Error
func TestFindByID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.password_hash, u.full_name`).
		WithArgs("user-123").
		WillReturnError(sql.ErrConnDone)

	user, err := repo.FindByID("user-123")

	assert.Error(t, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetUserPermissions - Success
func TestGetUserPermissions_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	rows := sqlmock.NewRows([]string{"name"}).
		AddRow("read:users").
		AddRow("write:users").
		AddRow("delete:users")

	mock.ExpectQuery(`SELECT p.name FROM permissions p`).
		WithArgs("role-1").
		WillReturnRows(rows)

	permissions, err := repo.GetUserPermissions("role-1")

	assert.NoError(t, err)
	assert.NotNil(t, permissions)
	assert.Len(t, permissions, 3)
	assert.Contains(t, permissions, "read:users")
	assert.Contains(t, permissions, "write:users")
	assert.Contains(t, permissions, "delete:users")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetUserPermissions - No Permissions
func TestGetUserPermissions_NoPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	rows := sqlmock.NewRows([]string{"name"})

	mock.ExpectQuery(`SELECT p.name FROM permissions p`).
		WithArgs("role-2").
		WillReturnRows(rows)

	permissions, err := repo.GetUserPermissions("role-2")

	assert.NoError(t, err)
	assert.NotNil(t, permissions)
	assert.Len(t, permissions, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetUserPermissions - Database Error
func TestGetUserPermissions_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	mock.ExpectQuery(`SELECT p.name FROM permissions p`).
		WithArgs("role-1").
		WillReturnError(sql.ErrConnDone)

	permissions, err := repo.GetUserPermissions("role-1")

	assert.Error(t, err)
	assert.Nil(t, permissions)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetUserPermissions - Scan Error
func TestGetUserPermissions_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	// Create rows with wrong type to cause scan error
	rows := sqlmock.NewRows([]string{"name"}).
		AddRow(123) // Should be string, not int

	mock.ExpectQuery(`SELECT p.name FROM permissions p`).
		WithArgs("role-1").
		WillReturnRows(rows)

	permissions, err := repo.GetUserPermissions("role-1")

	assert.Error(t, err)
	assert.Nil(t, permissions)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test FindByUsernameOrEmail - Inactive User
func TestFindByUsernameOrEmail_InactiveUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "full_name",
		"role_id", "role_name", "is_active",
	}).AddRow(
		"user-123", "testuser", "test@example.com", "hashedpassword",
		"Test User", "role-1", "Admin", false, // is_active = false
	)

	mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.password_hash, u.full_name`).
		WithArgs("testuser").
		WillReturnRows(rows)

	user, err := repo.FindByUsernameOrEmail("testuser")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.False(t, user.IsActive)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test FindByID - Inactive User
func TestFindByID_InactiveUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewAuthRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password_hash", "full_name",
		"role_id", "role_name", "is_active",
	}).AddRow(
		"user-123", "testuser", "test@example.com", "hashedpassword",
		"Test User", "role-1", "User", false, // is_active = false
	)

	mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.password_hash, u.full_name`).
		WithArgs("user-123").
		WillReturnRows(rows)

	user, err := repo.FindByID("user-123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.False(t, user.IsActive)

	assert.NoError(t, mock.ExpectationsWereMet())
}