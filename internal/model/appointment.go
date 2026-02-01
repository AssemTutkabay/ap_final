package model

import "time"

type Appointment struct {
	ID        int
	ClientID  int
	MasterID  int
	ServiceID int
	StartTime time.Time
	EndTime   time.Time
	Status    string // booked/cancelled/done
}
