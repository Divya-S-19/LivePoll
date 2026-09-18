package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"livepoll-backend/handlers"
	"livepoll-backend/middleware"
)

func main() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get MongoDB connection string
	mongodbURI := os.Getenv("MONGODB_URI")

	if mongodbURI == "" {
		log.Fatal("MONGODB_URI is not set")
	}

	// Connect to MongoDB
	client, err := mongo.Connect(options.Client().ApplyURI(mongodbURI))
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	// Test MongoDB connection
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	fmt.Println("MongoDB connected successfully!")

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	// Test Redis connection
	redisCtx, redisCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer redisCancel()

	_, err = redisClient.Ping(redisCtx).Result()
	if err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	fmt.Println("Redis connected successfully!")

	// MongoDB users collection
	userCollection := client.Database("livepoll").Collection("users")

	// Create authentication handler
	authHandler := &handlers.AuthHandler{
		UserCollection: userCollection,
	}

	// MongoDB polls collection
	pollCollection := client.Database("livepoll").Collection("polls")

	// Create poll handler
	pollHandler := &handlers.PollHandler{
	PollCollection: pollCollection,
	RedisClient:    redisClient,
    }

	// Create Gin router
	router := gin.Default()
	router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:5173"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}))

	// Test route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "LivePoll backend is running",
		})
	})

	// Authentication routes
	router.POST("/signup", authHandler.Signup)
	router.POST("/login", authHandler.Login)

	// Protected poll creation route
	router.POST(
		"/polls",
		middleware.AuthMiddleware(),
		pollHandler.CreatePoll,
	)
	router.GET("/polls/:id", pollHandler.GetPoll)
	router.GET("/polls/:id/results", pollHandler.GetResults)
    router.POST("/polls/:id/vote", pollHandler.Vote)
    router.GET("/polls/:id/live", pollHandler.Realtime)

	// Protected test route
	router.GET(
		"/protected",
		middleware.AuthMiddleware(),
		func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "You are authorized!",
			})
		},
	)

	// Start server
	router.Run(":8081")
}