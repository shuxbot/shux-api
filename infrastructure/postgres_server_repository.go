package infrastructure

import (
	"context"
	"errors"

	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresServerRepository struct {
	Pool *pgxpool.Pool
}

func (t *PostgresServerRepository) List() ([]string, error) {
	ctx := context.Background()
	idArr := make([]string, 0)

	rows, err := t.Pool.Query(ctx, `SELECT id FROM servers`)
	if err != nil {
		return idArr, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		idArr = append(idArr, id)
	}

	return idArr, rows.Err()
}

func (t *PostgresServerRepository) GetRanking(ServerId string) ([]map[string]interface{}, error) {
	ctx := context.Background()

	rows, err := t.Pool.Query(ctx, `
		SELECT user_id, points FROM users
		WHERE server_id = $1
		ORDER BY points DESC
		LIMIT 5
	`, ServerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userArr []map[string]interface{}
	pos := 1
	for rows.Next() {
		var userId string
		var points float64
		if err := rows.Scan(&userId, &points); err != nil {
			return nil, err
		}

		userArr = append(userArr, map[string]interface{}{
			"id":     userId,
			"points": points,
			"rank":   pos,
		})
		pos++
	}

	return userArr, rows.Err()
}

// A single indexed query replaces the old full-collection scan (count every
// user with more points, using the (server_id, points) index) instead of
// fetching every higher-ranked document to the application just to count them.
func (t *PostgresServerRepository) GetUserRank(ServerId string, UserId string) (map[string]interface{}, error) {
	ctx := context.Background()

	var points float64
	var rank int
	err := t.Pool.QueryRow(ctx, `
		SELECT u.points, 1 + (
			SELECT COUNT(*) FROM users o
			WHERE o.server_id = u.server_id AND o.points > u.points
		) AS rank
		FROM users u
		WHERE u.server_id = $1 AND u.user_id = $2
	`, ServerId, UserId).Scan(&points, &rank)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("Not found")
	} else if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":     UserId,
		"points": points,
		"rank":   rank,
	}, nil
}

func (t *PostgresServerRepository) GetTickets(ServerId string) (map[string]interface{}, error) {
	ctx := context.Background()
	ticketsMap := map[string]interface{}{
		"tickets": make(map[string]interface{}),
	}

	var raw []byte
	err := t.Pool.QueryRow(ctx, `SELECT tickets FROM servers WHERE id = $1`, ServerId).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return ticketsMap, nil
	} else if err != nil {
		return nil, err
	}

	var tickets map[string]interface{}
	if err := json.Unmarshal(raw, &tickets); err != nil {
		return nil, err
	}
	ticketsMap["tickets"] = tickets

	return ticketsMap, nil
}

func (t *PostgresServerRepository) EditTickets(ServerId string, tickets map[string]interface{}) error {
	ctx := context.Background()

	raw, err := json.Marshal(tickets)
	if err != nil {
		return err
	}

	_, err = t.Pool.Exec(ctx, `
		INSERT INTO servers (id, tickets) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET tickets = excluded.tickets
	`, ServerId, raw)

	return err
}

func NewPostgresServerRepo(pool *pgxpool.Pool) *PostgresServerRepository {
	return &PostgresServerRepository{Pool: pool}
}
