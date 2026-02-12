package domain

import "time"

type Salon struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	City               string    `json:"city"`
	MarketplaceEnabled bool      `json:"marketplaceEnabled"`
	CreatedAt          time.Time `json:"createdAt"`
}
