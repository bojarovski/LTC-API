package main

import (
	"backend/HTTP"
	"backend/Mongo"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	//var endpointRouter = HTTP.Routes{} // Inicializacija router-jev za endpoint-e

	Mongo.ConnectToMongoDB() // Vzpostavitev povezave s podatkovno bazo MongoDB
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	HTTP.Router(router)

	if err := router.Run("localhost:8080"); err != nil {
		panic(err)
	}

}
