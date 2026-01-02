package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
	"github.com/noteduco342/OMMessenger-backend/internal/validation"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo         repository.UserRepositoryInterface
	refreshTokenRepo repository.RefreshTokenRepositoryInterface
}

func NewAuthService(userRepo repository.UserRepositoryInterface, refreshTokenRepo repository.RefreshTokenRepositoryInterface) *AuthService {
	return &AuthService{userRepo: userRepo, refreshTokenRepo: refreshTokenRepo}
}

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthSession struct {
	AccessToken  string
	RefreshToken string
	User         models.UserResponse
}

func (s *AuthService) Register(input RegisterInput) (*AuthSession, error) {
	// Check if user exists
	input.Email = validation.NormalizeEmail(input.Email)
	input.Username = validation.NormalizeUsername(input.Username)
	input.FullName = validation.TrimAndLimit(input.FullName, 80)

	if _, err := s.userRepo.FindByEmail(input.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	if _, err := s.userRepo.FindByUsername(input.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		FullName:     input.FullName,
		Role:         "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	session, err := s.issueSession(user)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthService) Login(input LoginInput) (*AuthSession, error) {
	input.Email = validation.NormalizeEmail(input.Email)
	// Find user
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.issueSession(user)
}

func (s *AuthService) RefreshSession(refreshToken string) (*AuthSession, error) {
	if refreshToken == "" {
		return nil, errors.New("missing refresh token")
	}

	refreshHash := hashToken(refreshToken)
	stored, err := s.refreshTokenRepo.FindValidByHash(refreshHash)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.userRepo.FindByID(stored.UserID)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Rotate refresh token
	if err := s.refreshTokenRepo.RevokeByHash(refreshHash); err != nil {
		return nil, err
	}

	return s.issueSession(user)
}

func (s *AuthService) Logout(refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.refreshTokenRepo.RevokeByHash(hashToken(refreshToken))
}

func (s *AuthService) issueSession(user *models.User) (*AuthSession, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshHash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	refresh := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.refreshTokenRepo.Create(refresh); err != nil {
		return nil, err
	}

	return &AuthSession{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

func (s *AuthService) generateAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func generateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	hash = hashToken(raw)
	return raw, hash, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
