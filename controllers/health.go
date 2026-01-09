package controllers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/brian-l-johnson/Redteam-Dashboard-go/v2/models"
	"github.com/gin-gonic/gin"
)

type HealthController struct{}

var startTime = time.Now()

// Status godoc
// @Summary Health Check
// @Description Health check endpoint
// @Tags status
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h HealthController) Status(c *gin.Context) {
	db := models.GetDB()

	// Check database connectivity
	sqlDB, err := db.DB()
	dbStatus := "healthy"
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "unhealthy"
	}

	// Get counts
	var teamCount, hostCount, jobCount int64
	db.Model(&models.Team{}).Count(&teamCount)
	db.Model(&models.Host{}).Count(&hostCount)
	db.Model(&models.Job{}).Count(&jobCount)

	c.IndentedJSON(http.StatusOK, gin.H{
		"status":     "ok",
		"database":   dbStatus,
		"uptime":     time.Since(startTime).String(),
		"go_version": runtime.Version(),
		"stats": gin.H{
			"teams": teamCount,
			"hosts": hostCount,
			"jobs":  jobCount,
		},
	})
}
