package Functions

import (
	"backend/Mongo"
	"backend/Schemas"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func LikeComment(c *gin.Context) {
	var requestBody struct {
		CommentID string `json:"comment_id"`
	}

	// Bind JSON body to the requestBody struct
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	// Validate comment_id
	if requestBody.CommentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "comment_id is required"})
		return
	}

	// Convert comment_id to ObjectID
	commentId, err := primitive.ObjectIDFromHex(requestBody.CommentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid comment_id"})
		return
	}

	// Access the collection
	collection := Mongo.GetCollection("melje_district") // Assuming "comments" collection, change if needed

	// Increment the LikeCount by 1
	update := bson.M{"$inc": bson.M{"likeCount": 1}}
	filter := bson.M{"_id": commentId}

	result, err := collection.UpdateOne(c, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update comment", "error": err.Error()})
		return
	}

	// Check if a document was updated
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Comment not found"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "Comment liked successfully"})
}
