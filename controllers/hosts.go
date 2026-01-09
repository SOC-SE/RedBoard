package controllers

import (
	"net/http"

	"github.com/brian-l-johnson/Redteam-Dashboard-go/v2/models"
	"github.com/gin-gonic/gin"
)

type HostController struct{}

// GetHostsByTeam godoc
// @Summary Get hosts by team
// @Description Get all hosts for a specific team
// @Tags hosts
// @Accept json
// @Produce json
// @Param tid path string true "Team ID"
// @Success 200 {array} models.Host
// @Router /hosts/by-team/{tid} [get]
func (h HostController) GetHostsByTeam(c *gin.Context) {
	db := models.GetDB()
	var hosts []models.Host

	// Use eager loading to avoid N+1 queries
	results := db.Preload("Ports").
		Where("team_id = ?", c.Param("tid")).
		Order("ip ASC").
		Find(&hosts)

	if results.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": results.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, hosts)
}

// GetAllHostsByTeam godoc
// @Summary Get all hosts grouped by team
// @Description Get all teams with their hosts and ports
// @Tags hosts
// @Accept json
// @Produce json
// @Success 200 {array} models.Team
// @Router /hosts/by-team/ [get]
func (h HostController) GetAllHostsByTeam(c *gin.Context) {
	db := models.GetDB()
	var teams []models.Team

	// FIX: Use eager loading instead of N+1 queries
	// This single query replaces the loop that was doing N+1 queries
	results := db.
		Preload("Hosts").
		Preload("Hosts.Ports").
		Order("name ASC").
		Find(&teams)

	if results.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": results.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, teams)
}

// GetDashboardData godoc
// @Summary Get dashboard data
// @Description Get optimized data for the main dashboard
// @Tags hosts
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /dashboard/data [get]
func (h HostController) GetDashboardData(c *gin.Context) {
	db := models.GetDB()

	var teams []models.Team
	db.Preload("Hosts").Preload("Hosts.Ports").Order("name ASC").Find(&teams)

	// Calculate statistics
	totalHosts := 0
	totalPorts := 0
	dangerousPorts := 0

	for _, team := range teams {
		totalHosts += len(team.Hosts)
		for _, host := range team.Hosts {
			totalPorts += len(host.Ports)
			for _, port := range host.Ports {
				if port.IsDangerous() {
					dangerousPorts++
				}
			}
		}
	}

	// Get recent jobs
	var recentJobs []models.Job
	db.Order("created_at DESC").Limit(10).Find(&recentJobs)

	c.IndentedJSON(http.StatusOK, gin.H{
		"teams":           teams,
		"total_teams":     len(teams),
		"total_hosts":     totalHosts,
		"total_ports":     totalPorts,
		"dangerous_ports": dangerousPorts,
		"recent_jobs":     recentJobs,
	})
}
