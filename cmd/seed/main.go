package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/standoffvenus/future/internal/application"
	"github.com/standoffvenus/future/internal/configuration"
)

const Create = `
CREATE TABLE IF NOT EXISTS %s(
    id         TEXT PRIMARY KEY,
    trainer_id TEXT NOT NULL,
    user_id    TEXT NOT NULL,
    starts_at  INTEGER NOT NULL,
    ends_at    INTEGER NOT NULL
)
`

const Insert = `
INSERT OR REPLACE INTO %s(
	id,
	trainer_id,
	user_id,
	starts_at,
	ends_at
) VALUES (
	:id,
	:trainer_id,
	:user_id,
	:start,
	:end
)
`

var (
	JSONFile     = flag.String("json", "appointments.json", "Sets the path to the seeding JSON file")
	DatabaseFile = flag.String("db", "db.sqlite3", "Sets the SQLite3 database file to use")
)

func main() {
	flag.Parse()

	application.RunWithExit(func(ctx context.Context) error {
		db, err := application.OpenSQLite3DB(ctx, *DatabaseFile)
		if err != nil {
			return err
		}
		defer db.Close()

		txn, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer txn.Rollback()

		if _, err := txn.ExecContext(ctx, fmt.Sprintf(Create, configuration.Table)); err != nil {
			return err
		}

		fileBytes, err := fs.ReadFile(os.DirFS("."), *JSONFile)
		if err != nil {
			return err
		}

		var appointments []JSONAppointment
		if err := jsoniter.Unmarshal(fileBytes, &appointments); err != nil {
			return err
		}

		formattedInsert := fmt.Sprintf(Insert, configuration.Table)
		for _, apt := range appointments {
			_, err := txn.ExecContext(ctx, formattedInsert,
				sql.Named("id", toString[int](apt.ID)),
				sql.Named("trainer_id", toString[int](apt.TrainerID)),
				sql.Named("user_id", toString[int](apt.UserID)),
				sql.Named("start", apt.Start.Unix()),
				sql.Named("end", apt.End.Unix()))
			if err != nil {
				return err
			}
		}

		if err := txn.Commit(); err != nil {
			return err
		}

		return nil
	})
}

type JSONAppointment struct {
	End       time.Time `json:"ended_at"`
	ID        int       `json:"id"`
	TrainerID int       `json:"trainer_id"`
	UserID    int       `json:"user_id"`
	Start     time.Time `json:"started_at"`
}

func toString[I int | int32 | int64](v int) string { return strconv.FormatInt(int64(v), 10) }
