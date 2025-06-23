package models

import "time"

type Link struct {
	ID        uint      `gorm:"primaryKey"`
	ShortCode string    `gorm:"uniqueIndex;size:10;not null"`
	LongURL   string    `gorm:"not null"`
	CreatedAt time.Time
}
