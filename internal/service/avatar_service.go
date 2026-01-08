package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
	"github.com/noteduco342/OMMessenger-backend/internal/storage"
)

var ErrStorageNotConfigured = errors.New("storage not configured")

type AvatarService struct {
	userRepo repository.UserRepositoryInterface
	s3       *storage.S3Storage
}

func NewAvatarService(userRepo repository.UserRepositoryInterface, s3 *storage.S3Storage) *AvatarService {
	return &AvatarService{userRepo: userRepo, s3: s3}
}

// UploadAvatar processes an uploaded image and stores it as a JPEG avatar.
// Returns updated user.
func (s *AvatarService) UploadAvatar(ctx context.Context, userID uint, fileReader io.Reader, publicAPIBaseURL string) (*models.User, error) {
	if s.s3 == nil {
		return nil, ErrStorageNotConfigured
	}
	publicAPIBaseURL = strings.TrimRight(strings.TrimSpace(publicAPIBaseURL), "/")
	if publicAPIBaseURL == "" {
		return nil, errors.New("missing public api base url")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	opts := storage.DefaultAvatarOptions()
	jpegBytes, contentType, outSize, err := storage.ProcessAvatarImage(fileReader, opts)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("avatars/%d/%s.jpg", userID, uuid.NewString())
	st, err := s.s3.PutObject(ctx, key, bytes.NewReader(jpegBytes), outSize, contentType)
	if err != nil {
		return nil, err
	}

	avatarURL := publicAPIBaseURL + "/media/avatars/" + key

	// Keep old key; delete only after DB update succeeds.
	oldKey := strings.TrimSpace(user.AvatarKey)

	now := time.Now().UTC()
	user.Avatar = avatarURL
	user.AvatarKey = key
	user.AvatarContentType = contentType
	user.AvatarSizeBytes = outSize
	user.AvatarUpdatedAt = &now
	user.AvatarETag = st.ETag

	if err := s.userRepo.Update(user); err != nil {
		// Try to delete newly created object to avoid orphan.
		_ = s.s3.DeleteObject(ctx, key)
		return nil, err
	}

	// Best-effort delete previous object if present.
	if oldKey != "" && oldKey != key {
		_ = s.s3.DeleteObject(ctx, oldKey)
	}

	return user, nil
}

// DeleteAvatar removes the user's avatar reference and deletes the stored object
// (best-effort). Returns updated user.
func (s *AvatarService) DeleteAvatar(ctx context.Context, userID uint) (*models.User, error) {
	if s.s3 == nil {
		return nil, ErrStorageNotConfigured
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	oldKey := strings.TrimSpace(user.AvatarKey)

	user.Avatar = ""
	user.AvatarKey = ""
	user.AvatarContentType = ""
	user.AvatarSizeBytes = 0
	user.AvatarUpdatedAt = nil
	user.AvatarETag = ""

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Best-effort delete previous object if present.
	if oldKey != "" {
		_ = s.s3.DeleteObject(ctx, oldKey)
	}

	return user, nil
}
