package Functions

import (
	"backend/Mongo"
	"backend/Schemas"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetPost(c *gin.Context) {
	postId := c.Query("post_id")
	if postId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "post_id is required"})
		return
	}

	objId, _ := primitive.ObjectIDFromHex(postId)

	var post Schemas.Post
	err := Mongo.GetCollection("studenci_district").FindOne(c, bson.M{"_id": objId}).Decode(&post)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	comments, err := GetAllCommentsForPost(post.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error decoding comments for post"})
		return
	}

	post.Comments = comments

	c.JSON(http.StatusOK, post)
}

func GetAllPosts(c *gin.Context) {
	var posts = make([]Schemas.Post, 0)

	cursor, err := Mongo.GetCollection("studenci_district").Find(c, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving posts"})
		return
	}
	defer cursor.Close(c)

	for cursor.Next(c) {
		var post Schemas.Post
		if err := cursor.Decode(&post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error decoding post"})
			return
		}

		comments, err := GetAllCommentsForPost(post.ID.Hex())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error decoding comments for post"})
			return
		}

		post.Comments = comments
		posts = append(posts, post)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Cursor error"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// CreatePost handles the creation of a new post
func CreatePost(c *gin.Context) {
	var post Schemas.Post

	// Bind the JSON body to the post struct
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Validate that username and problem are not empty
	if post.Username == "" || post.Problem == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Username and problem cannot be empty"})
		return
	}

	if len(post.Username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Username cannot exceed 50 characters"})
		return
	}

	// Check the maximum length of the problem description (e.g., 500 characters)
	if len(post.Problem) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Problem description cannot exceed 500 characters"})
		return
	}

	// Set the current date automatically on the backend
	// The current date in "YYYY-MM-DD" format
	post.Date = time.Now().Format("2006-01-02")

	// Insert the post into the database
	_, err := Mongo.GetCollection("studenci_district").InsertOne(c, post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating post"})
		return
	}

	// Respond with success message
	c.JSON(http.StatusOK, gin.H{"message": "Post added successfully"})
}

func DeletePost(c *gin.Context) {
	postId := c.Query("post_id")
	if postId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "post_id is required"})
		return
	}

	objId, _ := primitive.ObjectIDFromHex(postId)

	_, err := Mongo.GetCollection("studenci_district").DeleteOne(c, bson.M{"_id": objId})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
func LikePost(c *gin.Context) {
	var requestBody struct {
		PostID string `json:"post_id"`
	}

	// Bind JSON body to the requestBody struct
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	// Validate post_id
	if requestBody.PostID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "post_id is required"})
		return
	}

	// Convert post_id to ObjectID
	objId, err := primitive.ObjectIDFromHex(requestBody.PostID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid post_id"})
		return
	}

	// Access the collection
	collection := Mongo.GetCollection("studenci_district")

	// Increment the LikeCount by 1
	update := bson.M{"$inc": bson.M{"likeCount": 1}}
	filter := bson.M{"_id": objId}

	result, err := collection.UpdateOne(c, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update post", "error": err.Error()})
		return
	}

	// Check if a document was updated
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	// Debugging: Log the update result
	fmt.Printf("Update result: %+v\n", result)

	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "Post liked successfully"})
}
