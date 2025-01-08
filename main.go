package main

import (
	"backend/HTTP"
	"backend/Mongo"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Connect to MongoDB
	Mongo.ConnectToMongoDB()

	// Create a Gin router
	router := gin.Default()

	// Configure CORS for the frontend
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Set up HTTP routes
	HTTP.Router(router)

	// Start the server
	if err := router.Run("localhost:8080"); err != nil {
		panic(err)
	}
}
