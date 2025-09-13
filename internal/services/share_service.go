package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/SrabanMondal/SecureStore/internal/models"
	"github.com/SrabanMondal/SecureStore/internal/repository"
	//"github.com/SrabanMondal/SecureStore/internal/utils"
)

type ShareService struct {
	ShareRepo *repositories.ShareRepository
	FileRepo  *repositories.FileRepository
	FileSvc   *FileService
}

func NewShareService(shareRepo *repositories.ShareRepository, fileRepo *repositories.FileRepository, fileSvc *FileService) *ShareService {
	return &ShareService{
		ShareRepo: shareRepo,
		FileRepo:  fileRepo,
		FileSvc:   fileSvc,
	}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *ShareService) CreateShareLink(ctx context.Context, fileID string, expiry time.Duration, password string) (*models.ShareLink, error) {
	file, err := s.FileRepo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	var passwordHash string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		passwordHash = string(hash)
	}

	share := &models.ShareLink{
		FileID:       file.ID,
		ShareToken:   token,
		ExpiresAt:    time.Now().Add(expiry),
		PasswordHash: passwordHash,
	}

	if err := s.ShareRepo.CreateShareLink(ctx, share); err != nil {
		return nil, err
	}

	return share, nil
}

func (s *ShareService) ValidateShareLink(ctx context.Context, token, password string) (*models.File, error) {
	share, err := s.ShareRepo.GetShareLinkByToken(ctx, token)
	if err != nil {
		return nil, errors.New("invalid share link")
	}

	if time.Now().After(share.ExpiresAt) {
		return nil, errors.New("share link expired")
	}

	// If password required
	if share.PasswordHash != "" {
		if password == "" {
			return nil, errors.New("password_required")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(share.PasswordHash), []byte(password)); err != nil {
			return nil, errors.New("invalid password")
		}
	}

	file, err := s.FileRepo.GetFileByID(ctx, share.FileID)
	if err != nil {
		return nil, errors.New("file not found")
	}

	return file, nil
}

func (s *ShareService) GetDownloadContent(ctx context.Context, file *models.File) (interface{}, error) {
	if file.IsEncrypted {
		data, err := s.FileSvc.DownloadDecrypt(ctx, file)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	url, err := s.FileSvc.GetDownloadURL(ctx, file)
	if err != nil {
		return nil, err
	}
	return url, nil 
}

func (s *ShareService) CleanupExpiredShares(ctx context.Context) error {
	return s.ShareRepo.DeleteExpiredShareLinks(ctx)
}
