package routers

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dinkelspiel/cdn/db"
	"github.com/dinkelspiel/cdn/services"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

func rotateImage(srcImage image.Image, rotationDegrees int) image.Image {
	switch rotationDegrees {
	case 90:
		return imaging.Rotate270(srcImage)
	case 180:
		return imaging.Rotate180(srcImage)
	case 270:
		return imaging.Rotate90(srcImage)
	default:
		return srcImage
	}
}

func FileRouter(r *gin.RouterGroup, db *db.DB) {
	r.GET("/files/:teamSlug/:projectSlug/*filepath", func(c *gin.Context) {
		teamSlug := c.Param("teamSlug")
		teamProjectSlug := c.Param("projectSlug")

		_, teamProject, err := services.GetTeamAndProjectBySlug(db, teamSlug, teamProjectSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filePath := c.Param("filepath")

		path, err := services.GetFilePathToFileInTeamProject(*teamProject, filePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		mimeType := mime.TypeByExtension(filepath.Ext(path))
		if strings.HasPrefix(mimeType, "image/") {
			srcImage, err := imaging.Open(path)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
				return
			}
			rotationDegrees, err := services.GetTeamProjectImageRotation(db, *teamProject, filePath)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			srcImage = rotateImage(srcImage, rotationDegrees)

			oldWidth := srcImage.Bounds().Dx()
			width := srcImage.Bounds().Dx()
			oldHeight := srcImage.Bounds().Dy()
			height := srcImage.Bounds().Dy()

			widthStr := c.Query("w")
			heightStr := c.Query("h")
			if w, err := strconv.Atoi(widthStr); err == nil {
				width = w
			}
			if h, err := strconv.Atoi(heightStr); err == nil {
				height = h
			}

			// Return early if no transformation has to be made
			if rotationDegrees == 0 && oldWidth == width && oldHeight == height {
				c.Header("Content-Description", "File Transfer")
				c.File(path)
				return
			}

			if rotationDegrees == 0 {
				cacheFilePath, _ := services.RetrieveCachedImagePath(db, filepath.Dir(path), filePath, width, height)
				if cacheFilePath != nil {
					c.Header("Content-Description", "File Transfer")
					c.File(*cacheFilePath)
					fmt.Printf("Retrieved cache %s", *cacheFilePath)
					return
				}
			}

			dstImage := imaging.Resize(srcImage, width, height, imaging.Lanczos)

			var buf bytes.Buffer
			format := imaging.PNG
			switch strings.ToLower(filepath.Ext(path)) {
			case ".png":
				format = imaging.PNG
			case ".jpg", ".jpeg":
				format = imaging.JPEG
			case ".gif":
				format = imaging.GIF
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported image format"})
				return
			}

			if err := imaging.Encode(&buf, dstImage, format); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode image"})
				return
			}

			imageBytes := buf.Bytes()

			c.Header("Content-Type", mimeType)
			c.Header("Content-Disposition", "inline")
			_, err = io.Copy(c.Writer, bytes.NewReader(imageBytes))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream image"})
				return
			}

			if rotationDegrees == 0 {
				go func(data []byte) {
					err := services.CacheImage(*teamProject, db, bytes.NewReader(data), filepath.Dir(path), filePath, width, height)
					if err != nil {
						log.Printf("Failed to cache image: %v", err)
					}
				}(imageBytes)
			}

			return
		}

		c.Header("Content-Description", "File Transfer")
		c.File(path)
	})
}
