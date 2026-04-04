package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/yertaypert/gym-booking-api/internal/config"
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
	authUsecase := usecase.NewAuthUsecase(userRepo)
	authHandler := handler.NewAuthHandler(authUsecase)

	// Setup routes
	mux := http.NewServeMux()

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)

	mux.Handle("/me", middleware.AuthMiddleware(http.HandlerFunc(authHandler.Me)))

	// Run server
	log.Printf("Server running on %s", cfg.ServerPort)
	err = http.ListenAndServe(cfg.ServerPort, mux)
	if err != nil {
		log.Fatal(err)
	}
}
