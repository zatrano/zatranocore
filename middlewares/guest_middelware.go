package middlewares

import (
	"zatrano/models"
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
)

func GuestMiddleware(c *fiber.Ctx) error {
	sess, err := utils.SessionStart(c)
	if err != nil {
		return c.Next() // Oturum başlatılamadı, guest olarak devam et
	}

	userID, err := utils.GetUserIDFromSession(sess)
	if err != nil {
		return c.Next() // Geçerli kullanıcı ID'si yok, guest olarak devam et
	}

	// AuthService ile kullanıcı bilgilerini doğrula
	authService := services.NewAuthService()
	user, err := authService.GetUserProfile(userID)
	if err != nil {
		// Kullanıcı bulunamazsa oturumu temizle
		_ = sess.Destroy()
		return c.Next()
	}

	// Kullanıcı tipine göre yönlendirme yap
	var redirectURL string
	switch user.Type {
	case models.Manager:
		redirectURL = "/manager/home"
	case models.Agent:
		redirectURL = "/agent/home"
	case models.System:
		redirectURL = "/dashboard/home"
	default:
		// Geçersiz kullanıcı tipi durumunda oturumu temizle
		_ = sess.Destroy()
		return c.Next()
	}

	return c.Redirect(redirectURL)
}
