package Functions

import (
	"backend/Mongo"
	"backend/Schemas"
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

func GetAllCommentsForPost(postId string) (comments []Schemas.Comment, err error) {
	comments = make([]Schemas.Comment, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := Mongo.GetCollection("melje_district").Find(ctx, bson.M{"post_id": postId})
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var comment Schemas.Comment
		err := cursor.Decode(&comment)
		if err == nil {
			comments = append(comments, comment)
		}
	}

	/*if err := cursor.Err(); err != nil {
		return
	}*/

	return
}

func CreateComment(c *gin.Context) {
	var comment Schemas.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	_, err := Mongo.GetCollection("melje_district").InsertOne(c, comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment added successfully"})
}

func DeleteComment(c *gin.Context) {
	commentId := c.Query("comment_id")
	if commentId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "comment_id is required"})
		return
	}

	objId, _ := primitive.ObjectIDFromHex(commentId)

	_, err := Mongo.GetCollection("melje_district").DeleteOne(c, bson.M{"_id": objId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error deleting comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}
