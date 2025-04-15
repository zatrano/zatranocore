package handlers

import (
	"log" // Hata loglama için log paketini ekleyin
	"zatrano/models"
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{service: services.NewAuthService()}
}

func (h *AuthHandler) ShowLogin(c *fiber.Ctx) error {

	// Flash mesajlarını yeni yöntemle al
	flashData, err := utils.GetFlashMessages(c)
	if err != nil {
		log.Printf("ShowLogin: Flash mesajları alınamadı: %v", err)
	}

	return c.Render("auth/auth_login", fiber.Map{
		"Title":     "Giriş",
		"CsrfToken": c.Locals("csrf"),
		"Success":   flashData.Success,
		"Error":     flashData.Error,
	}, "layouts/auth_layout")
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {

	var request struct {
		Account  string `form:"account"`
		Password string `form:"password"`
	}

	if err := c.BodyParser(&request); err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Lütfen tüm alanları doldurduğunuzdan emin olun.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	user, err := h.service.Authenticate(request.Account, request.Password)
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Kullanıcı adı veya şifre hatalı. Lütfen bilgilerinizi kontrol edin.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	sess, err := utils.SessionStart(c)
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Oturum başlatılamadı. Lütfen daha sonra tekrar deneyin.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	sess.Set("user_id", user.ID)
	sess.Set("user_type", user.Type)
	sess.Set("user_name", user.Name)
	if err := sess.Save(); err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Oturum bilgileri kaydedilemedi.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	var redirectURL string
	switch user.Type {
	case models.Manager:
		redirectURL = "/manager/home"
	case models.Agent:
		redirectURL = "/agent/home"
	case models.System:
		redirectURL = "/dashboard/home"
	default:
		_ = sess.Destroy() // Oturumu yok etmeye çalış
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz kullanıcı tipi.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	// --- DEĞİŞİKLİK BURADA ---
	// Başarılı giriş sonrası flash mesajı ayarla
	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Başarıyla giriş yapıldı.")
	// --- DEĞİŞİKLİK SONU ---

	return c.Redirect(redirectURL, fiber.StatusFound)
}

func (h *AuthHandler) Profile(c *fiber.Ctx) error {

	flashData, err := utils.GetFlashMessages(c)
	if err != nil {
		log.Printf("Profile: Flash mesajları alınamadı: %v", err)
	}

	sess, err := utils.SessionStart(c)
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Oturum hatası, lütfen tekrar giriş yapın.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	userIDValue := sess.Get("user_id")
	userID, ok := userIDValue.(uint)
	if !ok {
		_ = sess.Destroy()
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz oturum bilgisi, lütfen tekrar giriş yapın.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	user, err := h.service.GetUserProfile(userID)
	if err != nil {
		_ = sess.Destroy()
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Profil bilgileri alınamadı, lütfen tekrar giriş yapın.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	return c.Render("auth/auth_profile", fiber.Map{
		"Title":     "Profil",
		"User":      user,
		"CsrfToken": c.Locals("csrf"),
		"Success":   flashData.Success,
		"Error":     flashData.Error,
	}, "layouts/auth_layout")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {

	sess, err := utils.SessionStart(c)
	if err != nil {
		return c.Redirect("/auth/login", fiber.StatusFound)
	}

	destroyErr := sess.Destroy()
	if destroyErr != nil {
		log.Printf("Logout: Oturum yok edilirken hata: %v", destroyErr)
		_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Çıkış yapıldı (ancak oturum temizlenirken bir sorun oluştu).")
	} else {
		_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Başarıyla çıkış yapıldı.")
	}

	return c.Redirect("/auth/login", fiber.StatusFound)
}

func (h *AuthHandler) UpdatePassword(c *fiber.Ctx) error {

	sess, err := utils.SessionStart(c)
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Oturum hatası, lütfen tekrar giriş yapın.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	userIDValue := sess.Get("user_id")
	userID, ok := userIDValue.(uint)
	if !ok {
		_ = sess.Destroy()
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz oturum bilgisi, lütfen tekrar giriş yapın.")
		return c.Redirect("/auth/login", fiber.StatusSeeOther)
	}

	var request struct {
		CurrentPassword string `form:"current_password"`
		NewPassword     string `form:"new_password"`
		ConfirmPassword string `form:"confirm_password"`
	}
	if err := c.BodyParser(&request); err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Lütfen tüm şifre alanlarını doldurun.")
		return c.Redirect("/auth/profile", fiber.StatusSeeOther)
	}

	if request.NewPassword != request.ConfirmPassword {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Yeni şifreler uyuşmuyor.")
		return c.Redirect("/auth/profile", fiber.StatusSeeOther)
	}

	if err := h.service.UpdatePassword(userID, request.CurrentPassword, request.NewPassword); err != nil {
		log.Printf("UpdatePassword Error: %v (UserID: %d)", err, userID)
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Şifre güncellenemedi. Mevcut şifrenizi kontrol edin.")
		return c.Redirect("/auth/profile", fiber.StatusSeeOther)
	}

	destroyErr := sess.Destroy()
	if destroyErr != nil {
		log.Printf("UpdatePassword: Oturum yok edilirken hata: %v", destroyErr)
		_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Şifre başarıyla güncellendi.")
	} else {
		_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Şifre başarıyla güncellendi. Lütfen yeni şifrenizle tekrar giriş yapın.")
	}

	return c.Redirect("/auth/login", fiber.StatusFound)
}
