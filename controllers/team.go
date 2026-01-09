package controllers

import (
	"errors"
	"net/http"

	"github.com/brian-l-johnson/Redteam-Dashboard-go/v2/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TeamController struct{}

// GetTeams godoc
// @Summary Get Teams
// @Description Get all teams with optional host data
// @Tags teams
// @Accept json
// @Produce json
// @Param include_hosts query bool false "Include host data"
// @Success 200 {array} models.Team
// @Router /teams [get]
func (t TeamController) GetTeams(c *gin.Context) {
	db := models.GetDB()
	var teams []models.Team

	includeHosts := c.Query("include_hosts") == "true"

	query := db.Order("name ASC")
	if includeHosts {
		query = query.Preload("Hosts").Preload("Hosts.Ports")
	}

	result := query.Find(&teams)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, teams)
}

// GetTeam godoc
// @Summary Get Team by ID
// @Description Get a single team with hosts
// @Tags teams
// @Accept json
// @Produce json
// @Param tid path string true "Team ID"
// @Success 200 {object} models.Team
// @Router /teams/{tid} [get]
func (t TeamController) GetTeam(c *gin.Context) {
	db := models.GetDB()
	var team models.Team

	// FIX: Use TID field correctly
	result := db.Preload("Hosts").Preload("Hosts.Ports").First(&team, "t_id = ?", c.Param("tid"))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"status": "error", "message": "team not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, team)
}

// CreateTeam godoc
// @Summary Create team
// @Description Create a new team
// @Tags teams
// @Accept json
// @Produce json
// @Param team body models.TeamRequest true "Team data"
// @Success 201 {object} models.Team
// @Router /teams [post]
func (t TeamController) CreateTeam(c *gin.Context) {
	var req models.TeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Validate IP range
	if err := models.ValidateIPRange(req.IPRange); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	db := models.GetDB()

	// Check if team name already exists
	var existing models.Team
	result := db.First(&existing, "name = ?", req.Name)
	if result.Error == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"status": "error", "message": "team name already exists"})
		return
	}

	team := models.MakeTeam(req.Name, req.IPRange)
	team.Description = req.Description
	if req.Color != "" {
		team.Color = req.Color
	}

	result = db.Create(&team)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, team)
}

// UpdateTeam godoc
// @Summary Update team
// @Description Update an existing team
// @Tags teams
// @Accept json
// @Produce json
// @Param tid path string true "Team ID"
// @Param team body models.TeamRequest true "Team data"
// @Success 200 {object} models.Team
// @Router /teams/{tid} [put]
func (t TeamController) UpdateTeam(c *gin.Context) {
	db := models.GetDB()
	var team models.Team

	// FIX: Use TID field correctly
	result := db.First(&team, "t_id = ?", c.Param("tid"))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"status": "error", "message": "team not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	var req models.TeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Validate IP range
	if err := models.ValidateIPRange(req.IPRange); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Check if new name conflicts with another team
	if req.Name != team.Name {
		var existing models.Team
		result := db.First(&existing, "name = ? AND t_id != ?", req.Name, team.TID)
		if result.Error == nil {
			c.IndentedJSON(http.StatusConflict, gin.H{"status": "error", "message": "team name already exists"})
			return
		}
	}

	team.Name = req.Name
	team.IPRange = req.IPRange
	team.Description = req.Description
	if req.Color != "" {
		team.Color = req.Color
	}

	result = db.Save(&team)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, team)
}

// DeleteTeam godoc
// @Summary Delete team
// @Description Delete a team and its hosts
// @Tags teams
// @Accept json
// @Produce json
// @Param tid path string true "Team ID"
// @Success 200 {object} map[string]string
// @Router /teams/{tid} [delete]
func (t TeamController) DeleteTeam(c *gin.Context) {
	db := models.GetDB()
	var team models.Team

	// FIX: Use TID field correctly (was using ID which is wrong)
	result := db.First(&team, "t_id = ?", c.Param("tid"))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"status": "error", "message": "team not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	// Delete associated hosts and ports (cascade should handle this, but be explicit)
	var hosts []models.Host
	db.Find(&hosts, "team_id = ?", team.TID)
	for _, host := range hosts {
		db.Where("host_id = ?", host.ID).Delete(&models.Port{})
	}
	db.Where("team_id = ?", team.TID).Delete(&models.Host{})

	// Delete the team
	result = db.Delete(&team)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"status": "success", "message": "team deleted"})
}
