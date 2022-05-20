package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/standoffvenus/future/internal/application"
	"github.com/standoffvenus/future/internal/appointment"
	"github.com/standoffvenus/future/internal/configuration"
	"github.com/standoffvenus/future/internal/handler"
)

var (
	File = flag.String("file", "db.sqlite3", "Sets the file where the SQLite database is stored")
	Port = flag.Int("port", 8080, "Sets the port the server will run on")
)

func main() {
	flag.Parse()

	application.RunWithExit(func(ctx context.Context) error {
		if strings.TrimSpace(*File) == "" {
			return errors.New("no database file specified")
		}

		db, err := application.OpenSQLite3DB(ctx, *File)
		if err != nil {
			return err
		}

		repository := appointment.SQLRepository{
			Table:    configuration.Table,
			Database: db,
		}
		service := appointment.Service{
			Repository:          &repository,
			LengthOfAppointment: configuration.LengthOfAppointment,
			BusinessHours:       configuration.BusinessHours,
		}

		router := handler.NewRouter([]handler.Endpoint{
			{
				Path:    "/",
				Method:  http.MethodGet,
				Handler: handler.Health(),
			},
			{
				Path:    "/appointment",
				Method:  http.MethodPut,
				Handler: handler.CreateAppointment(&service),
			},
			{
				Path:    fmt.Sprintf("/appointment/trainer/:%s", handler.PathParameterTrainerID),
				Method:  http.MethodGet,
				Handler: handler.FindAppointmentsForTrainer(&service),
			},
		})

		return router.Serve(ctx, fmt.Sprintf(":%d", *Port))
	})
}
