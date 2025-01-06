package coupons

import (
	"context"
	"discountdb-api/internal/models"
	"discountdb-api/internal/repositories"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"strconv"
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
// @Produce json
// @Param dir path string true "Vote direction (up or down)"
// @Param id path string true "Coupon ID"
// @Success 200 {object} models.Success
// @Failure 400 {object} models.ErrorResponse
// @Router /coupons/vote/:dir/:id [post]
func PostVote(c *fiber.Ctx, rdb *redis.Client) error {
	// Get vote direction
	dir := c.Params("dir")
	if dir != "up" && dir != "down" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid vote direction",
		})
	}

	vote := models.VoteBody{
		Dir: dir,
	}

	if id, err := strconv.Atoi(c.Params("id")); err != nil || id < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.ErrorResponse{
				Message: "Invalid coupon ID",
			},
		)
	} else {
		vote.ID = int64(id)
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

	return c.JSON(models.Success{
		Message: "Vote successfully added to queue",
	})
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
