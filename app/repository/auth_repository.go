package repository

import (
	"database/sql"
	"time"

	"go-fiber/app/model"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// SaveRefreshToken menyimpan refresh token ke database
func (r *TokenRepository) SaveRefreshToken(token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)

	return err
}

// FindRefreshToken mencari refresh token
func (r *TokenRepository) FindRefreshToken(token string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1 AND expires_at > $2
	`

	refreshToken := &model.RefreshToken{}
	err := r.db.QueryRow(query, token, time.Now()).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return refreshToken, nil
}

// DeleteRefreshToken menghapus refresh token (untuk logout)
func (r *TokenRepository) DeleteRefreshToken(token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.Exec(query, token)
	return err
}

// DeleteUserRefreshTokens menghapus semua refresh token user
func (r *TokenRepository) DeleteUserRefreshTokens(userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

// CleanExpiredTokens membersihkan token yang expired
func (r *TokenRepository) CleanExpiredTokens() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`
	_, err := r.db.Exec(query, time.Now())
	return err
}


