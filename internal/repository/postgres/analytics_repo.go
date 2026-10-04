package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

type EventStats struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

func (r *AnalyticsRepository) GetEventStats(ctx context.Context, since time.Time) ([]EventStats, error) {
	query := `
		SELECT category, COUNT(*) as count 
		FROM events 
		WHERE start_time >= $1 
		GROUP BY category 
		ORDER BY count DESC
	`
	rows, err := r.db.Query(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []EventStats
	for rows.Next() {
		var s EventStats
		if err := rows.Scan(&s.Category, &s.Count); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}