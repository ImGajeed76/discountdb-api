package models

import "time"

type Vote struct {
	ID        int64
	Timestamp time.Time
}

type VoteBody struct {
	ID  int64  `json:"id"`
	Dir string `json:"dir"`
}
