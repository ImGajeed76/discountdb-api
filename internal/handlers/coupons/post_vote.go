package coupons

import (
	"context"
	"discountdb-api/internal/models"
	"discountdb-api/internal/repositories"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"time"
)

type VoteQueue struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	VoteType  string    `json:"vote_type"`
}

// PostVote godoc
// @Summary Vote on a coupon
// @Description Vote on a coupon by ID
// @Tags votes
// @Accept json
// @Produce json
// @Param vote body models.VoteBody true "Vote body"
// @Success 200
// @Failure 400 {object} models.ErrorResponse
// @Router /coupons/vote [post]
func PostVote(c *fiber.Ctx, rdb *redis.Client) error {

	var vote models.VoteBody
	if err := c.BodyParser(&vote); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Message: "Invalid request body"})
	}

	if vote.Dir != "up" && vote.Dir != "down" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Message: "Invalid vote direction"})
	}

	if vote.ID < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Message: "Invalid coupon ID"})
	}

	voteQueue := VoteQueue{
		ID:        vote.ID,
		Timestamp: time.Now(),
		VoteType:  vote.Dir,
	}

	queueJSON, err := json.Marshal(voteQueue)
	if err != nil {
		return err
	}

	err = rdb.RPush(c.Context(), "vote_queue", queueJSON).Err()
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func ProcessVoteQueue(ctx context.Context, couponRepo *repositories.CouponRepository, rdb *redis.Client, batchSize int) error {
	for {
		// Get votes batch
		results, err := rdb.LRange(ctx, "vote_queue", 0, int64(batchSize-1)).Result()
		if err != nil {
			return err
		}
		if len(results) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		upVotes := []models.Vote{}
		downVotes := []models.Vote{}

		// Parse votes
		for _, result := range results {
			var voteQueue VoteQueue
			if err := json.Unmarshal([]byte(result), &voteQueue); err != nil {
				continue
			}

			vote := models.Vote{ID: voteQueue.ID, Timestamp: voteQueue.Timestamp}
			if voteQueue.VoteType == "up" {
				upVotes = append(upVotes, vote)
			} else {
				downVotes = append(downVotes, vote)
			}
		}

		// Process votes
		if len(upVotes) > 0 {
			if err := couponRepo.BatchAddVotes(ctx, upVotes, "up"); err != nil {
				return err
			}
		}
		if len(downVotes) > 0 {
			if err := couponRepo.BatchAddVotes(ctx, downVotes, "down"); err != nil {
				return err
			}
		}

		// Remove processed votes
		rdb.LTrim(ctx, "vote_queue", int64(len(results)), -1)
	}
}
