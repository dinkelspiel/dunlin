package models

import "time"

type CachedImage struct {
	Id              *int64
	Width           int
	Height          int
	CacheFile       string
	Directory       string
	File            string
	RotationDegrees int
	SizeBytes       int64
	TeamProjectId   int64
	TeamProject     *TeamProject
	UpdatedAt       *time.Time
	CreatedAt       *time.Time
}
