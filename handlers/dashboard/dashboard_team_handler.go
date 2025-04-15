package handlers

import (
	"log" // Hata loglama için log paketini ekleyin
	"zatrano/models"
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
)

type TeamHandler struct {
	service *services.TeamService
}

func NewTeamHandler() *TeamHandler {
	return &TeamHandler{service: services.NewTeamService()}
}

func (h *TeamHandler) ListTeams(c *fiber.Ctx) error {
	// Flash mesajlarını yeni yöntemle al
	flashData, flashErr := utils.GetFlashMessages(c)
	if flashErr != nil {
		log.Printf("ListTeams: Flash mesajları alınamadı: %v", flashErr)
		// Hata olsa bile devam et, flashData varsayılan (boş) değerlere sahip olacak
	}

	teams, dbErr := h.service.GetAllTeams() // Veritabanından takımları al

	// Render edilecek temel verileri hazırla
	renderData := fiber.Map{
		"Title":     "Takımlar",
		"CsrfToken": c.Locals("csrf"),
		"Teams":     teams,             // Başlangıçta alınan takımlar (hata varsa nil/boş olabilir)
		"Success":   flashData.Success, // Flash başarı mesajı
		"Error":     flashData.Error,   // Flash hata mesajı
	}

	// Veritabanından veri alırken hata oluştuysa
	if dbErr != nil {
		dbErrMsg := "Takımlar getirilirken bir hata oluştu." // Kullanıcıya gösterilecek genel mesaj
		log.Printf("ListTeams DB Error: %v", dbErr)          // Asıl hatayı logla

		// Mevcut flash hatasının yanına (veya yerine) veritabanı hatasını ekle
		if renderData["Error"] != "" {
			// Zaten bir flash hatası varsa, ikisini birleştir (veya sadece db hatasını göster)
			renderData["Error"] = renderData["Error"].(string) + " | " + dbErrMsg
		} else {
			// Flash hatası yoksa, veritabanı hatasını ata
			renderData["Error"] = dbErrMsg
		}
		// Hata durumunda boş bir liste gönderelim ki şablon hata vermesin
		renderData["Teams"] = []models.Team{}
		// Hata mesajıyla birlikte sayfayı render et (500 yerine 200 OK ile)
		return c.Render("dashboard/teams/dashboard_teams_list", renderData, "layouts/dashboard_layout")
	}

	// Veritabanı hatası yoksa, normal şekilde render et
	return c.Render("dashboard/teams/dashboard_teams_list", renderData, "layouts/dashboard_layout")
}

func (h *TeamHandler) ShowCreateTeam(c *fiber.Ctx) error {
	// Bu fonksiyon GetFlashMessages kullanmıyor, değişiklik gerekmez.
	// Ancak, eğer bu sayfada da flash mesaj göstermek isterseniz, ListTeams'deki gibi ekleyebilirsiniz.
	// successFlash, errorFlash := utils.GetFlashMessages(c) // Gibi
	return c.Render("dashboard/teams/dashboard_teams_create", fiber.Map{
		"Title":     "Yeni Takım Ekle",
		"CsrfToken": c.Locals("csrf"),
		// "Success": successFlash, // Eğer eklerseniz
		// "Error":   errorFlash,   // Eğer eklerseniz
	}, "layouts/dashboard_layout")
}

func (h *TeamHandler) CreateTeam(c *fiber.Ctx) error {
	// Bu fonksiyon GetFlashMessages kullanmıyor, değişiklik gerekmez.
	// Sadece SetFlashMessage kullanıyor.
	type Request struct {
		Name   string `form:"name"`
		Status string `form:"status"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		// Hata durumunda formu tekrar render ederken flash yerine doğrudan Error gönderiyor.
		return c.Status(fiber.StatusBadRequest).Render("dashboard/teams/dashboard_teams_create", fiber.Map{
			"Title":     "Yeni Takım Ekle",
			"CsrfToken": c.Locals("csrf"),
			"Error":     "Lütfen tüm alanları doldurduğunuzdan emin olun.",
			"FormData":  req,
		}, "layouts/dashboard_layout")
	}

	status := req.Status != "false"
	team := models.Team{Name: req.Name, Status: status}

	if err := h.service.CreateTeam(&team); err != nil {
		// Hata durumunda formu tekrar render ederken flash yerine doğrudan Error gönderiyor.
		return c.Status(fiber.StatusInternalServerError).Render("dashboard/teams/dashboard_teams_create", fiber.Map{
			"Title":     "Yeni Takım Ekle",
			"CsrfToken": c.Locals("csrf"),
			"Error":     "Takım oluşturulamadı: " + err.Error(),
			"FormData":  req,
		}, "layouts/dashboard_layout")
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Takım başarıyla oluşturuldu.")
	return c.Redirect("/dashboard/teams", fiber.StatusFound)
}

func (h *TeamHandler) ShowUpdateTeam(c *fiber.Ctx) error {
	// Bu fonksiyon GetFlashMessages kullanmıyor, değişiklik gerekmez.
	// Ancak, eğer bu sayfada da flash mesaj göstermek isterseniz, ListTeams'deki gibi ekleyebilirsiniz.
	id, err := c.ParamsInt("id")
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz takım ID'si.")
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}

	team, err := h.service.GetTeamByID(uint(id))
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Takım bulunamadı.")
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}

	// successFlash, errorFlash := utils.GetFlashMessages(c) // İsterseniz eklenebilir
	return c.Render("dashboard/teams/dashboard_teams_update", fiber.Map{
		"Title":     "Takım Düzenle",
		"Team":      team,
		"CsrfToken": c.Locals("csrf"),
		// "Success": successFlash, // Eğer eklerseniz
		// "Error":   errorFlash,   // Eğer eklerseniz
	}, "layouts/dashboard_layout")
}

func (h *TeamHandler) UpdateTeam(c *fiber.Ctx) error {
	// Bu fonksiyon GetFlashMessages kullanmıyor, değişiklik gerekmez.
	// Sadece SetFlashMessage kullanıyor ve hata durumunda showUpdateTeamFormWithError çağırıyor.
	id, err := c.ParamsInt("id")
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz takım ID'si.")
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}

	type Request struct {
		Name   string `form:"name"`
		Status string `form:"status"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return h.showUpdateTeamFormWithError(c, uint(id), "Lütfen tüm alanları doldurduğunuzdan emin olun.", fiber.StatusBadRequest)
	}

	newStatus := req.Status != "false"
	team, err := h.service.GetTeamByID(uint(id))
	if err != nil {
		// GetTeamByID hatası verirse showUpdateTeamFormWithError çağrılmadan önce flash ayarlanabilir
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Güncellenecek takım bulunamadı.")
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther) // Veya hata sayfasına yönlendir
		// return h.showUpdateTeamFormWithError(c, uint(id), "Güncellenecek takım bulunamadı.", fiber.StatusNotFound) // Bu da bir seçenek
	}

	team.Name = req.Name
	team.Status = newStatus
	if err := h.service.UpdateTeam(uint(id), team); err != nil {
		// Servis hatası durumunda formu tekrar render ediyor
		return h.showUpdateTeamFormWithError(c, uint(id), "Takım güncellenemedi: "+err.Error(), fiber.StatusInternalServerError)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Takım başarıyla güncellendi.")
	return c.Redirect("/dashboard/teams", fiber.StatusFound)
}

// showUpdateTeamFormWithError GetFlashMessages kullanmıyor, doğrudan Error parametresi alıyor. Değişiklik gerekmez.
func (h *TeamHandler) showUpdateTeamFormWithError(c *fiber.Ctx, id uint, errorMsg string, statusCode int) error {
	team, err := h.service.GetTeamByID(id)
	if err != nil {
		// Bu durumda da flash mesaj ayarlayıp ana listeye yönlendirmek daha tutarlı olabilir.
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Takım düzenleme formu yüklenirken hata oluştu.")
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}

	return c.Status(statusCode).Render("dashboard/teams/dashboard_teams_update", fiber.Map{
		"Title":     "Takım Düzenle",
		"Team":      team,
		"CsrfToken": c.Locals("csrf"),
		"Error":     errorMsg, // Hata mesajı doğrudan şablona gönderiliyor
	}, "layouts/dashboard_layout")
}

func (h *TeamHandler) DeleteTeam(c *fiber.Ctx) error {
	// Bu fonksiyon GetFlashMessages kullanmıyor, değişiklik gerekmez.
	// Sadece SetFlashMessage kullanıyor.
	id, err := c.ParamsInt("id")
	if err != nil {
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz takım ID'si.")
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}

	if err := h.service.DeleteTeam(uint(id)); err != nil {
		// Servis hatası durumunda daha açıklayıcı loglama yapılabilir.
		log.Printf("DeleteTeam Error (ID: %d): %v", id, err)
		_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Takım silinemedi. İlişkili kayıtlar olabilir.") // Genel hata mesajı
		return c.Redirect("/dashboard/teams", fiber.StatusSeeOther)
	}

	_ = utils.SetFlashMessage(c, utils.FlashSuccessKey, "Takım başarıyla silindi.")
	return c.Redirect("/dashboard/teams", fiber.StatusFound)
}
