package Schemas

import "go.mongodb.org/mongo-driver/bson/primitive"

type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `json:"username" bson:"username"`
	Problem   string             `json:"problem" bson:"problem"`
	Date      string             `json:"date" bson:"date"`
	LikeCount int                `json:"likeCount" bson:"likeCount"`
	Comments  []Comment          `json:"comments"`
}
