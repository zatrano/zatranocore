package repositories

import (
	"zatrano/configs"
	"zatrano/models"

	"gorm.io/gorm"
)

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository() *TeamRepository {
	return &TeamRepository{db: configs.GetDB()}
}

func (r *TeamRepository) FindAll() ([]models.Team, error) {
	var teams []models.Team
	err := r.db.Find(&teams).Error
	return teams, err
}

func (r *TeamRepository) FindByID(id uint) (*models.Team, error) {
	var team models.Team
	err := r.db.First(&team, id).Error
	return &team, err
}

func (r *TeamRepository) Create(team *models.Team) error {
	return r.db.Create(team).Error
}

func (r *TeamRepository) Update(team *models.Team) error {
	return r.db.Save(team).Error
}

func (r *TeamRepository) Delete(id uint) error {
	return r.db.Delete(&models.Team{}, id).Error
}

func (r *TeamRepository) Count() (int64, error) {
	var count int64
	// Model belirterek team tablosundaki kayıtları say
	err := r.db.Model(&models.Team{}).Count(&count).Error
	return count, err
}
