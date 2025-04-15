package handlers

import (
	"log"           // Hata loglama için log paketini ekleyin
	"zatrano/utils" // Utils paketini import et

	"github.com/gofiber/fiber/v2"
)

// HomeHandler dashboard ana sayfasını işler ve flash mesajları gösterir
func HomeHandler(c *fiber.Ctx) error {
	// Flash mesajlarını al
	flashData, err := utils.GetFlashMessages(c)
	if err != nil {
		// Hata olursa logla, ama sayfayı render etmeye devam et
		log.Printf("HomeHandler: Flash mesajları alınamadı: %v", err)
	}

	// Render edilecek verileri hazırla (Flash mesajlarını ekle)
	renderData := fiber.Map{
		"Title": "Gösterge Paneli Ana Sayfa",
		// Flash mesajlarını şablona gönder
		"Success": flashData.Success,
		"Error":   flashData.Error,
	}

	// Sayfayı render et
	return c.Render("dashboard/home/dashboard_home", renderData, "layouts/dashboard_layout")
}
