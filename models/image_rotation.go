package models

import "time"

type ImageRotation struct {
	Id              *int64
	TeamProjectId   int64
	FilePath        string
	RotationDegrees int
	UpdatedAt       *time.Time
	CreatedAt       *time.Time
}
