package appointment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/standoffvenus/future/internal/empty"
)

var loc = time.UTC

var (
	ErrIDTaken              = errors.New("an appointment with the given ID already exists")
	ErrInvalidDateRange     = errors.New("supplied times are invalid")
	ErrOutsideBusinessHours = errors.New("proposed time outside business hours")
	ErrScheduleConflict     = errors.New("time not available")
	ErrNoTrainerID          = errors.New("no trainer ID supplied")
)

type BusinessHours struct {
	Location *time.Location
	Start    int
	End      int
}

type Service struct {
	Repository          Repository
	LengthOfAppointment time.Duration
	BusinessHours
}

func (s *Service) Create(ctx context.Context, apt Appointment) error {
	if err := s.ensureValidCreateTimes(apt.Start, apt.End); err != nil {
		return err
	}

	if err := s.Repository.Create(ctx, apt); err != nil {
		return err
	}

	return nil
}

func (s *Service) FindByTrainerIDInRange(ctx context.Context, trainerID string, timeRange Range) ([]Appointment, error) {
	if empty.String(trainerID) {
		return []Appointment{}, ErrNoTrainerID
	}

	if err := s.ensureValidGetTimes(timeRange.Start, timeRange.End); err != nil {
		return []Appointment{}, err
	}

	apts, err := s.Repository.GetByTrainerAndDate(ctx, trainerID, timeRange)
	if err != nil {
		return []Appointment{}, err
	}

	return apts, nil
}

func (s *Service) FindByTrainerID(ctx context.Context, trainerID string) ([]Appointment, error) {
	if empty.String(trainerID) {
		return []Appointment{}, ErrNoTrainerID
	}

	apts, err := s.Repository.GetByTrainer(ctx, trainerID)
	if err != nil {
		return []Appointment{}, err
	}

	return apts, nil
}

func (s *Service) ensureValidCreateTimes(start, end time.Time) error {
	if err := s.ensureValidTimes(start, end); err != nil {
		return err
	}

	if expectedEnd := start.Add(s.LengthOfAppointment); !expectedEnd.Equal(end) {
		log.
			Debug().
			Fields(map[string]any{
				"starts_at":        start,
				"ends_at":          end,
				"expected_ends_at": expectedEnd,
			}).
			Msg("Invalid appointment length from consumer.")

		return fmt.Errorf("%w: invalid appointment length (must be %s)", ErrInvalidDateRange, s.LengthOfAppointment)
	}

	if start.Minute()%30 != 0 {
		return fmt.Errorf("%w: appointment must be scheduled on :00 or :30", ErrInvalidDateRange)
	}

	if start.Before(time.Now()) {
		return fmt.Errorf("%w: appointment for the past", ErrInvalidDateRange)
	}

	if start.In(s.BusinessHours.Location).Hour() < s.BusinessHours.Start ||
		end.In(s.BusinessHours.Location).Hour() > s.BusinessHours.End {
		return ErrOutsideBusinessHours
	}

	return nil
}

func (s *Service) ensureValidGetTimes(start, end time.Time) error {
	if err := s.ensureValidTimes(start, end); err != nil {
		return err
	}

	return nil
}

func (s *Service) ensureValidTimes(start, end time.Time) error {
	if start.IsZero() {
		return fmt.Errorf("%w: no start time", ErrInvalidDateRange)
	}

	if end.IsZero() {
		return fmt.Errorf("%w: no end time", ErrInvalidDateRange)
	}

	return nil
}
