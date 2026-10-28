package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shuxbot/shux-api/domain"
)

type PostgresChannelRepository struct {
	Pool *pgxpool.Pool
}

func (t *PostgresChannelRepository) List(ServerId string) ([]domain.Channel, error) {
	ctx := context.Background()
	channelArr := make([]domain.Channel, 0)

	rows, err := t.Pool.Query(ctx, `SELECT channel_id, flags FROM channels WHERE server_id = $1`, ServerId)
	if err != nil {
		return channelArr, err
	}
	defer rows.Close()

	for rows.Next() {
		var c domain.Channel
		if err := rows.Scan(&c.ChannelId, &c.Flags); err != nil {
			return nil, err
		}
		channelArr = append(channelArr, c)
	}

	return channelArr, rows.Err()
}

func (t *PostgresChannelRepository) Get(ChannelId string, ServerId string) (domain.Channel, error) {
	ctx := context.Background()

	var c domain.Channel
	err := t.Pool.QueryRow(ctx, `
		SELECT flags FROM channels WHERE server_id = $1 AND channel_id = $2
	`, ServerId, ChannelId).Scan(&c.Flags)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Channel{}, errors.New("Not found")
	} else if err != nil {
		return domain.Channel{}, err
	}

	c.ChannelId = ChannelId

	return c, nil
}

func (t *PostgresChannelRepository) Delete(ChannelId string, ServerId string) error {
	ctx := context.Background()

	tag, err := t.Pool.Exec(ctx, `DELETE FROM channels WHERE server_id = $1 AND channel_id = $2`, ServerId, ChannelId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("Not found")
	}

	return nil
}

// Only touches fields the caller actually set (zero value means "not
// provided"), matching the previous Firestore MergeAll + omitempty behavior.
func (t *PostgresChannelRepository) Update(c *domain.Channel, ServerId string) error {
	if c.Flags == 0 {
		return nil
	}

	ctx := context.Background()
	_, err := t.Pool.Exec(ctx, `
		UPDATE channels SET flags = $1 WHERE server_id = $2 AND channel_id = $3
	`, c.Flags, ServerId, c.ChannelId)

	return err
}

func (t *PostgresChannelRepository) Replace(c *domain.Channel, ServerId string) error {
	return t.upsert(c, ServerId)
}

func (t *PostgresChannelRepository) Create(c *domain.Channel, ServerId string) error {
	return t.upsert(c, ServerId)
}

// Create and Replace are identical, mirroring the original Firestore
// repository where both ultimately called Set() (a full-document overwrite).
func (t *PostgresChannelRepository) upsert(c *domain.Channel, ServerId string) error {
	ctx := context.Background()

	_, err := t.Pool.Exec(ctx, `
		INSERT INTO channels (server_id, channel_id, flags)
		VALUES ($1, $2, $3)
		ON CONFLICT (server_id, channel_id) DO UPDATE SET flags = excluded.flags
	`, ServerId, c.ChannelId, c.Flags)

	return err
}

func NewPostgresChannelRepo(pool *pgxpool.Pool) *PostgresChannelRepository {
	return &PostgresChannelRepository{Pool: pool}
}
