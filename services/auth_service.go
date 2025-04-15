package services

import (
	"fmt"

	"zatrano/models"
	"zatrano/repositories"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo *repositories.AuthRepository
}

func NewAuthService() *AuthService {
	return &AuthService{repo: repositories.NewAuthRepository()}
}

func (s *AuthService) Authenticate(account, password string) (*models.User, error) {
	user, err := s.repo.FindUserByAccount(account)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

func (s *AuthService) GetUserProfile(id uint) (*models.User, error) {
	return s.repo.FindUserByID(id)
}

func (s *AuthService) UpdatePassword(userID uint, currentPass, newPassword string) error {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPass)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	if len(newPassword) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	if currentPass == newPassword {
		return fmt.Errorf("new password cannot be same as old password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password")
	}

	user.Password = string(hashedPassword)
	return s.repo.UpdateUser(user)
}
