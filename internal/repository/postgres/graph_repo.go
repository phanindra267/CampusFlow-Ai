package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GraphRepository struct {
	db *pgxpool.Pool
}

func NewGraphRepository(db *pgxpool.Pool) *GraphRepository {
	return &GraphRepository{db: db}
}

// GetUserNetwork uses a recursive CTE to find connected users (e.g. friends of friends)
func (r *GraphRepository) GetUserNetwork(ctx context.Context, userID string, depth int) ([]string, error) {
	query := `
	WITH RECURSIVE network AS (
		SELECT friend_id, 1 as current_depth
		FROM user_connections
		WHERE user_id = $1
		
		UNION
		
		SELECT uc.friend_id, n.current_depth + 1
		FROM user_connections uc
		INNER JOIN network n ON n.friend_id = uc.user_id
		WHERE n.current_depth < $2
	)
	SELECT DISTINCT friend_id FROM network WHERE friend_id != $1;
	`
	rows, err := r.db.Query(ctx, query, userID, depth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		connections = append(connections, id)
	}
	return connections, nil
}