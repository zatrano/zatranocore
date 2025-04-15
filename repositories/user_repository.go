package repositories

import (
	"zatrano/configs"
	"zatrano/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{db: configs.GetDB()}
}

func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Preload("Team").Find(&users).Error
	return users, err
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Team").First(&user, id).Error
	return &user, err
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *models.User) error {
	// Select ile güncellenecek alanları açıkça belirt
	return r.db.Model(user).Select(
		"Name",
		"Account",
		"Password",
		"Status",
		"Type",
		"TeamID", // TeamID'yi açıkça ekledik
		"UpdatedAt",
	).Updates(user).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *UserRepository) Count() (int64, error) {
	var count int64
	// Model belirterek user tablosundaki kayıtları say
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}
