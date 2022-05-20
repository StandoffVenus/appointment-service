package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/standoffvenus/future/internal/appointment"
	"github.com/standoffvenus/future/internal/empty"
)

const (
	PathParameterTrainerID = "trainer_id"
	QueryParameterStart    = "starts_at"
	QueryParameterEnd      = "ends_at"
)

var ErrNotATime = errors.New("expected an RFC3339 string or Unix timestamp")

type AppointmentDTO struct {
	ID        string    `json:"id"`
	TrainerID string    `json:"trainer_id"`
	UserID    string    `json:"user_id"`
	Start     time.Time `json:"starts_at"`
	End       time.Time `json:"ends_at"`
}

type AppointmentService interface {
	Create(ctx context.Context, apt appointment.Appointment) error
	FindByTrainerID(ctx context.Context, trainerID string) ([]appointment.Appointment, error)
	FindByTrainerIDInRange(ctx context.Context, trainerID string, timeRange appointment.Range) ([]appointment.Appointment, error)
}

func Health() Handler {
	return func(r Request) (Response, error) {
		return NoContent(), nil
	}
}

func CreateAppointment(svc AppointmentService) Handler {
	return func(r Request) (Response, error) {
		dto, err := ParseBody[AppointmentDTO](r)
		if err != nil {
			return BadRequest("invalid appointment body"), nil
		}

		apt, err := EnsureValidAppointment(dto)
		if err != nil {
			return BadRequest(err.Error()), nil
		}

		if err := svc.Create(r.Context, apt); err != nil {
			switch {
			case errors.Is(err, appointment.ErrInvalidDateRange),
				errors.Is(err, appointment.ErrOutsideBusinessHours):
				return BadRequest(err.Error()), nil
			case errors.Is(err, appointment.ErrScheduleConflict),
				errors.Is(err, appointment.ErrIDTaken):
				return Conflict(err.Error()), nil
			default:
				return Response{}, err
			}
		}

		return NoContent(), nil
	}
}

func FindAppointmentsForTrainer(svc AppointmentService) Handler {
	return func(r Request) (Response, error) {
		trainerID, ok := r.PathParameters[PathParameterTrainerID]
		if !ok {
			return BadRequest("no trainer ID provided"), nil
		}

		if r.QueryParameters.Has(QueryParameterStart) || r.QueryParameters.Has(QueryParameterEnd) {
			return findAppointmentsForTrainerInRange(r, svc, trainerID)
		}

		return findAppointmentsByTrainerID(r.Context, svc, trainerID)
	}
}

func EnsureValidAppointment(dto AppointmentDTO) (appointment.Appointment, error) {
	if empty.String(dto.TrainerID) {
		return appointment.Appointment{}, errors.New("trainer is required")
	}

	if empty.String(dto.UserID) {
		return appointment.Appointment{}, errors.New("user is required")
	}

	if dto.Start.IsZero() || dto.End.IsZero() {
		return appointment.Appointment{}, errors.New("time range is required")
	}

	id := dto.ID
	if empty.String(id) {
		id = uuid.NewString()
	}

	return appointment.Appointment{
		ID:        id,
		TrainerID: dto.TrainerID,
		UserID:    dto.UserID,
		Start:     dto.Start,
		End:       dto.End,
	}, nil
}

func findAppointmentsForTrainerInRange(
	r Request,
	svc AppointmentService,
	trainerID string,
) (Response, error) {
	start, err := parseTime(r.QueryParameters.Get(QueryParameterStart))
	if err != nil {
		return BadRequest(fmt.Sprintf("bad start time - %s", err)), nil
	}

	end, err := parseTime(r.QueryParameters.Get(QueryParameterEnd))
	if err != nil {
		return BadRequest(fmt.Sprintf("bad end time - %s", err)), nil
	}

	timeRange := appointment.Range{
		Start: start,
		End:   end,
	}
	apts, err := svc.FindByTrainerIDInRange(r.Context, trainerID, timeRange)
	if err != nil {
		switch {
		case errors.Is(err, appointment.ErrNoTrainerID),
			errors.Is(err, appointment.ErrInvalidDateRange):
			return BadRequest(err.Error()), nil
		}

		return Response{}, err
	}

	return OK(apts), nil

}

func findAppointmentsByTrainerID(ctx context.Context, svc AppointmentService, trainerID string) (Response, error) {
	apts, err := svc.FindByTrainerID(ctx, trainerID)
	if err != nil {
		switch {
		case errors.Is(err, appointment.ErrNoTrainerID):
			return BadRequest(err.Error()), nil
		}

		return Response{}, err
	}

	return OK(apts), nil
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}

	timestamp, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return time.Unix(timestamp, 0), nil
	}

	return time.Time{}, ErrNotATime
}
