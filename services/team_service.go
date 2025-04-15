package services

import (
	"log"
	"zatrano/models"
	"zatrano/repositories"
)

type TeamService struct {
	repo *repositories.TeamRepository
}

func NewTeamService() *TeamService {
	return &TeamService{repo: repositories.NewTeamRepository()}
}

func (s *TeamService) GetAllTeams() ([]models.Team, error) {
	return s.repo.FindAll()
}

func (s *TeamService) GetTeamByID(id uint) (*models.Team, error) {
	return s.repo.FindByID(id)
}

func (s *TeamService) CreateTeam(team *models.Team) error {
	return s.repo.Create(team)
}

func (s *TeamService) UpdateTeam(id uint, teamData *models.Team) error {
	team, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	team.Name = teamData.Name
	team.Status = teamData.Status

	return s.repo.Update(team)
}

func (s *TeamService) DeleteTeam(id uint) error {
	return s.repo.Delete(id)
}

func (s *TeamService) GetTeamCount() (int64, error) {
	count, err := s.repo.Count()
	if err != nil {
		log.Printf("TeamService: Error getting team count from repository: %v", err)
		// Hata yönetimi: 0 döndür ve logla veya hatayı yukarı ilet. Şimdilik iletiyoruz.
		return 0, err
	}
	return count, nil
}
