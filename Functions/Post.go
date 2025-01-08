package Functions

import (
	"backend/FunctionsHelper"
	"backend/Mongo"
	"backend/Schemas"
	"fmt"
	"log"
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

func SummarizePost(c *gin.Context) {
	postId := c.Query("post_id")
	if postId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "post_id is required"})
		return
	}

	objId, err := primitive.ObjectIDFromHex(postId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid post_id"})
		return
	}

	var post Schemas.Post
	err = Mongo.GetCollection("studenci_district").FindOne(c, bson.M{"_id": objId}).Decode(&post)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	// Fetch all comments for the post
	comments, err := GetAllCommentsForPost(post.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error decoding comments for post"})
		return
	}

	// Build a single string with post content and all comments
	contentToSummarize := fmt.Sprintf("Post: %s\n\nComments:\n", post.Problem)
	for _, comment := range comments {
		contentToSummarize += fmt.Sprintf("- %s: %s\n", comment.Username, comment.Description)
	}

	// Call AI service to summarize the content
	aiSummary, err := FunctionsHelper.CallAIService(contentToSummarize, 200, "Summarize the following post and its comments concisely:")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error summarizing content"})
		return
	}

	// Respond with the AI-generated summary
	c.JSON(http.StatusOK, gin.H{
		"post_id": post.ID.Hex(),
		"summary": aiSummary,
	})
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
	post.Date = time.Now().Format("2006-01-02")

	aiResponseV, err := FunctionsHelper.CallAIService(post.Problem, 1, "You are a bot that checks if the post is appropriate or not. By appropriate it is meant there are bad words. If it is appropriate return 1; else return 0.")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating AI response"})
		return
	}

	if aiResponseV == "0" { // Assuming aiResponse is a string; modify if it's a different type
		log.Println("AI Response not approved: AI returned 0")
		c.JSON(http.StatusForbidden, gin.H{"message": "Not approved by AI"}) // HTTP 403 Forbidden
		return
	}

	// Insert the post into the database
	insertResult, err := Mongo.GetCollection("studenci_district").InsertOne(c, post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating post"})
		return
	}

	// Generate an AI response for the post problem
	aiResponse, err := FunctionsHelper.CallAIService(post.Problem, 50, "You are an AI assistant for a Q&A site. Your purpose is to provide the first helpful and concise answer to users' questions. There is no followup. There is just your answer and it is not posible to ask for more information.")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating AI response"})
		return
	}
	log.Printf("AI Response: %s", aiResponse)

	postID, ok := insertResult.InsertedID.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving inserted post ID"})
		return
	}

	// Create a comment object with AI response
	comment := Schemas.Comment{
		Username:    "AI",
		Date:        time.Now().Format("2006-01-02"),
		Description: aiResponse,
		PostId:      postID.Hex(), // Use the post's ID as reference
	}

	// Insert the AI-generated comment into the comments collection
	_, err = Mongo.GetCollection("melje_district").InsertOne(c, comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error adding AI comment"})
		return
	}

	// Respond with success message
	c.JSON(http.StatusOK, gin.H{"message": "Post and AI comment added successfully"})
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
