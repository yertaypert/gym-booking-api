package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/yertaypert/gym-booking-api/internal/config"
	"github.com/yertaypert/gym-booking-api/internal/infrastructure/database"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()
	db := database.NewDB(cfg)
	defer db.Close()

	ctx := context.Background()

	adminID, err := seedAdmin(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	if shouldSeedDemoData() {
		if err := seedDemoData(ctx, db, adminID); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Skipping demo data seeding (SEED_DEMO_DATA is false)")
	}
}

func seedAdmin(ctx context.Context, db *sql.DB) (int, error) {
	email := strings.ToLower(strings.TrimSpace(getRequiredEnv("SEED_ADMIN_EMAIL")))
	password := getRequiredEnv("SEED_ADMIN_PASSWORD")
	fullName := strings.TrimSpace(getEnv("SEED_ADMIN_FULL_NAME", "Initial Admin"))

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash admin password: %w", err)
	}

	var id int
	err = db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password_hash, full_name, role, balance)
		VALUES ($1, $2, $3, 'admin', 0)
		ON CONFLICT (email) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
			full_name = EXCLUDED.full_name,
			role = 'admin'
		RETURNING id`,
		email,
		string(hashedPassword),
		fullName,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("seed admin user: %w", err)
	}

	log.Printf("Seeded admin user %s with id %d", email, id)
	return id, nil
}

func seedDemoData(ctx context.Context, db *sql.DB, ownerID int) error {
	log.Println("Seeding demo data...")
	gyms := []struct {
		name    string
		address string
		desc    string
	}{
		{"Downtown Gym", "123 Main St", "Open 24/7 demo gym"},
		{"Fitness First", "456 Oak Ave", "High-end fitness center"},
		{"Iron Works", "789 Industrial Rd", "Old school bodybuilding gym"},
	}

	classes := []struct {
		name        string
		maxCapacity int
	}{
		{"Yoga", 20},
		{"HIIT", 15},
		{"Pilates", 12},
	}

	for _, g := range gyms {
		gymID, err := getOrCreateGym(ctx, db, g.name, g.address, g.desc, ownerID)
		if err != nil {
			return err
		}

		for _, c := range classes {
			classID, maxCapacity, err := getOrCreateClass(ctx, db, gymID, c.name, c.maxCapacity)
			if err != nil {
				return err
			}

			// Seed 3 sessions for each class at different times
			sessionTimes := []struct {
				startOffset time.Duration
				duration    time.Duration
				price       float64
			}{
				{24 * time.Hour, time.Hour, 15.0},
				{48 * time.Hour, 90 * time.Minute, 20.0},
				{72 * time.Hour, time.Hour, 18.0},
			}

			for _, st := range sessionTimes {
				startTime := time.Now().UTC().Add(st.startOffset).Truncate(time.Hour)
				endTime := startTime.Add(st.duration)

				if err := getOrCreateSession(ctx, db, classID, startTime, endTime, maxCapacity, st.price); err != nil {
					return err
				}
			}
		}
	}

	log.Println("Successfully seeded gyms, classes, and sessions data")
	return nil
}

func getOrCreateGym(ctx context.Context, db *sql.DB, name, address, description string, ownerID int) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `SELECT id FROM gyms WHERE name = $1`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("query gym: %w", err)
	}

	err = db.QueryRowContext(
		ctx,
		`INSERT INTO gyms (name, address, description, owner_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		name,
		address,
		description,
		ownerID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert gym: %w", err)
	}

	return id, nil
}

func getOrCreateClass(ctx context.Context, db *sql.DB, gymID int, name string, maxCapacity int) (int, int, error) {
	var id int
	var existingCapacity int
	err := db.QueryRowContext(
		ctx,
		`SELECT id, max_capacity FROM classes WHERE gym_id = $1 AND name = $2`,
		gymID,
		name,
	).Scan(&id, &existingCapacity)
	if err == nil {
		return id, existingCapacity, nil
	}
	if err != sql.ErrNoRows {
		return 0, 0, fmt.Errorf("query class: %w", err)
	}

	err = db.QueryRowContext(
		ctx,
		`INSERT INTO classes (gym_id, name, max_capacity) VALUES ($1, $2, $3) RETURNING id, max_capacity`,
		gymID,
		name,
		maxCapacity,
	).Scan(&id, &existingCapacity)
	if err != nil {
		return 0, 0, fmt.Errorf("insert class: %w", err)
	}

	return id, existingCapacity, nil
}

func getOrCreateSession(ctx context.Context, db *sql.DB, classID int, startTime, endTime time.Time, availableSlots int, price float64) error {
	var id int
	err := db.QueryRowContext(
		ctx,
		`SELECT id FROM class_sessions WHERE class_id = $1 AND start_time = $2 AND end_time = $3`,
		classID,
		startTime,
		endTime,
	).Scan(&id)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("query session: %w", err)
	}

	err = db.QueryRowContext(
		ctx,
		`INSERT INTO class_sessions (class_id, start_time, end_time, available_slots, price, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
		RETURNING id`,
		classID,
		startTime,
		endTime,
		availableSlots,
		price,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	return nil
}

func shouldSeedDemoData() bool {
	value := strings.ToLower(strings.TrimSpace(getEnv("SEED_DEMO_DATA", "true")))
	return value == "1" || value == "true" || value == "yes"
}


func getRequiredEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		log.Fatalf("%s is required", key)
	}

	return value
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}
