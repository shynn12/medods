package item

import "time"

type Item struct {
	ID        string    `json:"id" bson:"_id, omitempty"`
	Email     string    `json:"email" bson:"email"`
	Refresh   string    `json:"refresh" bson:"refresh"`
	ExpiresAt time.Time `json:"exp"`
}

type ItemDTO struct {
	Email     string    `json:"email" bson:"email"`
	Refresh   string    `json:"refresh" bson:"refresh"`
	ExpiresAt time.Time `json:"exp"`
}

type Tokens struct {
	Refresh string `json:"refresh"`
	Access  string `json:"access"`
}
