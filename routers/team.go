package routers

import (
	"net/http"

	"github.com/dinkelspiel/cdn/dao"
	"github.com/dinkelspiel/cdn/db"
	"github.com/dinkelspiel/cdn/middleware"
	"github.com/dinkelspiel/cdn/models"
	"github.com/dinkelspiel/cdn/services"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

type CreateProjectBody struct {
	ProjectName string `json:"projectName" binding:"required"`
	TeamSlug    string `json:"teamSlug" binding:"required"`
}

func TeamRouter(v1 *gin.RouterGroup, db *db.DB) {
	team := v1.Group("/teams/:teamSlug")

	team.GET("", func(c *gin.Context) {
		teamSlug := c.Param("teamSlug")

		team, err := services.GetTeamBySlug(db, teamSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Found team",
			"team":    models.SerializeTeam(*team),
		})
	})

	team.GET("/projects", middleware.AuthMiddleware(db), func(c *gin.Context) {
		teamSlug := c.Param("teamSlug")

		team, err := services.GetTeamBySlug(db, teamSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		teamProjects, err := dao.GetTeamProjectsByTeam(db, *team)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		teamProjectsList := []gin.H{}

		for _, teamProject := range *teamProjects {
			teamProjectsList = append(teamProjectsList, models.SerializeTeamProject(teamProject))
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Found team projects",
			"teamProjects": teamProjectsList,
		})
	})

	team.POST("/projects", middleware.AuthMiddleware(db), func(c *gin.Context) {
		// authUser, _ := c.MustGet("authUser").(models.User)
		var body CreateProjectBody
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		teamSlug := body.TeamSlug

		team, err := dao.GetTeamBySlug(db, teamSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if team == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Team with slug doesn't exist"})
			return
		}

		// TODO: User authorization

		teamProject := models.TeamProject{
			Name:   body.ProjectName,
			Slug:   slug.Make(body.ProjectName),
			TeamId: *team.Id,
			Team:   team,
		}

		_, err = services.CreateTeamProject(db, teamProject)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Created project",
		})
	})
}
