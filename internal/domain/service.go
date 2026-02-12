package domain

import "time"

type Service struct {
	ID          string    `json:"id"`
	SalonID     string    `json:"salonId"`
	Name        string    `json:"name"`
	PriceKZT    int       `json:"priceKzt"`
	DurationMin int       `json:"durationMin"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}
