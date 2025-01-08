package cronjobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"backend/Mongo"
	"backend/Schemas"

	"go.mongodb.org/mongo-driver/bson"
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

func LockOldPosts() {
	// Context for MongoDB operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch posts where locked is false or null
	query := bson.M{
		"$or": []bson.M{
			{"locked": bson.M{"$eq": false}},
			{"locked": bson.M{"$eq": nil}},
		},
	}

	cursor, err := Mongo.GetCollection("studenci_district").Find(ctx, query)
	if err != nil {
		log.Fatalf("Error fetching posts: %v", err)
	}
	defer cursor.Close(ctx)

	var posts []Schemas.Post
	for cursor.Next(ctx) {
		var post Schemas.Post
		if err := cursor.Decode(&post); err != nil {
			log.Printf("Error decoding post: %v", err)
			continue
		}

		// Fetch comments for each post
		comments, err := GetAllCommentsForPost(post.ID.Hex())
		if err != nil {
			log.Printf("Error fetching comments for post ID %s: %v", post.ID.Hex(), err)
			continue
		}

		post.Comments = comments
		posts = append(posts, post)
	}

	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error: %v", err)
	}

	// Filter posts based on comment date and update if necessary
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	layout := "2006-01-02T15:04:05Z07:00"

	for _, post := range posts {
		var newestComment time.Time
		for _, comment := range post.Comments {
			commentDate, err := time.Parse(layout, comment.Date)
			if err != nil {
				log.Printf("Error parsing date for comment in post ID %s: %v", post.ID.Hex(), err)
				continue
			}

			if commentDate.After(newestComment) {
				newestComment = commentDate
			}
		}

		// Check if the newest comment is older than 7 days
		if newestComment.Before(sevenDaysAgo) {
			// Update the `locked` field of the post to true
			update := bson.M{"$set": bson.M{"locked": true}}
			_, err := Mongo.GetCollection("studenci_district").UpdateOne(ctx, bson.M{"_id": post.ID}, update)
			if err != nil {
				log.Printf("Error updating post ID %s: %v", post.ID.Hex(), err)
				continue
			}

			fmt.Printf("Post ID: %s locked as the newest comment is older than 7 days\n", post.ID.Hex())
		}
	}

	fmt.Println("Job completed!")
}
