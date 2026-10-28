package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shuxbot/shux-api/domain"
)

type PostgresRoleRepository struct {
	Pool *pgxpool.Pool
}

func (t *PostgresRoleRepository) List(ServerId string) ([]domain.Role, error) {
	ctx := context.Background()
	roleArr := make([]domain.Role, 0)

	rows, err := t.Pool.Query(ctx, `SELECT role_id, name, flags, level FROM roles WHERE server_id = $1`, ServerId)
	if err != nil {
		return roleArr, err
	}
	defer rows.Close()

	for rows.Next() {
		var r domain.Role
		if err := rows.Scan(&r.RoleId, &r.Name, &r.Flags, &r.Level); err != nil {
			return nil, err
		}
		roleArr = append(roleArr, r)
	}

	return roleArr, rows.Err()
}

func (t *PostgresRoleRepository) Get(RoleId string, ServerId string) (domain.Role, error) {
	ctx := context.Background()

	var r domain.Role
	err := t.Pool.QueryRow(ctx, `
		SELECT name, flags, level FROM roles WHERE server_id = $1 AND role_id = $2
	`, ServerId, RoleId).Scan(&r.Name, &r.Flags, &r.Level)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Role{}, errors.New("Not found")
	} else if err != nil {
		return domain.Role{}, err
	}

	r.RoleId = RoleId

	return r, nil
}

func (t *PostgresRoleRepository) Delete(RoleId string, ServerId string) error {
	ctx := context.Background()

	tag, err := t.Pool.Exec(ctx, `DELETE FROM roles WHERE server_id = $1 AND role_id = $2`, ServerId, RoleId)
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
func (t *PostgresRoleRepository) Update(r *domain.Role, ServerId string) error {
	ctx := context.Background()

	sets := make([]string, 0, 3)
	args := []interface{}{ServerId, r.RoleId}

	addSet := func(column string, value interface{}) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	if r.Name != "" {
		addSet("name", r.Name)
	}
	if r.Flags != 0 {
		addSet("flags", r.Flags)
	}
	if r.Level != 0 {
		addSet("level", r.Level)
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf(`UPDATE roles SET %s WHERE server_id = $1 AND role_id = $2`, strings.Join(sets, ", "))
	_, err := t.Pool.Exec(ctx, query, args...)

	return err
}

func (t *PostgresRoleRepository) Replace(r *domain.Role, ServerId string) error {
	return t.upsert(r, ServerId)
}

func (t *PostgresRoleRepository) Create(r *domain.Role, ServerId string) error {
	return t.upsert(r, ServerId)
}

// Create and Replace are identical, mirroring the original Firestore
// repository where both ultimately called Set() (a full-document overwrite).
func (t *PostgresRoleRepository) upsert(r *domain.Role, ServerId string) error {
	ctx := context.Background()

	_, err := t.Pool.Exec(ctx, `
		INSERT INTO roles (server_id, role_id, name, flags, level)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (server_id, role_id) DO UPDATE SET
			name = excluded.name,
			flags = excluded.flags,
			level = excluded.level
	`, ServerId, r.RoleId, r.Name, r.Flags, r.Level)

	return err
}

func NewPostgresRoleRepo(pool *pgxpool.Pool) *PostgresRoleRepository {
	return &PostgresRoleRepository{Pool: pool}
}
