package configuration

import (
	"fmt"
	"time"

	"github.com/standoffvenus/future/internal/appointment"
)

const (
	Table               string        = "appointments"
	LengthOfAppointment time.Duration = 30 * time.Minute
)

var (
	LocationPST   = mustParse("America/Los_Angeles")
	BusinessHours = appointment.BusinessHours{
		Location: LocationPST,
		Start:    am(8),
		End:      pm(5),
	}
)

func mustParse(s string) *time.Location {
	loc, err := time.LoadLocation(s)
	if err != nil {
		panic(fmt.Sprintf("configuration: could not parse time location %q", s))
	}

	return loc
}

func am(i int) int {
	return i
}

func pm(i int) int {
	// Hours are 0-indexed, hence the subtraction of 1
	return (i + 12) - 1
}
