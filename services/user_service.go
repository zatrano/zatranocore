package services

import (
	"log"
	"zatrano/models"
	"zatrano/repositories"
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService() *UserService {
	return &UserService{repo: repositories.NewUserRepository()}
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.repo.FindAll()
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.repo.Create(user)
}

func (s *UserService) UpdateUser(id uint, userData *models.User) error {
	// Mevcut kullanıcıyı bul
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// Tüm alanları güncelle
	user.Name = userData.Name
	user.Account = userData.Account
	user.Status = userData.Status
	user.Type = userData.Type
	user.TeamID = userData.TeamID // TeamID'yi açıkça güncelle

	// Şifre güncellemesi
	if userData.Password != "" {
		if err := user.SetPassword(userData.Password); err != nil {
			return err
		}
	}

	return s.repo.Update(user)
}

func (s *UserService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}

func (s *UserService) GetUserCount() (int64, error) {
	count, err := s.repo.Count()
	if err != nil {
		log.Printf("UserService: Error getting user count from repository: %v", err)
		// Belki burada 0 döndürmek ve hatayı loglamak yeterlidir,
		// ya da hatayı yukarıya iletmek. Şimdilik iletiyoruz.
		return 0, err
	}
	return count, nil
}
