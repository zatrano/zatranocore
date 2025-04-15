package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"zatrano/configs"
	"zatrano/routes"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	configs.InitDB()
	defer configs.CloseDB()

	configs.InitSession()

	engine := html.New("./views", ".html")

	engine.AddFunc("getFlashMessages", utils.GetFlashMessages)

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Static("/", "./public")

	setupMiddlewares(app)

	routes.SetupRoutes(app, configs.GetDB())

	startServer(app)
}

func setupMiddlewares(app *fiber.App) {
	app.Use(configs.SetupCSRF())
	app.Use(func(c *fiber.Ctx) error {
		return c.Next()
	})
}

func startServer(app *fiber.App) {

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		port := os.Getenv("APP_PORT")
		if port == "" {
			port = "3000"
		}
		log.Printf("Uygulama http://localhost:%s adresinde başlatılıyor...", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Sunucu başlatılamadı: %v", err)
		}
	}()

	<-shutdown
	log.Println("Uygulama kapatılıyor...")

	if err := app.Shutdown(); err != nil {
		log.Printf("Sunucu kapatılırken hata: %v", err)
	}

	log.Println("Uygulama başarıyla kapatıldı")
}
