package Schemas

type Tag struct {
	ID        string `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string `json:"name" bson:"name"`
	DateAdded string `json:"date_added" bson:"date_added"`
}
