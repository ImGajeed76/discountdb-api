package jobs

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type ScoreUpdater struct {
	db        *sql.DB
	batchSize int
	interval  time.Duration
	done      chan bool
}

func NewScoreUpdater(db *sql.DB, batchSize int, interval time.Duration) *ScoreUpdater {
	return &ScoreUpdater{
		db:        db,
		batchSize: batchSize,
		interval:  interval,
		done:      make(chan bool),
	}
}

func (s *ScoreUpdater) updateScores() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get total count of coupons that need updating
	var totalToUpdate int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM coupons 
		WHERE last_score_update IS NULL 
		OR last_score_update < CURRENT_TIMESTAMP - INTERVAL '1 hour'
	`).Scan(&totalToUpdate)
	if err != nil {
		return err
	}

	// No coupons need updating
	if totalToUpdate == 0 {
		return nil
	}

	// Calculate number of batches needed
	batches := (totalToUpdate + s.batchSize - 1) / s.batchSize

	// Process all batches
	for i := 0; i < batches; i++ {
		_, err := s.db.ExecContext(ctx, "SELECT update_materialized_scores_batch($1)", s.batchSize)
		if err != nil {
			return err
		}

		// Small sleep between batches to reduce database load
		time.Sleep(100 * time.Millisecond)
	}

	// Log success
	log.Printf("Updated scores for %d coupons", totalToUpdate)

	return nil
}

func (s *ScoreUpdater) Start() {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.updateScores(); err != nil {
					log.Printf("Error updating scores: %v", err)
				}
			case <-s.done:
				return
			}
		}
	}()
}

func (s *ScoreUpdater) Stop() {
	s.done <- true
}
