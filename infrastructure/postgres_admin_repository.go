package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shuxbot/shux-api/domain"
)

type PostgresAdminRepository struct {
	Pool *pgxpool.Pool
}

func (t *PostgresAdminRepository) Register(username string, password string) error {
	ctx := context.Background()

	_, err := t.Pool.Exec(ctx, `INSERT INTO admins (username, password) VALUES ($1, $2)`, username, password)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errors.New("Username already exists")
	}

	return err
}

func (t *PostgresAdminRepository) Login(username string) (domain.Admin, error) {
	ctx := context.Background()

	var admin domain.Admin
	err := t.Pool.QueryRow(ctx, `
		SELECT username, password FROM admins WHERE username = $1
	`, username).Scan(&admin.Username, &admin.Password)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Admin{}, errors.New("Username does not exist")
	} else if err != nil {
		return domain.Admin{}, err
	}

	return admin, nil
}

// Instantiates a new admin repository
func NewPostgresAdminRepo(pool *pgxpool.Pool) *PostgresAdminRepository {
	return &PostgresAdminRepository{Pool: pool}
}
