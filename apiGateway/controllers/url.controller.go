package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/manlikehenryy/url-shortener-go/configs"
	"github.com/manlikehenryy/url-shortener-go/database"
	"github.com/manlikehenryy/url-shortener-go/helpers"
	"github.com/manlikehenryy/url-shortener-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ctx = context.Background()

func CreateUrl(c *gin.Context) {
	var url models.Url

	if err := c.ShouldBindJSON(&url); err != nil {
		log.Println("Unable to parse body:", err)
		helpers.SendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	userId, ok := c.MustGet("userId").(primitive.ObjectID)
	if !ok {
		helpers.SendError(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	url.UserId = userId

	shortURL := helpers.GenerateShortURL(url.OriginalUrl)
	expiration := time.Duration(url.Expiration) * time.Second

	// Store the original URL in Redis with expiration
	err := database.RDB.Set(ctx, shortURL, url.OriginalUrl, expiration).Err()
	if err != nil {
		log.Println(err)
		helpers.SendError(c, http.StatusInternalServerError, "Failed to store URL")
		return
	}

	url.ShortUrl = shortURL
	url.ClickDetails = []models.Click{}
	url.CreatedAt = time.Now()
	url.UpdatedAt = time.Now() 

	insertResult, err := urlCollection.InsertOne(context.Background(), url)
	if err != nil {
		log.Println("Database error:", err)
		helpers.SendError(c, http.StatusInternalServerError, "Failed to create url")
		return
	}

	url.ID = insertResult.InsertedID.(primitive.ObjectID)
	url.ShortUrl = fmt.Sprintf("%s/%s", configs.Env.APP_URL, shortURL)

	helpers.SendJSON(c, http.StatusCreated, gin.H{
		"data":    url,
		"message": "Url created successfully",
	})
}

func RedirectURL(c *gin.Context) {
	shortURL := c.Param("shortURL")

	// Fetch the original URL from Redis
	originalURL, err := database.RDB.Get(ctx, shortURL).Result()
	if err == redis.Nil {
		helpers.SendError(c, http.StatusNotFound, "URL not found or expired")
		return
	} else if err != nil {
		helpers.SendError(c, http.StatusInternalServerError, "Failed to retrieve URL")
		return
	}

	// Update click count and append user details
	url, err := urlCollection.UpdateOne(
		context.TODO(),
		bson.M{"shortUrl": shortURL},
		bson.M{
			"$inc": bson.M{"clickCount": 1},
			"$push": bson.M{"clickDetails": bson.M{
				"ipAddress": helpers.GetClientIP(c),
				"timestamp": time.Now(),
			}},
		},
	)

	if err != nil {
		log.Println("Database error:", err)
	}

	if url.MatchedCount == 0 {
		log.Println("Url not found")
	}

	// Redirect to the original URL
	c.Redirect(http.StatusFound, originalURL)
}


func GetUrl(c *gin.Context) {

	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		helpers.SendError(c, http.StatusBadRequest, "Invalid url ID")
		return
	}

	userId, ok := c.MustGet("userId").(primitive.ObjectID)
	if !ok {
		helpers.SendError(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var url models.Url
	err = urlCollection.FindOne(context.Background(), bson.M{"_id": id, "userId": userId}).Decode(&url)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			helpers.SendError(c, http.StatusNotFound, "Url not found")
		} else {
			helpers.SendError(c, http.StatusInternalServerError, "Failed to retrieve url")
		}
		return
	}

	helpers.SendJSON(c, http.StatusOK, gin.H{
		"data": url,
	})
}

func UpdateUrl(c *gin.Context) {

	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		helpers.SendError(c, http.StatusBadRequest, "Invalid url ID")
		return
	}

	userId, ok := c.MustGet("userId").(primitive.ObjectID)
	if !ok {
		helpers.SendError(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var url models.Url
	if err := c.ShouldBindJSON(&url); err != nil {
		helpers.SendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	var existingUrl models.Url
	err = urlCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&existingUrl)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			helpers.SendError(c, http.StatusNotFound, "Url not found")
		} else {
			helpers.SendError(c, http.StatusInternalServerError, "Failed to retrieve url")
		}
		return
	}

	if existingUrl.UserId != userId {
		helpers.SendError(c, http.StatusForbidden, "Unauthorized to update this url")
		return
	}

	expiration := time.Duration(url.Expiration) * time.Second
	err = database.RDB.Set(ctx, existingUrl.ShortUrl, url.OriginalUrl, expiration).Err()
	if err != nil {
		log.Println(err)
		helpers.SendError(c, http.StatusInternalServerError, "Failed to update URL")
		return
	}

	update := bson.M{
		"$set": bson.M{
			"originalUrl": url.OriginalUrl,
			"expiration":  url.Expiration,
			"updatedAt": time.Now(),
		},
	}

	result, err := urlCollection.UpdateOne(context.Background(), bson.M{"_id": id, "userId": userId}, update)
	if err != nil {
		log.Println("Database error:", err)
		helpers.SendError(c, http.StatusInternalServerError, "Failed to update url")
		return
	}

	if result.MatchedCount == 0 {
		helpers.SendError(c, http.StatusNotFound, "Url not found or unauthorized")
		return
	}

	helpers.SendJSON(c, http.StatusOK, gin.H{
		"message": "Url updated successfully",
	})
}

func GetAllUrl(c *gin.Context) {
	userId, ok := c.MustGet("userId").(primitive.ObjectID)
	if !ok {
		helpers.SendError(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	filter := bson.M{"userId": userId}

	var urls []models.Url
	params, err := helpers.PaginateCollection(c, urlCollection, filter, &urls)
	if err != nil {
		helpers.SendError(c, http.StatusInternalServerError, "Failed to retrieve urls")
		return
	}

	helpers.SendPaginatedResponse(c, urls, params)
}

func DeleteUrl(c *gin.Context) {

	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		helpers.SendError(c, http.StatusBadRequest, "Invalid url ID")
		return
	}

	userId, ok := c.MustGet("userId").(primitive.ObjectID)
	if !ok {
		helpers.SendError(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var existingUrl models.Url
	err = urlCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&existingUrl)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			helpers.SendError(c, http.StatusNotFound, "Url not found")
		} else {
			helpers.SendError(c, http.StatusInternalServerError, "Failed to retrieve url")
		}
		return
	}

	if existingUrl.UserId != userId {
		helpers.SendError(c, http.StatusForbidden, "Unauthorized to delete this url")
		return
	}

	filter := bson.M{"_id": id, "userId": userId}
	result, err := urlCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		helpers.SendError(c, http.StatusInternalServerError, "Failed to delete url")
		return
	}

	if result.DeletedCount == 0 {
		helpers.SendError(c, http.StatusNotFound, "Url not found or unauthorized")
		return
	}

	err = database.RDB.Del(ctx, existingUrl.ShortUrl).Err()
	if err != nil {
		helpers.SendError(c, http.StatusInternalServerError, "Failed to delete URL")
		return
	}

	helpers.SendJSON(c, http.StatusOK, gin.H{
		"message": "Url deleted successfully",
	})
}
