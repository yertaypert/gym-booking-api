package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/yertaypert/gym-booking-api/internal/config"
	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/handler"
	"github.com/yertaypert/gym-booking-api/internal/infrastructure/database"
	"github.com/yertaypert/gym-booking-api/internal/middleware"
	"github.com/yertaypert/gym-booking-api/internal/repository"
	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

func main() {
	// Conf
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	db := database.NewDB(cfg)

	// Init layers
	userRepo := repository.NewUserRepository(db)
	gymRepo := repository.NewGymRepository(db)
	authUsecase := usecase.NewAuthUsecase(userRepo)
	gymUsecase := usecase.NewGymUsecase(gymRepo)
	authHandler := handler.NewAuthHandler(authUsecase)
	gymHandler := handler.NewGymHandler(gymUsecase)

	// Setup routes
	mux := http.NewServeMux()

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("GET /gyms", gymHandler.ListGyms)
	mux.HandleFunc("GET /gyms/{id}", gymHandler.GetGym)
	mux.HandleFunc("GET /gyms/{id}/classes", gymHandler.ListGymClasses)
	mux.HandleFunc("GET /gyms/{gymId}/classes/{classId}/sessions", gymHandler.ListClassSessions)
	mux.Handle(
		"POST /gyms",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin)(http.HandlerFunc(gymHandler.CreateGym)),
		),
	)
	mux.Handle(
		"POST /gyms/{id}/classes",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin)(http.HandlerFunc(gymHandler.CreateClass)),
		),
	)
	mux.Handle(
		"POST /gyms/{gymId}/classes/{classId}/sessions",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin)(http.HandlerFunc(gymHandler.CreateSession)),
		),
	)

	mux.Handle("/me", middleware.AuthMiddleware(http.HandlerFunc(authHandler.Me)))
	mux.Handle(
		"/admin/me",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin)(http.HandlerFunc(authHandler.Me)),
		),
	)

	// Run server
	log.Printf("Server running on %s", cfg.ServerPort)
	err = http.ListenAndServe(cfg.ServerPort, mux)
	if err != nil {
		log.Fatal(err)
	}
}
