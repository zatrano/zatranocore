package handlers

import (
	"log"
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
)

type HomeHandler struct {
	userService *services.UserService
	teamService *services.TeamService
}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{
		userService: services.NewUserService(),
		teamService: services.NewTeamService(),
	}
}

func (h *HomeHandler) HomePage(c *fiber.Ctx) error {
	// Flash mesajlarını al
	flashData, err := utils.GetFlashMessages(c)
	if err != nil {
		// Hata olursa logla, ama sayfayı render etmeye devam et
		log.Printf("HomeHandler: Flash mesajları alınamadı: %v", err)
	}

	// Kullanıcı sayısını al
	userCount, userErr := h.userService.GetUserCount()
	if userErr != nil {
		// Hata olursa logla ve belki kullanıcıya bir hata göster
		log.Printf("HomeHandler: Kullanıcı sayısı alınamadı: %v", userErr)
		// Hata durumunda sayaçları 0 olarak ayarlayabilir veya hata mesajı ekleyebilirsiniz
		// flashData.Error = append(flashData.Error, "Kullanıcı sayısı alınamadı.") // İsteğe bağlı: Flash mesajına ekle
		userCount = 0 // Hata durumunda 0 göster
	}

	// Takım sayısını al
	teamCount, teamErr := h.teamService.GetTeamCount()
	if teamErr != nil {
		// Hata olursa logla
		log.Printf("HomeHandler: Takım sayısı alınamadı: %v", teamErr)
		// flashData.Error = append(flashData.Error, "Takım sayısı alınamadı.") // İsteğe bağlı: Flash mesajına ekle
		teamCount = 0 // Hata durumunda 0 göster
	}

	mapData := fiber.Map{
		"Title": "Dashboard",
		// Flash mesajları
		"Success": flashData.Success,
		"Error":   flashData.Error,
		// Sayılar
		"UserCount": userCount,
		"TeamCount": teamCount,
	}

	return c.Render("dashboard/home/dashboard_home", mapData, "layouts/dashboard_layout")
}
