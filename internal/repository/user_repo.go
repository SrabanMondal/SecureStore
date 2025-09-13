package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"github.com/SrabanMondal/SecureStore/internal/utils"
	"github.com/SrabanMondal/SecureStore/internal/models"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.DB.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		utils.Error.Err(err).Msg("failed to create user")
		return err
	}
	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email=$1`
	var u models.User
	err := r.DB.QueryRow(ctx, query, email).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		utils.Error.Err(err).Str("email", email).Msg("user not found")
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE username=$1`
	var u models.User
	err := r.DB.QueryRow(ctx, query, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		utils.Error.Err(err).Str("username", username).Msg("user not found")
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id=$1`
	var u models.User
	err := r.DB.QueryRow(ctx, query, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		utils.Error.Err(err).Str("id", id).Msg("user not found")
		return nil, err
	}
	return &u, nil
}
