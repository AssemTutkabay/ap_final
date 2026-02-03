package domain

type Service struct {
	ID          string `json:"id"`
	SalonID     string `json:"salonId"`
	Name        string `json:"name"`
	Price       int64  `json:"price"`       // в тенге, без float
	DurationMin int    `json:"durationMin"` // важно
}
