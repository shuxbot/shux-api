package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shuxbot/shux-api/domain"
)

type PostgresUserRepository struct {
	Pool *pgxpool.Pool
}

func (t *PostgresUserRepository) Get(UserId string, ServerId string) (domain.User, error) {
	ctx := context.Background()

	row := t.Pool.QueryRow(ctx, `
		SELECT u.description, u.points, u.warnings, u.tickets, u.flags,
		       COALESCE(
		           json_agg(json_build_object('reason', w.reason, 'date', w.occurred_at) ORDER BY w.occurred_at)
		           FILTER (WHERE w.id IS NOT NULL),
		           '[]'
		       )
		FROM users u
		LEFT JOIN user_warnings w ON w.server_id = u.server_id AND w.user_id = u.user_id
		WHERE u.server_id = $1 AND u.user_id = $2
		GROUP BY u.server_id, u.user_id
	`, ServerId, UserId)

	var u domain.User
	var warningsJSON []byte
	err := row.Scan(&u.Description, &u.Points, &u.Warnings, &u.Tickets, &u.Flags, &warningsJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, errors.New("Not found")
	} else if err != nil {
		return domain.User{}, err
	}

	if err := json.Unmarshal(warningsJSON, &u.WarningsRecord); err != nil {
		return domain.User{}, err
	}

	u.UserId = UserId

	return u, nil
}

func (t *PostgresUserRepository) Delete(UserId string, ServerId string) error {
	ctx := context.Background()

	tag, err := t.Pool.Exec(ctx, `DELETE FROM users WHERE server_id = $1 AND user_id = $2`, ServerId, UserId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("Not found")
	}

	return nil
}

// Update only touches fields the caller actually set, matching the previous
// Firestore MergeAll behavior driven by `omitempty` JSON tags: a zero value
// means "not provided", not "set to zero". Warnings are appended as new rows
// rather than overwriting existing ones, matching the old ArrayUnion call.
func (t *PostgresUserRepository) Update(u *domain.User, ServerId string) error {
	ctx := context.Background()

	sets := make([]string, 0, 5)
	args := []interface{}{ServerId, u.UserId}

	addSet := func(column string, value interface{}) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	if u.Description != "" {
		addSet("description", u.Description)
	}
	if u.Points != 0 {
		addSet("points", u.Points)
	}
	if u.Warnings != 0 {
		addSet("warnings", u.Warnings)
	}
	if u.Tickets != 0 {
		addSet("tickets", u.Tickets)
	}
	if u.Flags != 0 {
		addSet("flags", u.Flags)
	}

	if len(sets) > 0 {
		query := fmt.Sprintf(`UPDATE users SET %s WHERE server_id = $1 AND user_id = $2`, strings.Join(sets, ", "))
		if _, err := t.Pool.Exec(ctx, query, args...); err != nil {
			return err
		}
	}

	for _, warning := range u.WarningsRecord {
		if _, err := t.Pool.Exec(ctx, `
			INSERT INTO user_warnings (server_id, user_id, reason, occurred_at)
			VALUES ($1, $2, $3, $4)
		`, ServerId, u.UserId, warning.Reason, warning.Date); err != nil {
			return err
		}
	}

	return nil
}

func (t *PostgresUserRepository) Replace(u *domain.User, ServerId string) error {
	return t.upsert(u, ServerId)
}

func (t *PostgresUserRepository) Create(u *domain.User, ServerId string) error {
	return t.upsert(u, ServerId)
}

// Create and Replace are identical, mirroring the original Firestore
// repository where both ultimately called Set() (a full-document overwrite).
func (t *PostgresUserRepository) upsert(u *domain.User, ServerId string) error {
	ctx := context.Background()

	tx, err := t.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO users (server_id, user_id, description, points, warnings, tickets, flags)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (server_id, user_id) DO UPDATE SET
			description = excluded.description,
			points = excluded.points,
			warnings = excluded.warnings,
			tickets = excluded.tickets,
			flags = excluded.flags
	`, ServerId, u.UserId, u.Description, u.Points, u.Warnings, u.Tickets, u.Flags)
	if err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `DELETE FROM user_warnings WHERE server_id = $1 AND user_id = $2`, ServerId, u.UserId); err != nil {
		return err
	}

	for _, warning := range u.WarningsRecord {
		if _, err = tx.Exec(ctx, `
			INSERT INTO user_warnings (server_id, user_id, reason, occurred_at)
			VALUES ($1, $2, $3, $4)
		`, ServerId, u.UserId, warning.Reason, warning.Date); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Instantiates a new user repository
func NewPostgresUserRepo(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{Pool: pool}
}
