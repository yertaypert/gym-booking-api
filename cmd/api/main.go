package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/yertaypert/gym-booking-api/internal/auth"
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
		globalLogger.Warn("No .env file found")
	}

	cfg := config.Load()
	auth.SetJWTSecret(cfg.JWTSecret)
	db := database.NewDB(cfg)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	gymRepo := repository.NewGymRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	classRepo := repository.NewClassRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Usecases
	authUsecase := usecase.NewAuthUsecase(userRepo)
	gymUsecase := usecase.NewGymUsecase(gymRepo)
	bookingUsecase := usecase.NewBookingUsecase(db, bookingRepo, walletRepo, userRepo, sessionRepo, gymRepo)
	classUsecase := usecase.NewClassUsecase(classRepo)
	transactionUsecase := usecase.NewTransactionUsecase(transactionRepo)
	walletUsecase := usecase.NewWalletUsecase(db, walletRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	gymHandler := handler.NewGymHandler(gymUsecase)
	bookingHandler := handler.NewBookingHandler(bookingUsecase)
	classHandler := handler.NewClassHandler(classUsecase)
	transactionHandler := handler.NewTransactionHandler(transactionUsecase)
	walletHandler := handler.NewWalletHandler(walletUsecase)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)

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
	mux.Handle(
		"POST /sessions/{sessionId}/attendance-qr",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin, domain.RoleGymOwner)(http.HandlerFunc(bookingHandler.GenerateAttendanceQR)),
		),
	)
	mux.Handle(
		"POST /attendance/scan",
		middleware.AuthMiddleware(http.HandlerFunc(bookingHandler.ScanAttendanceQR)),
	)
	// Admin views all attendees for a session
	mux.Handle(
		"GET /sessions/{sessionId}/bookings",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin, domain.RoleGymOwner)(http.HandlerFunc(bookingHandler.GetSessionAttendees)),
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

	mux.Handle("GET /me", middleware.AuthMiddleware(http.HandlerFunc(authHandler.Me)))
	mux.Handle(
		"GET /admin/me",
		middleware.AuthMiddleware(
			middleware.RequireRoles(domain.RoleAdmin)(http.HandlerFunc(authHandler.Me)),
		),
	)

	loggedMux := middleware.RequestLogger(mux.ServeHTTP)

	srv := &http.Server{
		Addr:         cfg.ServerPort,
		Handler:      loggedMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		globalLogger.Info("Server is starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			globalLogger.Error("Server crashed", "error", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	globalLogger.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		globalLogger.Error("Server forced to shutdown", "error", err)
	}

	if err := db.Close(); err != nil {
		globalLogger.Error("Database connection close error", "error", err)
	}

	globalLogger.Info("Server exiting")
}
