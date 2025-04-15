package routes

import (
	handlers "zatrano/handlers/dashboard"
	"zatrano/middlewares"
	"zatrano/models"

	"github.com/gofiber/fiber/v2"
)

func registerDashboardRoutes(app *fiber.App) {
	dashboardGroup := app.Group("/dashboard")
	dashboardGroup.Use(
		middlewares.AuthMiddleware,
		middlewares.StatusMiddleware,
		middlewares.TypeMiddleware(models.System),
	)

	// Ana sayfa
	dashboardGroup.Get("/home", handlers.HomeHandler)

	// Takım yönetimi - Yeni yapıya uygun
	teamHandler := handlers.NewTeamHandler()
	dashboardGroup.Get("/teams", teamHandler.ListTeams)
	dashboardGroup.Get("/teams/create", teamHandler.ShowCreateTeam)
	dashboardGroup.Post("/teams/create", teamHandler.CreateTeam)
	dashboardGroup.Get("/teams/update/:id", teamHandler.ShowUpdateTeam)
	dashboardGroup.Post("/teams/update/:id", teamHandler.UpdateTeam)
	dashboardGroup.Post("/teams/delete/:id", teamHandler.DeleteTeam)
	dashboardGroup.Delete("/teams/delete/:id", teamHandler.DeleteTeam)

	// Kullanıcı yönetimi - Yeni yapıya uygun
	userHandler := handlers.NewUserHandler()
	dashboardGroup.Get("/users", userHandler.ListUsers)
	dashboardGroup.Get("/users/create", userHandler.ShowCreateUser)
	dashboardGroup.Post("/users/create", userHandler.CreateUser)
	dashboardGroup.Get("/users/update/:id", userHandler.ShowUpdateUser)
	dashboardGroup.Post("/users/update/:id", userHandler.UpdateUser)
	dashboardGroup.Post("/users/delete/:id", userHandler.DeleteUser)
	dashboardGroup.Delete("/users/delete/:id", userHandler.DeleteUser)
}
