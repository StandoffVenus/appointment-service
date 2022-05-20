package appointment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
)

type Repository interface {
	GetByTrainer(context.Context, string) ([]Appointment, error)
	GetByTrainerAndDate(context.Context, string, Range) ([]Appointment, error)
	Create(context.Context, Appointment) error
}

type SQLRepository struct {
	Database *sql.DB
	Table    string
}

type Range struct {
	Start time.Time
	End   time.Time
}

type entity struct {
	ID        string
	TrainerID string
	UserID    string
	Start     int64
	End       int64
}

var _ Repository = new(SQLRepository)

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *SQLRepository) GetByTrainer(ctx context.Context, trainerID string) ([]Appointment, error) {
	const Query = `
SELECT id, trainer_id, user_id, starts_at, ends_at
  FROM %s
 WHERE trainer_id = :trainer_id
`

	formattedQuery := fmt.Sprintf(Query, r.Table)
	rows, err := r.Database.QueryContext(ctx, formattedQuery, sql.Named("trainer_id", trainerID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAll(rows)
}

func (r *SQLRepository) GetByTrainerAndDate(ctx context.Context, trainerID string, times Range) ([]Appointment, error) {
	const Query = `
SELECT id, trainer_id, user_id, starts_at, ends_at
  FROM %s
 WHERE trainer_id = :trainer_id
   AND starts_at >= :start
   AND ends_at <= :end
`

	formattedQuery := fmt.Sprintf(Query, r.Table)
	rows, err := r.Database.QueryContext(ctx, formattedQuery,
		sql.Named("trainer_id", trainerID),
		sql.Named("start", times.Start.Unix()),
		sql.Named("end", times.End.Unix()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAll(rows)
}

func (r *SQLRepository) Create(ctx context.Context, apt Appointment) error {
	txn, err := r.Database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer txn.Rollback()

	if n, err := r.countAppointments(ctx, txn, apt); err != nil {
		return err
	} else if n > 0 {
		return ErrScheduleConflict
	}

	const Insert = `
INSERT INTO %s(id, trainer_id, user_id, starts_at, ends_at)
	 VALUES (:id, :trainer_id, :user_id, :start, :end)
`

	formattedInsert := fmt.Sprintf(Insert, r.Table)
	_, err = txn.ExecContext(ctx, formattedInsert,
		sql.Named("id", apt.ID),
		sql.Named("trainer_id", apt.TrainerID),
		sql.Named("user_id", apt.UserID),
		sql.Named("start", apt.Start.Unix()),
		sql.Named("end", apt.End.Unix()))
	if err != nil {
		// TODO: This is not portable to other SQL DB's.
		if isSQLiteError(err, sqlite3.ErrConstraintPrimaryKey, sqlite3.ErrConstraintUnique) {
			return ErrIDTaken
		}

		return err
	}

	return txn.Commit()
}

func (r *SQLRepository) countAppointments(ctx context.Context, txn *sql.Tx, apt Appointment) (int64, error) {
	const Query = `
SELECT COUNT(*) AS c
  FROM %s
 WHERE trainer_id = :trainer_id
   AND starts_at >= :start
   AND ends_at <= :end
`

	formattedQuery := fmt.Sprintf(Query, r.Table)
	row := txn.QueryRowContext(ctx, formattedQuery,
		sql.Named("trainer_id", apt.TrainerID),
		sql.Named("start", apt.Start.Unix()),
		sql.Named("end", apt.End.Unix()))

	var count int64
	if err := row.Scan(&count); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	return count, nil
}

func scanAll(rows *sql.Rows) ([]Appointment, error) {
	apts := make([]Appointment, 0, 16)
	for rows.Next() {
		apt, err := scanRow(rows)
		if err != nil {
			return nil, err
		}

		apts = append(apts, apt)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return apts, nil
}

func isSQLiteError(err error, codes ...sqlite3.ErrNoExtended) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		for _, c := range codes {
			if sqliteErr.ExtendedCode == c {
				return true
			}
		}
	}

	return false
}

func scanRow(r rowScanner) (Appointment, error) {
	var ent entity
	err := r.Scan(
		&ent.ID,
		&ent.TrainerID,
		&ent.UserID,
		&ent.Start,
		&ent.End)

	return entityToAppointment(ent), err
}

func entityToAppointment(ent entity) Appointment {
	return Appointment{
		ID:        ent.ID,
		TrainerID: ent.TrainerID,
		UserID:    ent.UserID,
		Start:     time.Unix(ent.Start, 0),
		End:       time.Unix(ent.End, 0),
	}
}
