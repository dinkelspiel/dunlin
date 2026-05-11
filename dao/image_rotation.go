package dao

import (
	"database/sql"
	"time"

	"github.com/dinkelspiel/cdn/db"
	"github.com/dinkelspiel/cdn/models"
)

func scanImageRotation(scanner interface {
	Scan(dest ...any) error
}) (*models.ImageRotation, error) {
	var imageRotation models.ImageRotation
	var createdAt string
	var updatedAt sql.NullString

	if err := scanner.Scan(&imageRotation.Id, &imageRotation.TeamProjectId, &imageRotation.FilePath, &imageRotation.RotationDegrees, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	createdAtTime, err := time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		return nil, err
	}
	imageRotation.CreatedAt = &createdAtTime

	if updatedAt.Valid {
		updatedAtTime, err := time.Parse("2006-01-02 15:04:05", updatedAt.String)
		if err != nil {
			return nil, err
		}
		imageRotation.UpdatedAt = &updatedAtTime
	}

	return &imageRotation, nil
}

func GetImageRotationByProjectAndPath(db *db.DB, teamProjectId int64, filePath string) (*models.ImageRotation, error) {
	row := db.MariaDB.QueryRow("SELECT id, team_project_id, file_path, rotation_degrees, created_at, updated_at FROM image_rotations WHERE team_project_id = ? AND file_path = ?", teamProjectId, filePath)

	imageRotation, err := scanImageRotation(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return imageRotation, nil
}

func GetImageRotationsByProject(db *db.DB, teamProjectId int64) (map[string]int, error) {
	rows, err := db.MariaDB.Query("SELECT file_path, rotation_degrees FROM image_rotations WHERE team_project_id = ?", teamProjectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rotations := map[string]int{}
	for rows.Next() {
		var filePath string
		var rotationDegrees int
		if err := rows.Scan(&filePath, &rotationDegrees); err != nil {
			return nil, err
		}
		rotations[filePath] = rotationDegrees
	}

	return rotations, rows.Err()
}

func UpsertImageRotation(db *db.DB, imageRotation models.ImageRotation) (*models.ImageRotation, error) {
	res, err := db.MariaDB.Exec(
		"INSERT INTO image_rotations(team_project_id, file_path, rotation_degrees) VALUES(?, ?, ?) ON DUPLICATE KEY UPDATE rotation_degrees = VALUES(rotation_degrees), updated_at = CURRENT_TIMESTAMP",
		imageRotation.TeamProjectId,
		imageRotation.FilePath,
		imageRotation.RotationDegrees,
	)
	if err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()
	result := imageRotation
	if id != 0 {
		result.Id = &id
	}

	return &result, nil
}

func IncrementImageRotationClockwise(db *db.DB, teamProjectId int64, filePath string) (*models.ImageRotation, error) {
	_, err := db.MariaDB.Exec(
		"INSERT INTO image_rotations(team_project_id, file_path, rotation_degrees) VALUES(?, ?, 90) ON DUPLICATE KEY UPDATE rotation_degrees = MOD(rotation_degrees + 90, 360), updated_at = CURRENT_TIMESTAMP",
		teamProjectId,
		filePath,
	)
	if err != nil {
		return nil, err
	}

	return GetImageRotationByProjectAndPath(db, teamProjectId, filePath)
}
