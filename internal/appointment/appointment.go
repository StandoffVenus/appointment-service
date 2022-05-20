package appointment

import "time"

type Appointment struct {
	ID        string
	TrainerID string
	UserID    string
	Start     time.Time
	End       time.Time
}
