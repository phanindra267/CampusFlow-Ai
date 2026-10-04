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

// CountEvents returns the number of events starting at or after `since`. It is
// paired with a second call to compute period-over-period trends.
func (r *AnalyticsRepository) CountEvents(ctx context.Context, since time.Time) (int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM events WHERE start_time >= $1`, since).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
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

	stats := make([]EventStats, 0, 8)
	for rows.Next() {
		var s EventStats
		if err := rows.Scan(&s.Category, &s.Count); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
