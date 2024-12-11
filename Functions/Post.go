package Functions

import (
	"backend/Mongo"
	"backend/Schemas"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
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

func CreatePost(c *gin.Context) {
	var post Schemas.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	_, err := Mongo.GetCollection("studenci_district").InsertOne(c, post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating post"})
		return
	}

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
