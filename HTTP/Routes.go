package HTTP

import (
	"backend/Functions"

	"github.com/gin-gonic/gin"
)

func Router(router *gin.Engine) {
	router.POST("/register", Functions.Register)
	router.POST("/login", Functions.Login)
	router.GET("/profile", Functions.GetProfile)
	router.POST("/changePassword", Functions.ChangePassword)

	router.GET("/post", Functions.GetPost)
	router.GET("/posts", Functions.GetAllPosts)
	router.POST("/post", Functions.CreatePost)
	router.DELETE("/post", Functions.DeletePost)
	router.POST("/post/like", Functions.LikePost)
	router.GET("/post/summarize", Functions.SummarizePost)

	router.POST("/comment", Functions.CreateComment)
	router.DELETE("/comment", Functions.DeleteComment)
	router.POST("/comment/like", Functions.LikeComment)

	router.GET("/lock_old_posts", Functions.LockOldPostsHandler)

	router.POST("/create_room", Functions.CreateRoom) // Create a new chatroom
	router.GET("/rooms", Functions.GetAllRooms)       // List all available chatrooms
	router.GET("/ws", Functions.HandleConnections)    // WebSocket for joining a specific chatroom

	router.POST("/add_tag", Functions.AddTag)
	router.GET("/tags/names", Functions.GetAllTagNames)

}
