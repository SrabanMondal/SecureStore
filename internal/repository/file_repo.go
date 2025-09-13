package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SrabanMondal/SecureStore/internal/utils"
	"github.com/SrabanMondal/SecureStore/internal/models"
)

type FileRepository struct {
	DB *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{DB: db}
}

func (r *FileRepository) CreateFile(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (user_id, file_path, size, is_encrypted, storage_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.DB.QueryRow(ctx, query,
		file.UserID, file.FilePath, file.Size, file.IsEncrypted, file.StorageKey,
	).Scan(&file.ID, &file.CreatedAt)
	if err != nil {
		utils.Error.Err(err).Str("file_path", file.FilePath).Msg("failed to insert file")
		return err
	}
	return nil
}

func (r *FileRepository) GetFileByID(ctx context.Context, id string) (*models.File, error) {
	query := `SELECT id, user_id, file_path, size, is_encrypted, storage_key, created_at FROM files WHERE id=$1`
	var f models.File
	err := r.DB.QueryRow(ctx, query, id).
		Scan(&f.ID, &f.UserID, &f.FilePath, &f.Size, &f.IsEncrypted, &f.StorageKey, &f.CreatedAt)
	if err != nil {
		utils.Error.Err(err).Str("id", id).Msg("file not found")
		return nil, err
	}
	return &f, nil
}

func (r *FileRepository) GetFileByPath(ctx context.Context, userID, path string) (*models.File, error) {
	query := `SELECT id, user_id, file_path, size, is_encrypted, storage_key, created_at FROM files WHERE user_id=$1 AND file_path=$2`
	var f models.File
	err := r.DB.QueryRow(ctx, query, userID, path).
		Scan(&f.ID, &f.UserID, &f.FilePath, &f.Size, &f.IsEncrypted, &f.StorageKey, &f.CreatedAt)
	if err != nil {
		utils.Error.Err(err).Str("path", path).Msg("file not found for user")
		return nil, err
	}
	return &f, nil
}

func (r *FileRepository) ListFilesByUser(ctx context.Context, userID string) ([]models.File, error) {
	query := `SELECT id, user_id, file_path, size, is_encrypted, storage_key, created_at FROM files WHERE user_id=$1`
	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		utils.Error.Err(err).Str("user_id", userID).Msg("failed to list files")
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.UserID, &f.FilePath, &f.Size, &f.IsEncrypted, &f.StorageKey, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (r *FileRepository) DeleteFile(ctx context.Context, id string) error {
	query := `DELETE FROM files WHERE id=$1`
	_, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		utils.Error.Err(err).Str("id", id).Msg("failed to delete file")
		return err
	}
	return nil
}
