package service

import (
	"context"
	"errors"
	"fmt"
	"go-banking-service/internal/dto"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/model"
	"go-banking-service/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
}

type userService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

func NewUserService(userRepo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *userService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("Failed to check email uniqueness during registration", "error", err)
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if existingUser != nil {
		logger.Warn("Registration attempt with existing email", "email", req.Email)
		return nil, errors.New("email already exists")
	}

	existingUserByUsername, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		logger.Error("Failed to check username uniqueness during registration", "error", err)
		return nil, fmt.Errorf("failed to check username uniqueness: %w", err)
	}
	if existingUserByUsername != nil {
		logger.Warn("Registration attempt with existing username", "username", req.Username)
		return nil, errors.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password during registration", "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		ID:           generateUUID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			logger.Warn("Registration failed due to database unique constraint", "email", req.Email, "username", req.Username)
			return nil, errors.New("email or username already exists")
		}
		logger.Error("Failed to create user in database", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.generateJWT(user.ID)
	if err != nil {
		logger.Error("Failed to generate JWT after registration", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	logger.Info("User registered successfully", "user_id", user.ID, "email", user.Email)

	return &dto.AuthResponse{
		Token:  token,
		UserID: user.ID,
	}, nil
}

func (s *userService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("Failed to get user by email during login", "error", err, "email", req.Email)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	if user == nil {
		logger.Warn("Login attempt with non-existent email", "email", req.Email)
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		logger.Warn("Login attempt with invalid password", "email", req.Email, "user_id", user.ID)
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateJWT(user.ID)
	if err != nil {
		logger.Error("Failed to generate JWT after login", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	logger.Info("User logged in successfully", "user_id", user.ID)

	return &dto.AuthResponse{
		Token:  token,
		UserID: user.ID,
	}, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *userService) generateJWT(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
