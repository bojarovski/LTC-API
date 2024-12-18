package Schemas

import "go.mongodb.org/mongo-driver/bson/primitive"

type Comment struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	PostId      string             `json:"post_id" bson:"post_id"`
	Username    string             `json:"username" bson:"username"`
	Description string             `json:"description" bson:"description"`
	Date        string             `json:"date" bson:"date"`
	LikeCount   int                `json:"likeCount" bson:"likeCount"`
}
