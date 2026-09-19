package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"livepoll-backend/models"
)

type PollHandler struct {
	PollCollection *mongo.Collection
	RedisClient    *redis.Client
}

type CreatePollRequest struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

func (h *PollHandler) CreatePoll(c *gin.Context) {
	var req CreatePollRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	req.Question = strings.TrimSpace(req.Question)

	if req.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Question is required",
		})
		return
	}

	// Server-side question length validation.
	if len(req.Question) > 300 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Question must be 300 characters or less",
		})
		return
	}

	if len(req.Options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least 2 options are required",
		})
		return
	}

	// Prevent excessively large polls.
	if len(req.Options) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "A maximum of 10 options is allowed",
		})
		return
	}

	// Track duplicate options.
	seenOptions := make(map[string]bool)

	for i := range req.Options {
		req.Options[i] = strings.TrimSpace(req.Options[i])

		if req.Options[i] == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Options cannot be empty",
			})
			return
		}

		if len(req.Options[i]) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Each option must be 100 characters or less",
			})
			return
		}

		// Prevent duplicate options.
		optionKey := strings.ToLower(req.Options[i])

		if seenOptions[optionKey] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Duplicate options are not allowed",
			})
			return
		}

		seenOptions[optionKey] = true
	}

	poll := models.Poll{
		ID:        bson.NewObjectID(),
		Question:  req.Question,
		Options:   req.Options,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.PollCollection.InsertOne(ctx, poll)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create poll",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Poll created successfully",
		"poll":    poll,
	})
}

func (h *PollHandler) GetPoll(c *gin.Context) {
	id := c.Param("id")

	pollID, err := bson.ObjectIDFromHex(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid poll ID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var poll models.Poll

	err = h.PollCollection.FindOne(
		ctx,
		bson.M{"_id": pollID},
	).Decode(&poll)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Poll not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"poll": poll,
	})
}

type VoteRequest struct {
	Option string `json:"option"`
}

func (h *PollHandler) Vote(c *gin.Context) {
	id := c.Param("id")

	pollID, err := bson.ObjectIDFromHex(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid poll ID",
		})
		return
	}

	var req VoteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	req.Option = strings.TrimSpace(req.Option)

	if req.Option == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Option is required",
		})
		return
	}

	if len(req.Option) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Option is too long",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var poll models.Poll

	err = h.PollCollection.FindOne(
		ctx,
		bson.M{"_id": pollID},
	).Decode(&poll)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Poll not found",
		})
		return
	}

	// Server-side validation:
	// make sure the submitted option belongs to this poll.
	validOption := false

	for _, option := range poll.Options {
		if option == req.Option {
			validOption = true
			break
		}
	}

	if !validOption {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid option",
		})
		return
	}

	// Save vote in MongoDB.
	vote := models.Vote{
		ID:        bson.NewObjectID(),
		PollID:    pollID,
		Option:    req.Option,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	voteCollection := h.PollCollection.Database().Collection("votes")

	_, err = voteCollection.InsertOne(ctx, vote)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not save vote",
		})
		return
	}

	// Update live count in Redis.
	redisKey := "poll:" + id + ":option:" + req.Option

	count, err := h.RedisClient.Incr(ctx, redisKey).Result()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update vote count",
		})
		return
	}

	// Publish a valid JSON message through Redis Pub/Sub.
	channel := "poll:" + id + ":updates"

	updateMessage := map[string]interface{}{
		"option": req.Option,
		"count":  count,
	}

	messageBytes, err := json.Marshal(updateMessage)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create vote update",
		})
		return
	}

	err = h.RedisClient.Publish(
		ctx,
		channel,
		string(messageBytes),
	).Err()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not publish vote update",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Vote submitted successfully",
		"count":   count,
	})
}

func (h *PollHandler) GetResults(c *gin.Context) {
	id := c.Param("id")

	pollID, err := bson.ObjectIDFromHex(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid poll ID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var poll models.Poll

	err = h.PollCollection.FindOne(
		ctx,
		bson.M{"_id": pollID},
	).Decode(&poll)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Poll not found",
		})
		return
	}

	counts := make(map[string]int64)
	var total int64

	for _, option := range poll.Options {
		key := "poll:" + id + ":option:" + option

		count, err := h.RedisClient.Get(ctx, key).Int64()

		if err != nil && err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not get results",
			})
			return
		}

		counts[option] = count
		total += count
	}

	c.JSON(http.StatusOK, gin.H{
		"counts": counts,
		"total":  total,
	})
}