package persistance

import (
	"context"
	_ "embed"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

var Pool *pgxpool.Pool
var ctx = context.Background()

// Connects to Postgres and applies the schema. All statements in schema.sql
// are idempotent (CREATE ... IF NOT EXISTS), so this is safe to run on every boot.
func init() {
	databaseURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatalln("DATABASE_URL environment variable not set")
	}

	var err error
	Pool, err = pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err = Pool.Exec(ctx, schema); err != nil {
		log.Fatalln(err)
	}
}
