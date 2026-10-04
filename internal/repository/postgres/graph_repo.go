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
	return connections, rows.Err()
}

// GraphNode is a vertex of the institutional knowledge graph.
type GraphNode struct {
	ID         string                 `json:"id"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Label      string                 `json:"label"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// GraphEdge is a typed, weighted connection between two graph nodes.
type GraphEdge struct {
	SourceID         string  `json:"source_node_id"`
	TargetID         string  `json:"target_node_id"`
	RelationshipType string  `json:"relationship_type"`
	Weight           float64 `json:"weight"`
}

// SearchNodes finds nodes whose label matches the given text. An empty term
// returns the most recently created nodes so that callers always receive a
// usable slice instead of having to special-case an empty graph.
func (r *GraphRepository) SearchNodes(ctx context.Context, term string, limit int) ([]GraphNode, error) {
	if limit <= 0 {
		limit = 25
	}

	query := `
		SELECT id, entity_type, entity_id, label, metadata
		FROM graph_nodes
		WHERE $1 = '' OR label ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, term, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]GraphNode, 0, limit)
	for rows.Next() {
		var node GraphNode
		if err := rows.Scan(&node.ID, &node.EntityType, &node.EntityID, &node.Label, &node.Metadata); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

// EdgesForNodes returns every edge connecting any of the supplied node IDs.
func (r *GraphRepository) EdgesForNodes(ctx context.Context, nodeIDs []string) ([]GraphEdge, error) {
	if len(nodeIDs) == 0 {
		return []GraphEdge{}, nil
	}

	query := `
		SELECT source_node_id, target_node_id, relationship_type, weight
		FROM graph_edges
		WHERE source_node_id = ANY($1::uuid[])
		   OR target_node_id = ANY($1::uuid[])
		ORDER BY relationship_type
	`
	rows, err := r.db.Query(ctx, query, nodeIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	edges := make([]GraphEdge, 0, len(nodeIDs))
	for rows.Next() {
		var edge GraphEdge
		if err := rows.Scan(&edge.SourceID, &edge.TargetID, &edge.RelationshipType, &edge.Weight); err != nil {
			return nil, err
		}
		edges = append(edges, edge)
	}
	return edges, rows.Err()
}
