package Functions

import (
	"backend/Mongo"
	"backend/Schemas"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func GetProfile(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Username required"})
		return
	}

	client := Mongo.GetMongoDB()
	var user Schemas.User
	err := client.Database("tezno_district").Collection("users").FindOne(c, bson.M{"username": username}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username": user.Name,
		"email":    user.Email,
	})
}

func Login(c *gin.Context) {
	var loginDetails struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginDetails); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	client := Mongo.GetMongoDB()
	var user Schemas.User
	err := client.Database("tezno_district").Collection("users").FindOne(c, bson.M{"username": loginDetails.Username}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginDetails.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user.Name,
		"id":      user.ID.Hex(),
	})
}

func Register(c *gin.Context) {
	var user Schemas.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error hashing password"})
		return
	}
	user.Password = string(hashedPassword)

	client := Mongo.GetMongoDB()
	_, err = client.Database("tezno_district").Collection("users").InsertOne(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error registering user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func ChangePassword(c *gin.Context) {
	var changePassword struct {
		Username    string `json:"username"`
		NewPassword string `json:"newPassword"`
	}

	if err := c.ShouldBindJSON(&changePassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	client := Mongo.GetMongoDB()
	var user Schemas.User
	err := client.Database("tezno_district").Collection("users").FindOne(c, bson.M{"username": changePassword.Username}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(changePassword.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error hashing password"})
		return
	}

	_, err = client.Database("tezno_district").Collection("users").UpdateOne(c, bson.M{"username": changePassword.Username}, bson.M{"$set": bson.M{"password": string(hashedPassword)}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error changing password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
