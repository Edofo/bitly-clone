package models

import "time"

type Click struct {
	ID        uint      `gorm:"primaryKey"`
	LinkID    uint      `gorm:"index"`
	Link      Link      `gorm:"foreignKey:LinkID"`
	Timestamp time.Time
	UserAgent string    `gorm:"size:255"`
	IPAddress string    `gorm:"size:50"`
}


type ClickEvent struct {
	LinkID    uint
	Timestamp time.Time
	UserAgent string
	IPAddress string
}
