package routers

import (
	"net/http"

	"github.com/dinkelspiel/cdn/db"
	"github.com/dinkelspiel/cdn/middleware"
	"github.com/dinkelspiel/cdn/models"
	"github.com/dinkelspiel/cdn/services"
	"github.com/dinkelspiel/cdn/storage"
	"github.com/gin-gonic/gin"
)

type CreateFolderBody struct {
	Path string `json:"path" binding:"required"`
}

type RotateAlbumImageBody struct {
	Path string `json:"path" binding:"required"`
}

func TeamProjectRouter(v1 *gin.RouterGroup, db *db.DB) {
	team := v1.Group("/teams/:teamSlug/projects/:projectSlug")
	team.GET("", func(c *gin.Context) {
		teamSlug := c.Param("teamSlug")
		teamProjectSlug := c.Param("projectSlug")

		_, teamProject, err := services.GetTeamAndProjectBySlug(db, teamSlug, teamProjectSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Found team project",
			"teamProject": models.SerializeTeamProject(*teamProject),
		})
	})

	team.GET("/files/*filepath", func(c *gin.Context) {
		teamSlug := c.Param("teamSlug")
		teamProjectSlug := c.Param("projectSlug")

		_, teamProject, err := services.GetTeamAndProjectBySlug(db, teamSlug, teamProjectSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filepath := c.Param("filepath")

		files, err := services.GetTeamProjectFiles(*teamProject, filepath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fileList := []gin.H{}

		for _, file := range files {
			fileList = append(fileList, storage.SerializeFSItem(file))
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Found files",
			"files":   fileList,
		})
	})

	team.GET("/album", func(c *gin.Context) {
		teamSlug := c.Param("teamSlug")
		teamProjectSlug := c.Param("projectSlug")

		_, teamProject, err := services.GetTeamAndProjectBySlug(db, teamSlug, teamProjectSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		images, err := services.GetTeamProjectImages(db, *teamProject)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		imageList := []gin.H{}
		for _, image := range images {
			imageList = append(imageList, services.SerializeTeamProjectImage(image))
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Found album images",
			"images":  imageList,
		})
	})

	team.POST("/album/rotate", middleware.AuthMiddleware(db), func(c *gin.Context) {
		var body RotateAlbumImageBody
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		teamSlug := c.Param("teamSlug")
		teamProjectSlug := c.Param("projectSlug")

		_, teamProject, err := services.GetTeamAndProjectBySlug(db, teamSlug, teamProjectSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		imageRotation, err := services.RotateTeamProjectImageClockwise(db, *teamProject, body.Path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         "Rotated album image",
			"path":            imageRotation.FilePath,
			"rotationDegrees": imageRotation.RotationDegrees,
		})
	})

	team.PUT("/files/*filepath", middleware.AuthMiddleware(db), func(c *gin.Context) {
		teamSlug := c.Param("teamSlug")
		teamProjectSlug := c.Param("projectSlug")

		_, teamProject, err := services.GetTeamAndProjectBySlug(db, teamSlug, teamProjectSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
			return
		}

		relativePath := c.Param("filepath")
		fullPath, err := services.GetFilePathToFileInTeamProject(*teamProject, relativePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := c.SaveUploadedFile(file, fullPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Uploaded file",
		})
	})

	team.POST("/folders", middleware.AuthMiddleware(db), func(c *gin.Context) {
		var body CreateFolderBody
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		teamSlug := c.Param("teamSlug")
		teamProjectSlug := c.Param("projectSlug")

		_, teamProject, err := services.GetTeamAndProjectBySlug(db, teamSlug, teamProjectSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		path, err := services.CreateTeamProjectFolder(*teamProject, body.Path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Created folder " + path,
		})
	})
}
