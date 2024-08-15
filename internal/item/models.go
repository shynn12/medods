package item

type Item struct {
	ID      string `json:"id" bson:"_id, omitempty"`
	Email   string `json:"email" bson:"email"`
	Refresh string `json:"refresh" bson:"refresh"`
}

type ItemDTO struct {
	Email   string `json:"email" bson:"email"`
	Refresh string `json:"refresh" bson:"refresh"`
}

type Tokens struct {
	Refresh string `json:"refresh"`
	Access  string `json:"access"`
}
