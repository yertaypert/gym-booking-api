package worker

import (
	"context"
	"log"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type Manager struct {
	sessionRepo usecase.SessionRepository
	stopChan    chan struct{}
}

func NewManager(sessionRepo usecase.SessionRepository) *Manager {
	return &Manager{
		sessionRepo: sessionRepo,
		stopChan:    make(chan struct{}),
	}
}

func (m *Manager) Start(ctx context.Context) {
	log.Println("Background workers manager starting...")

	// Run session worker
	go m.runSessionWorker(ctx)
}

func (m *Manager) Stop() {
	log.Println("Background workers manager stopping...")
	close(m.stopChan)
}

func (m *Manager) runSessionWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	m.updateSessions(ctx)

	for {
		select {
		case <-ticker.C:
			m.updateSessions(ctx)
		case <-m.stopChan:
			log.Println("Session worker stopped")
			return
		case <-ctx.Done():
			log.Println("Session worker stopped due to context cancellation")
			return
		}
	}
}

func (m *Manager) updateSessions(ctx context.Context) {
	rows, err := m.sessionRepo.UpdateExpiredSessions(ctx)
	if err != nil {
		log.Printf("Error updating expired sessions: %v", err)
		return
	}
	if rows > 0 {
		log.Printf("Successfully completed %d expired sessions", rows)
	}
}
