package handlers

import (
	"log" // Hata loglama için log paketini ekleyin
	"strconv"
	"zatrano/models"
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service     *services.UserService
	teamService *services.TeamService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		service:     services.NewUserService(),
		teamService: services.NewTeamService(),
	}
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	// Flash mesajlarını yeni yöntemle al
	flashData, flashErr := utils.GetFlashMessages(c)
	if flashErr != nil {
		log.Printf("ListUsers: Flash mesajları alınamadı: %v", flashErr)
	}

	users, dbErr := h.service.GetAllUsers() // Veritabanından kullanıcıları al

	// Render edilecek temel verileri hazırla
	renderData := fiber.Map{
		"Title":     "Kullanıcılar",
		"CsrfToken": c.Locals("csrf"),
		"Users":     users,
		"Success":   flashData.Success,
		"Error":     flashData.Error,
	}

	// Veritabanından veri alırken hata oluştuysa
	if dbErr != nil {
		dbErrMsg := "Kullanıcılar getirilirken bir hata oluştu."
		log.Printf("ListUsers DB Error: %v", dbErr)

		if renderData["Error"] != "" {
			renderData["Error"] = renderData["Error"].(string) + " | " + dbErrMsg
		} else {
			renderData["Error"] = dbErrMsg
		}
		renderData["Users"] = []models.User{}
		return c.Render("dashboard/users/dashboard_users_list", renderData, "layouts/dashboard_layout")
	}

	// Veritabanı hatası yoksa, normal şekilde render et
	return c.Render("dashboard/users/dashboard_users_list", renderData, "layouts/dashboard_layout")
}

func (h *UserHandler) ShowCreateUser(c *fiber.Ctx) error {
	// Bu fonksiyon GetFlashMessages kullanmıyor, değişiklik gerekmez.
	// Ancak, eğer bu sayfada da flash mesaj göstermek isterseniz, ListUsers'daki gibi ekleyebilirsiniz.
	teams, err := h.teamService.GetAllTeams()
	mapData := fiber.Map{
		"Title":     "Yeni Kullanıcı Ekle",
		"CsrfToken": c.Locals("csrf"),
	}
	if err != nil {
		// Takımlar getirilemezse, formu yine de göster ama bir hata mesajı ekle
		log.Printf("ShowCreateUser: Takımlar alınamadı: %v", err)
		mapData["Error"] = "Takım listesi yüklenemedi, ancak kullanıcı ekleyebilirsiniz."
		mapData["Teams"] = []models.Team{} // Boş liste gönder
	} else {
		mapData["Teams"] = teams
	}

	// successFlash, errorFlash := utils.GetFlashMessages(c) // İsterseniz eklenebilir
	// mapData["Success"] = successFlash
	// mapData["Error"] = errorFlash // Veya mevcut hatayla birleştir

	return c.Render("dashboard/users/dashboard_users_create", mapData, "layouts/dashboard_layout")
}

// renderCreateUserFormWithError GetFlashMessages kullanmıyor, doğrudan Error parametresi alıyor. Değişiklik gerekmez.
func (h *UserHandler) renderCreateUserFormWithError(c *fiber.Ctx, errorMsg string, statusCode int, formData interface{}) error {
	teams, teamErr := h.teamService.GetAllTeams()
	mapData := fiber.Map{
		"Title":     "Yeni Kullanıcı Ekle",
		"CsrfToken": c.Locals("csrf"),
		"Error":     errorMsg,
		"FormData":  formData, // Form verisini geri göndererek kullanıcının tekrar doldurmasını kolaylaştır
	}
	if teamErr != nil {
		log.Printf("renderCreateUserFormWithError: Takımlar alınamadı: %v", teamErr)
		mapData["Error"] = errorMsg + " (Ayrıca takım listesi yüklenemedi.)"
		mapData["Teams"] = []models.Team{}
	} else {
		mapData["Teams"] = teams
	}
	return c.Status(statusCode).Render("dashboard/users/dashboard_users_create", mapData, "layouts/dashboard_layout")
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	// Bu fonksiyon GetFlashMessages kullanmıyor, değişiklik gerekmez.
	// Sadece SetFlashMessage kullanıyor ve hata durumunda renderCreateUserFormWithError çağırıyor.
	type Request struct {
		Name     string `form:"name"`
		Account  string `form:"account"`
		Password string `form:"password"`
		Status   string `form:"status"`
		Type     string `form:"type"`
		TeamID   string `form:"team_id"` // Formdan string olarak gelir
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return h.renderCreateUserFormWithError(c, "Geçersiz veri formatı veya eksik alanlar.", fiber.StatusBadRequest, req)
	}

	// Formdan gelen değerleri işle
	status := req.Status != "false"
	var teamID *uint                           // Pointer olarak tanımla, nil olabilir
	if req.TeamID != "" && req.TeamID != "0" { // "0" veya boş değilse işle
		id, err := strconv.ParseUint(req.TeamID, 10, 32)
		if err != nil {
			return h.renderCreateUserFormWithError(c, "Geçersiz takım ID formatı.", fiber.StatusBadRequest, req)
		}
		// Takımın varlığını kontrol et
		uintVal := uint(id)
		if _, err := h.teamService.GetTeamByID(uintVal); err != nil {
			log.Printf("CreateUser: Geçersiz takım ID'si %d seçildi: %v", uintVal, err)
			return h.renderCreateUserFormWithError(c, "Seçilen takım bulunamadı.", fiber.StatusBadRequest, req)
		}
		teamID = &uintVal // Geçerliyse pointer'a ata
	}
	// else { teamID nil kalır }

	if req.Password == "" {
		return h.renderCreateUserFormWithError(c, "Şifre alanı zorunludur.", fiber.StatusBadRequest, req)
	}

	// Kullanıcı modelini oluştur
	user := models.User{
		Name:    req.Name,
		Account: req.Account,
		Status:  status,
		Type:    models.UserType(req.Type), // String'i UserType'a çevir
		TeamID:  teamID,                    // Pointer'ı ata (nil olabilir)
	}
	// Şifreyi hash'le
	if err := user.SetPassword(req.Password); err != nil {
		log.Printf("CreateUser: Şifre hashlenemedi: %v", err)
		return h.renderCreateUserFormWithError(c, "Şifre oluşturulurken bir hata oluştu.", fiber.StatusInternalServerError, req)
	}

	// Kullanıcıyı veritabanına kaydet
	if err := h.service.CreateUser(&user); err != nil {
		log.Printf("CreateUser DB Error: %v", err)
		// Hata mesajını daha spesifik hale getirebiliriz (örn: duplicate account)
		return h.renderCreateUserFormWithError(c, "Kullanıcı oluşturulamadı (örn: hesap adı zaten mevcut olabilir).", fiber.StatusInternalServerError, req)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Kullanıcı başarıyla oluşturuldu.")
	return c.Redirect("/dashboard/users", fiber.StatusFound)
}

func (h *UserHandler) ShowUpdateUser(c *fiber.Ctx) error {
	// Bu fonksiyon GetFlashMessages kullanmıyor, değişiklik gerekmez.
	// Ancak, eğer bu sayfada da flash mesaj göstermek isterseniz, ListUsers'daki gibi ekleyebilirsiniz.
	id, err := c.ParamsInt("id")
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz kullanıcı ID'si.")
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}

	user, err := h.service.GetUserByID(uint(id))
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Kullanıcı bulunamadı.")
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}

	teams, err := h.teamService.GetAllTeams()
	mapData := fiber.Map{
		"Title":     "Kullanıcı Düzenle",
		"User":      user,
		"CsrfToken": c.Locals("csrf"),
		// TeamID'yi şablonda seçili göstermek için hazırlayalım
	}
	if user.TeamID != nil {
		mapData["SelectedTeamID"] = *user.TeamID // Eğer varsa ID'yi gönder
	} else {
		mapData["SelectedTeamID"] = 0 // Yoksa 0 veya başka bir değer gönder
	}

	if err != nil {
		log.Printf("ShowUpdateUser: Takımlar alınamadı: %v", err)
		mapData["Error"] = "Takım listesi yüklenemedi."
		mapData["Teams"] = []models.Team{}
	} else {
		mapData["Teams"] = teams
	}

	// successFlash, errorFlash := utils.GetFlashMessages(c) // İsterseniz eklenebilir
	// mapData["Success"] = successFlash
	// mapData["Error"] = errorFlash // Veya mevcut hatayla birleştir

	return c.Render("dashboard/users/dashboard_users_update", mapData, "layouts/dashboard_layout")
}

// renderUpdateUserFormWithError GetFlashMessages kullanmıyor, doğrudan Error parametresi alıyor. Değişiklik gerekmez.
func (h *UserHandler) renderUpdateUserFormWithError(c *fiber.Ctx, userID uint, errorMsg string, statusCode int, formData interface{}) error {
	// Bu yardımcı fonksiyon, formu hata mesajı ve önceki verilerle tekrar render eder.
	user, userErr := h.service.GetUserByID(userID)
	if userErr != nil {
		// Kullanıcı bile yüklenemiyorsa ana listeye yönlendir.
		log.Printf("renderUpdateUserFormWithError: Kullanıcı %d yüklenemedi: %v", userID, userErr)
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, errorMsg+" (Ek olarak: Kullanıcı bilgileri yüklenemedi)")
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}

	teams, teamErr := h.teamService.GetAllTeams()
	mapData := fiber.Map{
		"Title":     "Kullanıcı Düzenle",
		"User":      user, // Mevcut kullanıcı verisi (şifre hariç)
		"CsrfToken": c.Locals("csrf"),
		"Error":     errorMsg,
		"FormData":  formData, // Hatalı giriş yapılan veriyi geri gönder
	}
	// Şablonda seçili takımı göstermek için
	if user.TeamID != nil {
		mapData["SelectedTeamID"] = *user.TeamID
	} else {
		mapData["SelectedTeamID"] = 0
	}

	if teamErr != nil {
		log.Printf("renderUpdateUserFormWithError: Takımlar alınamadı: %v", teamErr)
		mapData["Error"] = errorMsg + " (Ayrıca takım listesi yüklenemedi.)"
		mapData["Teams"] = []models.Team{}
	} else {
		mapData["Teams"] = teams
	}

	// Dikkat: Şablon dosyasının yolu muhtemelen "dashboard/users/dashboard_users_update" olmalı.
	return c.Status(statusCode).Render("dashboard/users/dashboard_users_update", mapData, "layouts/dashboard_layout")
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	// Bu fonksiyon GetFlashMessages kullanmıyor, değişiklik gerekmez.
	// Sadece SetFlashMessage kullanıyor ve hata durumunda renderUpdateUserFormWithError çağırıyor.
	id, err := c.ParamsInt("id")
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz kullanıcı ID'si.")
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}
	userID := uint(id)

	type Request struct {
		Name     string `form:"name"`
		Account  string `form:"account"`
		Password string `form:"password"` // Şifre boş bırakılırsa güncellenmez
		Status   string `form:"status"`
		Type     string `form:"type"`
		TeamID   string `form:"team_id"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return h.renderUpdateUserFormWithError(c, userID, "Geçersiz veri formatı veya eksik alanlar.", fiber.StatusBadRequest, req)
	}

	// Formdan gelen değerleri işle
	status := req.Status != "false"
	var teamID *uint
	if req.TeamID != "" && req.TeamID != "0" {
		tid, err := strconv.ParseUint(req.TeamID, 10, 32)
		if err != nil {
			return h.renderUpdateUserFormWithError(c, userID, "Geçersiz takım ID formatı.", fiber.StatusBadRequest, req)
		}
		uintVal := uint(tid)
		if _, err := h.teamService.GetTeamByID(uintVal); err != nil {
			log.Printf("UpdateUser: Geçersiz takım ID'si %d seçildi: %v", uintVal, err)
			return h.renderUpdateUserFormWithError(c, userID, "Seçilen takım bulunamadı.", fiber.StatusBadRequest, req)
		}
		teamID = &uintVal
	}
	// else { teamID nil kalır }

	// Mevcut kullanıcıyı bul
	existingUser, err := h.service.GetUserByID(userID)
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Güncellenecek kullanıcı bulunamadı.")
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}

	// Kullanıcı bilgilerini güncelle
	existingUser.Name = req.Name
	existingUser.Account = req.Account
	existingUser.Status = status
	existingUser.Type = models.UserType(req.Type)
	existingUser.TeamID = teamID

	// Şifre alanı doluysa, şifreyi güncelle
	if req.Password != "" {
		if err := existingUser.SetPassword(req.Password); err != nil {
			log.Printf("UpdateUser: Şifre hashlenemedi (UserID: %d): %v", userID, err)
			return h.renderUpdateUserFormWithError(c, userID, "Şifre güncellenirken bir hata oluştu.", fiber.StatusInternalServerError, req)
		}
	}
	// else { Şifre değişmez }

	// Kullanıcıyı veritabanında güncelle
	if err := h.service.UpdateUser(userID, existingUser); err != nil {
		log.Printf("UpdateUser DB Error (UserID: %d): %v", userID, err)
		// Hata mesajını daha spesifik hale getir
		return h.renderUpdateUserFormWithError(c, userID, "Kullanıcı güncellenemedi (örn: hesap adı zaten mevcut olabilir).", fiber.StatusInternalServerError, req)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Kullanıcı başarıyla güncellendi.")
	return c.Redirect("/dashboard/users", fiber.StatusFound)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	// Bu fonksiyon GetFlashMessages kullanmıyor, değişiklik gerekmez.
	// Sadece SetFlashMessage kullanıyor.
	id, err := c.ParamsInt("id")
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz kullanıcı ID'si.")
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}
	userID := uint(id)

	if err := h.service.DeleteUser(userID); err != nil {
		log.Printf("DeleteUser Error (UserID: %d): %v", userID, err)
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Kullanıcı silinemedi.") // Genel hata
		return c.Redirect("/dashboard/users", fiber.StatusSeeOther)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Kullanıcı başarıyla silindi.")
	return c.Redirect("/dashboard/users", fiber.StatusFound)
}
