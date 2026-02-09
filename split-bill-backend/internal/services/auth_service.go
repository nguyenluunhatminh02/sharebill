package services

import (
	"context"
	"errors"
	"time"

	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// VerifyAndGetUser finds or creates a user based on Firebase UID
func (s *AuthService) VerifyAndGetUser(ctx context.Context, firebaseUID, phone string) (*models.User, error) {
	// Try to find existing user
	user, err := s.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	// User doesn't exist, create new one
	newUser := &models.User{
		FirebaseUID:  firebaseUID,
		Phone:        phone,
		DisplayName:  "User",
		BankAccounts: []models.BankAccount{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// GetUserByID gets a user by their ID
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	return s.userRepo.FindByID(ctx, objID)
}

// GetUserByFirebaseUID gets a user by their Firebase UID
func (s *AuthService) GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*models.User, error) {
	return s.userRepo.FindByFirebaseUID(ctx, firebaseUID)
}

// UpdateProfile updates a user's profile
func (s *AuthService) UpdateProfile(ctx context.Context, firebaseUID string, req models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.BankAccounts != nil {
		user.BankAccounts = req.BankAccounts
	}
	if req.PreferredPayment != "" {
		user.PreferredPayment = req.PreferredPayment
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
