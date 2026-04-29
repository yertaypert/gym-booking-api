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
	"github.com/yertaypert/gym-booking-api/pkg/logger"
)

func main() {
	globalLogger := logger.SetupLogger()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()
	db := database.NewDB(cfg)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	gymRepo := repository.NewGymRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	classRepo := repository.NewClassRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	trainerSlotRepo := repository.NewTrainerSlotRepository(db)
	trainerBookingRepo := repository.NewTrainerBookingRepository(db)

	// Usecases
	authUsecase := usecase.NewAuthUsecase(userRepo)
	gymUsecase := usecase.NewGymUsecase(gymRepo)
	bookingUsecase := usecase.NewBookingUsecase(db, bookingRepo, walletRepo, userRepo, sessionRepo, gymRepo)
	classUsecase := usecase.NewClassUsecase(classRepo)
	transactionUsecase := usecase.NewTransactionUsecase(transactionRepo)
	walletUsecase := usecase.NewWalletUsecase(db, walletRepo)
	trainerBookingUsecase := usecase.NewTrainerBookingUsecase(trainerSlotRepo, trainerBookingRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	gymHandler := handler.NewGymHandler(gymUsecase)
	bookingHandler := handler.NewBookingHandler(bookingUsecase)
	classHandler := handler.NewClassHandler(classUsecase)
	transactionHandler := handler.NewTransactionHandler(transactionUsecase)
	walletHandler := handler.NewWalletHandler(walletUsecase)
	trainerBookingHandler := handler.NewTrainerBookingHandler(trainerBookingUsecase)
	mux := http.NewServeMux()

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)

	mux.HandleFunc("GET /gyms", gymHandler.ListGyms)
	mux.HandleFunc("GET /gyms/{id}", gymHandler.GetGym)
	mux.HandleFunc("GET /gyms/{id}/classes", gymHandler.ListGymClasses)
	mux.HandleFunc("GET /gyms/{gymId}/classes/{classId}/sessions", gymHandler.ListClassSessions)

	mux.HandleFunc("GET /classes", classHandler.ListClasses)
	mux.HandleFunc("GET /classes/{name}", classHandler.ListGymsByClass)
	mux.HandleFunc("GET /classes/{name}/sessions", classHandler.SearchSessions)
	mux.HandleFunc("GET /sessions/{id}", classHandler.GetSession)

	mux.Handle(
		"GET /me/bookings",
		middleware.AuthMiddleware(http.HandlerFunc(bookingHandler.GetMyBookings)),
	)
	mux.Handle(
		"POST /gyms",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin)(http.HandlerFunc(gymHandler.CreateGym)),
		),
	)
	mux.Handle(
		"GET /me/gyms",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin, domain.RoleGymOwner)(http.HandlerFunc(gymHandler.ListMyGyms)),
		),
	)
	mux.Handle(
		"POST /gyms/{id}/classes",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin, domain.RoleGymOwner)(http.HandlerFunc(gymHandler.CreateClass)),
		),
	)
	mux.Handle(
		"POST /gyms/{gymId}/classes/{classId}/sessions",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin, domain.RoleGymOwner)(http.HandlerFunc(gymHandler.CreateSession)),
		),
	)
	mux.Handle(
		"POST /trainer-slots/{id}/book",
		middleware.AuthMiddleware(http.HandlerFunc(trainerBookingHandler.BookTrainerSlot)),
	)
	mux.Handle(
		"POST /sessions/{sessionId}/bookings",
		middleware.AuthMiddleware(http.HandlerFunc(bookingHandler.CreateBooking)),
	)
	mux.Handle(
		"POST /bookings/{bookingId}/cancel",
		middleware.AuthMiddleware(http.HandlerFunc(bookingHandler.CancelBooking)),
	)
	mux.Handle(
		"GET /gyms/{id}/bookings",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin, domain.RoleGymOwner)(http.HandlerFunc(bookingHandler.ListGymBookings)),
		),
	)
	mux.Handle(
		"POST /gyms/{id}/trainers",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin, domain.RoleGymOwner)(http.HandlerFunc(gymHandler.AssignTrainer)),
		),
	)
	// Mark a booking as attended — admin only
	mux.Handle(
		"POST /bookings/{bookingId}/attend",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin)(http.HandlerFunc(bookingHandler.MarkAttended)),
		),
	)
	// Admin views all attendees for a session
	mux.Handle(
		"GET /sessions/{sessionId}/bookings",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin)(http.HandlerFunc(bookingHandler.GetSessionAttendees)),
		),
	)
	mux.Handle(
		"GET /transactions",
		middleware.AuthMiddleware(http.HandlerFunc(transactionHandler.GetMyTransactions)),
	)
	mux.Handle(
		"POST /wallet/topup",
		middleware.AuthMiddleware(http.HandlerFunc(walletHandler.TopUp)),
	)

	mux.Handle("/me", middleware.AuthMiddleware(http.HandlerFunc(authHandler.Me)))
	mux.Handle(
		"/admin/me",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin)(http.HandlerFunc(authHandler.Me)),
		),
	)

	loggedMux := middleware.RequestLogger(mux.ServeHTTP)

	globalLogger.Info("Server is starting", "port", cfg.ServerPort)

	err = http.ListenAndServe(cfg.ServerPort, loggedMux)
	if err != nil {
		globalLogger.Error("Server crashed", "error", err.Error())
	}
}
