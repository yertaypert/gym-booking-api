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

func seedUser(ctx context.Context, db *sql.DB, email, password, fullName, role string, balance float64) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password for %s: %w", email, err)
	}

	var id int
	err = db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password_hash, full_name, role, balance)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
			full_name = EXCLUDED.full_name,
			role = EXCLUDED.role
		RETURNING id`,
		email,
		string(hashedPassword),
		fullName,
		role,
		balance,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("seed user %s: %w", email, err)
	}

	log.Printf("Seeded %s user %s with id %d", role, email, id)
	return id, nil
}

func seedAdmin(ctx context.Context, db *sql.DB) (int, error) {
	email := strings.ToLower(strings.TrimSpace(getRequiredEnv("SEED_ADMIN_EMAIL")))
	password := getRequiredEnv("SEED_ADMIN_PASSWORD")
	fullName := strings.TrimSpace(getEnv("SEED_ADMIN_FULL_NAME", "Initial Admin"))

	return seedUser(ctx, db, email, password, fullName, "admin", 0)
}

func seedDemoData(ctx context.Context, db *sql.DB, adminID int) error {
	log.Println("Seeding demo data...")

	// Seed some gym owners
	owners := []struct {
		email    string
		fullName string
	}{
		{"owner1@example.com", "John Gym Owner"},
		{"owner2@example.com", "Jane Gym Owner"},
	}

	ownerIDs := make([]int, 0)
	for _, o := range owners {
		id, err := seedUser(ctx, db, o.email, "Password123", o.fullName, "gym_owner", 0)
		if err != nil {
			return err
		}
		ownerIDs = append(ownerIDs, id)
	}

	// Seed some trainers
	trainers := []struct {
		email    string
		fullName string
	}{
		{"trainer1@example.com", "Mike Trainer"},
		{"trainer2@example.com", "Sarah Trainer"},
	}

	trainerIDs := make([]int, 0)
	for _, t := range trainers {
		id, err := seedUser(ctx, db, t.email, "Password123", t.fullName, "trainer", 0)
		if err != nil {
			return err
		}
		trainerIDs = append(trainerIDs, id)
	}

	// Seed some regular users
	users := []struct {
		email    string
		fullName string
		balance  float64
	}{
		{"user1@example.com", "Alice User", 100.0},
		{"user2@example.com", "Bob User", 50.0},
	}

	for _, u := range users {
		_, err := seedUser(ctx, db, u.email, "Password123", u.fullName, "user", u.balance)
		if err != nil {
			return err
		}
	}

	gyms := []struct {
		name    string
		address string
		desc    string
		ownerIdx int
	}{
		{"Downtown Gym", "123 Main St", "Open 24/7 demo gym", 0},
		{"Fitness First", "456 Oak Ave", "High-end fitness center", 0},
		{"Iron Works", "789 Industrial Rd", "Old school bodybuilding gym", 1},
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
		ownerID := ownerIDs[g.ownerIdx]
		gymID, err := getOrCreateGym(ctx, db, g.name, g.address, g.desc, ownerID)
		if err != nil {
			return err
		}

		// Assign a trainer to each gym
		trainerID := trainerIDs[g.ownerIdx%len(trainerIDs)]
		if err := assignTrainerToGym(ctx, db, gymID, trainerID); err != nil {
			return err
		}

		for _, c := range classes {
			classID, maxCapacity, err := getOrCreateClass(ctx, db, gymID, c.name, c.maxCapacity)
			if err != nil {
				return err
			}

			// Seed 4 sessions for each class
			sessionTimes := []struct {
				startOffset time.Duration
				duration    time.Duration
				price       float64
			}{
				{-time.Hour, 2 * time.Hour, 10.0}, // Session starting 1 hour ago (for immediate testing)
				{time.Minute, time.Minute, 5.0},   // Session starting in 1 min and ending in 2 mins (for worker test)
				{24 * time.Hour, time.Hour, 15.0},
				{48 * time.Hour, 90 * time.Minute, 20.0},
				{72 * time.Hour, time.Hour, 18.0},
			}

			for _, st := range sessionTimes {
				// Ensure sessions start after May 23, 2026 (except the first one which is for today)
				var seedStartTime time.Time
				if st.startOffset < 0 {
					seedStartTime = time.Now().UTC()
				} else {
					seedStartTime = time.Date(2026, time.May, 23, 0, 0, 0, 0, time.UTC)
				}
				startTime := seedStartTime.Add(st.startOffset).Truncate(time.Hour)
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

func assignTrainerToGym(ctx context.Context, db *sql.DB, gymID, trainerID int) error {
	_, err := db.ExecContext(ctx, `INSERT INTO gym_trainers (gym_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, gymID, trainerID)
	if err != nil {
		return fmt.Errorf("assign trainer to gym: %w", err)
	}
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
