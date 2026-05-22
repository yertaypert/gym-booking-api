package main

import (
	"database/sql"
	"log"
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

	password := "Secret12345"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	hp := string(hashedPassword)

	log.Println("Starting seeding...")

	// 1. Seed Users
	adminID, err := getOrCreateUser(db, "admin@email.com", "System Admin", "admin", 0, hp)
	if err != nil {
		log.Fatal(err)
	}

	ownerID, err := getOrCreateUser(db, "owner@email.com", "Gym Owner", "gym_owner", 0, hp)
	if err != nil {
		log.Fatal(err)
	}

	trainerUserID, err := getOrCreateUser(db, "trainer@email.com", "Senior Trainer", "trainer", 0, hp)
	if err != nil {
		log.Fatal(err)
	}

	customerID, err := getOrCreateUser(db, "user@email.com", "Regular Customer", "user", 1000.0, hp)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Seed Trainer Profile
	trainerID, err := getOrCreateTrainer(db, trainerUserID, "Yoga & Pilates", 25.0)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Users seeded: Admin(%d), Owner(%d), Trainer(%d), Customer(%d)", adminID, ownerID, trainerUserID, customerID)

	// 3. Seed Demo Data
	if err := seedDemoData(db, ownerID, trainerID); err != nil {
		log.Fatal(err)
	}

	log.Println("Seeding completed successfully!")
}

func getOrCreateUser(db *sql.DB, email, fullName, role string, balance float64, hp string) (int, error) {
	var id int
	err := db.QueryRow(`SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	if err == nil {
		_, err = db.Exec(`UPDATE users SET full_name = $1, role = $2, balance = $3, password_hash = $4 WHERE id = $5`, 
			fullName, role, balance, hp, id)
		return id, err
	}

	err = db.QueryRow(
		`INSERT INTO users (email, password_hash, full_name, role, balance)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		email, hp, fullName, role, balance,
	).Scan(&id)
	return id, err
}

func getOrCreateTrainer(db *sql.DB, userID int, specialization string, extraFee float64) (int, error) {
	var id int
	err := db.QueryRow(`SELECT id FROM trainers WHERE user_id = $1`, userID).Scan(&id)
	if err == nil {
		_, err = db.Exec(`UPDATE trainers SET specialization = $1, extra_fee = $2 WHERE id = $3`, 
			specialization, extraFee, id)
		return id, err
	}

	err = db.QueryRow(
		`INSERT INTO trainers (user_id, specialization, extra_fee)
		VALUES ($1, $2, $3) RETURNING id`,
		userID, specialization, extraFee,
	).Scan(&id)
	return id, err
}

func seedDemoData(db *sql.DB, ownerID int, trainerID int) error {
	gyms := []struct {
		name    string
		address string
		desc    string
	}{
		{"Elite Fitness", "123 Main St", "Premium gym in downtown"},
		{"Power House", "456 Oak Ave", "Hardcore bodybuilding center"},
	}

	for _, g := range gyms {
		gymID, err := getOrCreateGym(db, g.name, g.address, g.desc, ownerID)
		if err != nil {
			return err
		}

		// Assign trainer to gym
		_, _ = db.Exec(`INSERT INTO gym_trainers (gym_id, user_id) 
			SELECT $1, user_id FROM trainers WHERE id = $2
			ON CONFLICT DO NOTHING`, gymID, trainerID)

		classes := []struct {
			name        string
			maxCapacity int
		}{
			{"Yoga", 20},
			{"HIIT", 15},
		}

		for _, c := range classes {
			classID, err := getOrCreateClass(db, gymID, c.name, c.maxCapacity)
			if err != nil {
				return err
			}

			// Seed sessions
			sessionTimes := []struct {
				startOffset time.Duration
				price       float64
				withTrainer bool
			}{
				{24 * time.Hour, 15.0, false},
				{48 * time.Hour, 25.0, true},
			}

			for _, st := range sessionTimes {
				startTime := time.Now().UTC().Add(st.startOffset).Truncate(time.Hour)
				endTime := startTime.Add(time.Hour)

				var tid *int
				if st.withTrainer {
					tid = &trainerID
				}

				if err := getOrCreateSession(db, classID, startTime, endTime, c.maxCapacity, st.price, tid); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getOrCreateGym(db *sql.DB, name, address, description string, ownerID int) (int, error) {
	var id int
	err := db.QueryRow(`SELECT id FROM gyms WHERE name = $1`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	err = db.QueryRow(
		`INSERT INTO gyms (name, address, description, owner_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		name, address, description, ownerID,
	).Scan(&id)
	return id, err
}

func getOrCreateClass(db *sql.DB, gymID int, name string, maxCapacity int) (int, error) {
	var id int
	err := db.QueryRow(`SELECT id FROM classes WHERE gym_id = $1 AND name = $2`, gymID, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	err = db.QueryRow(
		`INSERT INTO classes (gym_id, name, max_capacity) VALUES ($1, $2, $3) RETURNING id`,
		gymID, name, maxCapacity,
	).Scan(&id)
	return id, err
}

func getOrCreateSession(db *sql.DB, classID int, startTime, endTime time.Time, availableSlots int, price float64, trainerID *int) error {
	var id int
	err := db.QueryRow(
		`SELECT id FROM class_sessions WHERE class_id = $1 AND start_time = $2`,
		classID, startTime,
	).Scan(&id)
	if err == nil {
		_, err = db.Exec(`UPDATE class_sessions SET trainer_id = $1, price = $2 WHERE id = $3`, trainerID, price, id)
		return err
	}

	_, err = db.Exec(
		`INSERT INTO class_sessions (class_id, start_time, end_time, available_slots, price, status, trainer_id)
		VALUES ($1, $2, $3, $4, $5, 'active', $6)`,
		classID, startTime, endTime, availableSlots, price, trainerID,
	)
	return err
}
