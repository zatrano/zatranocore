package configs

import (
	"log"
	"time"
	zatranoUtils "zatrano/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/utils"
)

func SetupCSRF() fiber.Handler {
	config := csrf.Config{
		KeyLookup:      "form:csrf_token",
		CookieName:     "csrf_",
		CookieHTTPOnly: true,
		CookieSecure:   false,
		CookieSameSite: "Lax",
		Expiration:     1 * time.Hour,
		KeyGenerator:   utils.UUID,
		ContextKey:     "csrf",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			_ = zatranoUtils.SetFlashMessage(c, zatranoUtils.FlashErrorKey, "Geçersiz işlem. Lütfen sayfayı yenileyin.")
			return c.RedirectBack("/auth/login")
		},
	}

	log.Println("CSRF middleware yapılandırıldı")
	return csrf.New(config)
}
