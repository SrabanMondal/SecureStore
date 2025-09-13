package repositories

import (
	"context"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SrabanMondal/SecureStore/internal/utils"
	"github.com/SrabanMondal/SecureStore/internal/models"
)

type ShareRepository struct {
	DB *pgxpool.Pool
}

func NewShareRepository(db *pgxpool.Pool) *ShareRepository {
	return &ShareRepository{DB: db}
}

func (r *ShareRepository) CreateShareLink(ctx context.Context, link *models.ShareLink) error {
	query := `
		INSERT INTO share_links (file_id, share_token, expires_at, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := r.DB.QueryRow(ctx, query,
		link.FileID, link.ShareToken, link.ExpiresAt, link.PasswordHash,
	).Scan(&link.ID, &link.CreatedAt)
	if err != nil {
		utils.Error.Err(err).Msg("failed to create share link")
		return err
	}
	return nil
}

func (r *ShareRepository) GetShareLinkByToken(ctx context.Context, token string) (*models.ShareLink, error) {
	query := `SELECT id, file_id, share_token, expires_at, password_hash, created_at FROM share_links WHERE share_token=$1`
	var s models.ShareLink
	err := r.DB.QueryRow(ctx, query, token).
		Scan(&s.ID, &s.FileID, &s.ShareToken, &s.ExpiresAt, &s.PasswordHash, &s.CreatedAt)
	if err != nil {
		utils.Error.Err(err).Str("token", token).Msg("share link not found")
		return nil, err
	}
	return &s, nil
}

func (r *ShareRepository) DeleteExpiredShareLinks(ctx context.Context) error {
	query := `DELETE FROM share_links WHERE expires_at < $1`
	_, err := r.DB.Exec(ctx, query, time.Now())
	if err != nil {
		utils.Error.Err(err).Msg("failed to delete expired share links")
		return err
	}
	return nil
}

func (r *ShareRepository) DeleteShareLink(ctx context.Context, id string) error {
	query := `DELETE FROM share_links WHERE id=$1`
	_, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		utils.Error.Err(err).Str("id", id).Msg("failed to delete share link")
		return err
	}
	return nil
}
