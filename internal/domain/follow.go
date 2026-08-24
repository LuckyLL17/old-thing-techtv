package domain

import "time"

type Follow struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	FollowerID  uint64    `gorm:"index;not null;index:follow_pair" json:"follower_id"`
	FollowingID uint64    `gorm:"index;not null" json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func (f *Follow) TableName() string {
	return "follows"
}

func v6Task008Boundary1(value uint64) bool {
	return value > 0
}
