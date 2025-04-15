package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	FlashSuccessKey = "flash_success_message"
	FlashErrorKey   = "flash_error_message"
)

type FlashMessagesData struct {
	Success string
	Error   string
}

func SetFlashMessage(c *fiber.Ctx, key string, message string) error {
	sess, err := SessionStart(c)
	if err != nil {
		fmt.Printf("Error starting session for flash message: %v\n", err)

		return fmt.Errorf("session başlatılamadı: %w", err)
	}
	sess.Set(key, message)
	if err := sess.Save(); err != nil {
		fmt.Printf("Error saving session for flash message: %v\n", err)

		return fmt.Errorf("session kaydedilemedi: %w", err)
	}
	return nil
}

func GetFlashMessages(c *fiber.Ctx) (FlashMessagesData, error) {
	messages := FlashMessagesData{}
	sess, err := SessionStart(c)
	if err != nil {
		fmt.Printf("Error starting session to get flash messages: %v\n", err)

		return messages, fmt.Errorf("flash mesajları için session başlatılamadı: %w", err)
	}

	var sessionNeedsSave bool

	if success := sess.Get(FlashSuccessKey); success != nil {
		if msg, ok := success.(string); ok {
			messages.Success = msg
			sess.Delete(FlashSuccessKey)
			sessionNeedsSave = true
		}
	}

	if errorFlash := sess.Get(FlashErrorKey); errorFlash != nil {
		if msg, ok := errorFlash.(string); ok {
			messages.Error = msg
			sess.Delete(FlashErrorKey)
			sessionNeedsSave = true
		}
	}

	if sessionNeedsSave {
		if err := sess.Save(); err != nil {
			fmt.Printf("Error saving session after getting flash messages: %v\n", err)
			// Kaydetme hatası olsa bile okunan mesajları ve hatayı döndür
			return messages, fmt.Errorf("flash mesajları alındıktan sonra session kaydedilemedi: %w", err)
		}
	}

	return messages, nil
}
