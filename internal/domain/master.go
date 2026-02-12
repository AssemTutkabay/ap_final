package domain

import "time"

type Master struct {
	ID           string    `json:"id"`
	SalonID      string    `json:"salonId"`
	UserID       string    `json:"userId,omitempty"` // <-- добавили
	FullName     string    `json:"fullName"`
	Bio          string    `json:"bio,omitempty"`
	PortfolioURL string    `json:"portfolioUrl,omitempty"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
}
