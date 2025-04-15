package middlewares

import (
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware temel kimlik doğrulama kontrolü yapar
func AuthMiddleware(c *fiber.Ctx) error {
	sess, err := utils.SessionStart(c)
	if err != nil {
		return c.Redirect("/auth/login")
	}

	userID, err := utils.GetUserIDFromSession(sess)
	if err != nil {
		return c.Redirect("/auth/login")
	}

	// AuthService ile kullanıcının varlığını doğrula
	authService := services.NewAuthService()
	_, err = authService.GetUserProfile(userID)
	if err != nil {
		// Kullanıcı veritabanında yoksa oturumu temizle ve login sayfasına yönlendir
		_ = sess.Destroy()
		return c.Redirect("/auth/login")
	}

	return c.Next()
}
