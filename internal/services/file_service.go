package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"

	"github.com/SrabanMondal/SecureStore/internal/models"
	"github.com/SrabanMondal/SecureStore/internal/repository"
	"github.com/SrabanMondal/SecureStore/internal/utils"
)

type FileService struct {
	FileRepo *repositories.FileRepository
	Minio    *minio.Client
	Bucket   string
	FileKey  []byte
}

func NewFileService(repo *repositories.FileRepository, minio *minio.Client, bucket string, fileKey []byte) *FileService {
	return &FileService{
		FileRepo: repo,
		Minio:    minio,
		Bucket:   bucket,
		FileKey:  fileKey,
	}
}


func (s *FileService) GeneratePresignedUpload(ctx context.Context, userID, filePath string, size int64) (string, *models.File, error) {
	storageKey := fmt.Sprintf("%s/%s", userID, filePath)

	url, err := s.Minio.PresignedPutObject(ctx, s.Bucket, storageKey, 15*time.Minute)
	if err != nil {
		return "", nil, err
	}

	file := &models.File{
		UserID:      userID,
		FilePath:    filePath,
		Size:        size,
		IsEncrypted: false,
		StorageKey:  storageKey,
	}

	if err := s.FileRepo.CreateFile(ctx, file); err != nil {
		return "", nil, err
	}

	return url.String(), file, nil
}

func (s *FileService) MarkUploaded(ctx context.Context, fileID string) error {
	return s.FileRepo.MarkFileUploaded(ctx, fileID)
}


func (s *FileService) UploadEncrypted(ctx context.Context, userID, filePath string, file io.Reader, size int64) error {
	storageKey := fmt.Sprintf("%s/%s", userID, filePath)

	dbFile := &models.File{
		UserID:      userID,
		FilePath:    filePath,
		Size:        size,
		IsEncrypted: true,
		StorageKey:  storageKey,
	}
	if err := s.FileRepo.CreateFile(ctx, dbFile); err != nil {
		return err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		_ = s.FileRepo.DeleteFile(ctx, dbFile.ID)
		return err
	}

	cipherData, nonce, err := utils.Encrypt(data, s.FileKey)
	if err != nil {
		_ = s.FileRepo.DeleteFile(ctx, dbFile.ID)
		return err
	}
	finalData := append(nonce, cipherData...)

	_, err = s.Minio.PutObject(ctx, s.Bucket, storageKey, bytes.NewReader(finalData), int64(len(finalData)), minio.PutObjectOptions{})
	if err != nil {
		_ = s.FileRepo.DeleteFile(ctx, dbFile.ID)
		return err
	}

	return s.FileRepo.MarkFileUploaded(ctx, dbFile.ID)
}


func (s *FileService) GetDownloadURL(ctx context.Context, file *models.File) (string, error) {
	url, err := s.Minio.PresignedGetObject(ctx, s.Bucket, file.StorageKey, 15*time.Minute, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (s *FileService) DownloadDecrypt(ctx context.Context, file *models.File) ([]byte, error) {
	obj, err := s.Minio.GetObject(ctx, s.Bucket, file.StorageKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, err
	}

	nonceSize := 12
	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	return utils.Decrypt(cipherData, nonce, s.FileKey)
}

func (s *FileService) DeleteFile(ctx context.Context, fileID string) error {
	file, err := s.FileRepo.GetFileByID(ctx, fileID)
	if err != nil {
		return err
	}

	if err := s.Minio.RemoveObject(ctx, s.Bucket, file.StorageKey, minio.RemoveObjectOptions{}); err != nil {
		return err
	}

	return s.FileRepo.DeleteFile(ctx, fileID)
}
