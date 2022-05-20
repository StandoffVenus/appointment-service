package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

func RunWithExit(fn func(context.Context) error) {
	var code int
	defer func() { os.Exit(code) }()

	if err := Run(fn); err != nil {
		log.
			Error().
			Err(err).
			Msg("Application failure.")

		code = 1
	}
}

func Run(fn func(context.Context) error) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := fn(ctx); err != nil {
		return err
	}

	return nil
}

func OpenSQLite3DB(ctx context.Context, file string) (*sql.DB, error) {
	connectionString := fmt.Sprintf("file:%s?cache=shared", file)
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
