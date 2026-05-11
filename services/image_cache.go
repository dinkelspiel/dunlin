package services

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/dinkelspiel/cdn/dao"
	"github.com/dinkelspiel/cdn/db"
	"github.com/dinkelspiel/cdn/models"
	"github.com/google/uuid"
)

func getCacheDirectory() string {
	config, _ := LoadConfig()
	cacheDirectory := filepath.Join(config.StorageUrl, "/.cache")
	return cacheDirectory
}

func CacheImage(teamProject models.TeamProject, db *db.DB, image io.Reader, directory string, file string, width int, height int, rotationDegrees int) error {
	cacheDirectory := getCacheDirectory()
	cacheFile := uuid.NewString() + filepath.Ext(file)
	cacheFilePath := filepath.Join(cacheDirectory, cacheFile)

	if err := os.MkdirAll(cacheDirectory, os.ModePerm); err != nil {
		return err
	}

	outFile, err := os.Create(cacheFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	size, err := io.Copy(outFile, image)
	if err != nil {
		return err
	}

	cachedImage := models.CachedImage{
		Width:           width,
		Height:          height,
		CacheFile:       cacheFile,
		Directory:       directory,
		File:            file,
		RotationDegrees: rotationDegrees,
		SizeBytes:       size,
		TeamProjectId:   *teamProject.Id,
		TeamProject:     &teamProject,
	}

	_, err = dao.CreateCachedImage(db, cachedImage)
	if err != nil {
		return err
	}
	return nil
}

func RetrieveCachedImagePath(db *db.DB, directory string, file string, width int, height int, rotationDegrees int) (*string, error) {
	cachedImage, err := dao.GetCachedImageByOriginalAndWidthAndHeightAndRotation(db, directory, file, width, height, rotationDegrees)
	if err != nil {
		return nil, err
	}
	if cachedImage == nil {
		return nil, errors.New("no cache exists")
	}

	cacheDirectory := getCacheDirectory()
	cacheFilePath := filepath.Join(cacheDirectory, cachedImage.CacheFile)
	return &cacheFilePath, nil
}
