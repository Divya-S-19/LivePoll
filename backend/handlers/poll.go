package handlers

import (
	"context"
	"fmt"
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

	if len(req.Options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least 2 options are required",
		})
		return
	}

	for i := range req.Options {
		req.Options[i] = strings.TrimSpace(req.Options[i])

		if req.Options[i] == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Options cannot be empty",
			})
			return
		}
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

	// Check whether the submitted option actually belongs to this poll.
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

	// Update the live vote count in Redis.
	redisKey := "poll:" + id + ":option:" + req.Option

	count, err := h.RedisClient.Incr(ctx, redisKey).Result()

if err != nil {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "Could not update vote count",
	})
	return
}

// Publish the vote update through Redis Pub/Sub
channel := "poll:" + id + ":updates"

message := `{"option":"` + req.Option + `","count":` + fmt.Sprint(count) + `}`

err = h.RedisClient.Publish(ctx, channel, message).Err()

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