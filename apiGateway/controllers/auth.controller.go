package controllers

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/helpers"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/dto"
)


func Register(c *gin.Context) {
	var data map[string]interface{}

	if err := c.ShouldBindJSON(&data); err != nil {
		log.Println("Unable to parse body:", err)
		helpers.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	password, passwordOk := data["password"].(string)
	if !passwordOk || len(password) <= 6 {
		helpers.SendError(c, http.StatusBadRequest, "Password must be greater than 6 characters")
		return
	}

	email, emailOk := data["email"].(string)
	if !emailOk || !validateEmail(strings.TrimSpace(email)) {
		helpers.SendError(c, http.StatusBadRequest, "Invalid email address")
		return
	}

	var existingUser models.User
	err := usersCollection.FindOne(context.Background(), bson.M{"email": strings.TrimSpace(email)}).Decode(&existingUser)
	if err != mongo.ErrNoDocuments {
		if err != nil {
			log.Println("Database error:", err)
			helpers.SendError(c, http.StatusInternalServerError, "Failed to check email")
			return
		}
		helpers.SendError(c, http.StatusBadRequest, "Email already exists")
		return
	}

	user := models.User{
		FirstName: data["firstName"].(string),
		LastName:  data["lastName"].(string),
		Phone:     data["phone"].(string),
		Email:     strings.TrimSpace(email),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user.SetPassword(password)

	insertResult, err := usersCollection.InsertOne(context.Background(), user)
	if err != nil {
		log.Println("Database error:", err)
		helpers.SendError(c, http.StatusInternalServerError, "Failed to create account")
		return
	}

	user.ID = insertResult.InsertedID.(primitive.ObjectID)

	helpers.SendJSON(c, http.StatusCreated, gin.H{
		"data":    user,
		"message": "Account created successfully",
	})
}

func Login(c *gin.Context) {
	var data map[string]string

	if err := c.ShouldBindJSON(&data); err != nil {
		log.Println("Unable to parse body:", err)
		helpers.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	email, emailOk := data["email"]
	password, passwordOk := data["password"]
	if !emailOk || !passwordOk {
		helpers.SendError(c, http.StatusBadRequest, "Email and password are required")
		return
	}

	email = strings.TrimSpace(email)
	filter := bson.M{"email": email}

	var user models.User
	err := usersCollection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			helpers.SendError(c, http.StatusUnauthorized, "Incorrect email address or password")
			return
		}
		log.Println("Database error:", err)
		helpers.SendError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if err := user.ComparePassword(password); err != nil {
		helpers.SendError(c, http.StatusUnauthorized, "Incorrect email address or password")
		return
	}

	token, err := helpers.GenerateJwt(user.ID.Hex())
	if err != nil {
		log.Println("Token generation error:", err)
		helpers.SendError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	maxAge := int(time.Hour * 24 / time.Second)
	c.SetCookie("jwt", token, maxAge, "/", configs.Env.APP_URL, configs.Env.MODE == "production", true)

	helpers.SendJSON(c, http.StatusOK, gin.H{
		"data":    user,
		"message": "Logged in successfully",
	})
}

func Logout(c *gin.Context) {

	maxAge := -1 * int(time.Hour*24/time.Second)

	c.SetCookie("jwt", "", maxAge, "/", configs.Env.APP_URL, configs.Env.MODE == "production", true)

	helpers.SendJSON(c, http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
