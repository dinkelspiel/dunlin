package dao

import (
	"database/sql"
	"time"

	"github.com/dinkelspiel/cdn/db"
	"github.com/dinkelspiel/cdn/models"
)

func scanCachedImageRow(rows *sql.Rows, db *db.DB) (*models.CachedImage, error) {
	var cachedImage models.CachedImage
	var createdAt string
	var updatedAt sql.NullString

	if err := rows.Scan(&cachedImage.Id, &cachedImage.Width, &cachedImage.Height, &cachedImage.CacheFile, &cachedImage.Directory, &cachedImage.File, &cachedImage.RotationDegrees, &cachedImage.SizeBytes, &cachedImage.TeamProjectId, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	createdAtTime, err := time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		return nil, err
	}
	cachedImage.CreatedAt = &createdAtTime

	if updatedAt.Valid {
		updatedAtTime, err := time.Parse("2006-01-02 15:04:05", updatedAt.String)
		if err != nil {
			return nil, err
		}
		cachedImage.UpdatedAt = &updatedAtTime
	} else {
		cachedImage.UpdatedAt = nil
	}

	cachedImage.TeamProject, err = GetTeamProjectById(db, cachedImage.TeamProjectId)
	if err != nil {
		return nil, err
	}

	return &cachedImage, nil
}

func GetCachedImageByOriginalAndWidthAndHeightAndRotation(db *db.DB, directory string, file string, width int, height int, rotationDegrees int) (*models.CachedImage, error) {
	rows, err := db.MariaDB.Query("SELECT id, width, height, cache_file, directory, file, rotation_degrees, size_bytes, team_project_id, created_at, updated_at FROM cached_images WHERE directory = ? AND file = ? AND width = ? AND height = ? AND rotation_degrees = ? ORDER BY created_at DESC LIMIT 1", directory, file, width, height, rotationDegrees)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanCachedImageRow(rows, db)
	}
	return nil, nil
}

func CreateCachedImage(db *db.DB, cachedImage models.CachedImage) (*models.CachedImage, error) {
	insertCachedImage := "INSERT INTO cached_images(width, height, cache_file, directory, file, rotation_degrees, size_bytes, team_project_id) VALUES(?, ?, ?, ?, ?, ?, ?, ?)"

	res, err := db.MariaDB.Exec(insertCachedImage, cachedImage.Width, cachedImage.Height, cachedImage.CacheFile, cachedImage.Directory, cachedImage.File, cachedImage.RotationDegrees, cachedImage.SizeBytes, cachedImage.TeamProject.Id)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()

	result := cachedImage
	result.Id = &id
	return &result, nil
}
