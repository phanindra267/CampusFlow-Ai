package postgres

import (
	"context"

	"github.com/campuscare/api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, e *domain.Event) error {
	query := `INSERT INTO events (id, title, description, category, organizer_id, capacity, start_time, end_time, registration_deadline) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query, e.ID, e.Title, e.Description, e.Category, e.OrganizerID, e.Capacity, e.StartTime, e.EndTime, e.RegistrationDeadline)
	return err
}

func (r *EventRepository) List(ctx context.Context) ([]domain.Event, error) {
	query := `SELECT id, title, description, category, organizer_id, capacity FROM events ORDER BY start_time DESC LIMIT 50`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.Category, &e.OrganizerID, &e.Capacity); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}