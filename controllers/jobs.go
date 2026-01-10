package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/brian-l-johnson/Redteam-Dashboard-go/v2/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobController struct{}

// GetJobManagerState godoc
// @Summary Get job manager state
// @Description Get the current state of job schedulers
// @Tags jobs
// @Accept json
// @Produce json
// @Success 200 {array} models.JobStatus
// @Router /jobs/manager [get]
func (j JobController) GetJobManagerState(c *gin.Context) {
	db := models.GetDB()
	var js []models.JobStatus
	results := db.Find(&js)
	if results.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": results.Error.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, js)
}

// NewJob godoc
// @Summary Get next job
// @Description Get the next job to process for a given job type
// @Tags jobs
// @Accept json
// @Produce json
// @Param jobtype path string true "Job Type (e.g., nmap)"
// @Success 200 {object} models.Job
// @Router /jobs/{jobtype}/next [get]
func (j JobController) NewJob(c *gin.Context) {
	db := models.GetDB()

	var js models.JobStatus
	result := db.First(&js, "name = ?", c.Param("jobtype"))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"status": "error", "message": "unknown job type"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	job, err := js.GetNextJob()
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}

	job.StartedAt = time.Now()
	job.Status = "running"
	db.Create(&job)

	c.IndentedJSON(http.StatusOK, job)
}

// GetJobs godoc
// @Summary Get all jobs
// @Description Get list of all jobs with optional filtering
// @Tags jobs
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {array} models.Job
// @Router /jobs [get]
func (j JobController) GetJobs(c *gin.Context) {
	db := models.GetDB()
	var jobs []models.Job

	query := db.Order("created_at DESC")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		// Parse limit from query
	}
	query = query.Limit(limit)

	result := query.Find(&jobs)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, jobs)
}

// UploadScan godoc
// @Summary Upload scan results
// @Description Upload nmap scan results for a job
// @Tags jobs
// @Accept json
// @Produce json
// @Param jid path string true "Job ID"
// @Param scan body models.Scan true "Scan data"
// @Success 200 {object} map[string]interface{}
// @Router /jobs/nmap/{jid} [post]
func (j JobController) UploadScan(c *gin.Context) {
	var scan models.Scan
	if err := c.ShouldBindJSON(&scan); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid scan data: " + err.Error()})
		return
	}

	db := models.GetDB()
	jid := c.Param("jid")

	var job models.Job
	if err := db.First(&job, "j_id = ?", jid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"status": "error", "message": "job not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Start transaction for atomic update
	tx := db.Begin()

	// FIX: Instead of deleting all hosts, update existing ones and add new ones
	// This preserves historical data and prevents data loss

	// Get existing hosts for this team
	var existingHosts []models.Host
	tx.Where("team_id = ?", job.TID).Find(&existingHosts)

	// Create a map of existing hosts by IP for quick lookup
	existingHostMap := make(map[string]*models.Host)
	for i := range existingHosts {
		existingHostMap[existingHosts[i].IP] = &existingHosts[i]
	}

	hostsProcessed := 0
	portsProcessed := 0

	for _, scanHost := range scan.Hosts {
		var host *models.Host

		if existing, found := existingHostMap[scanHost.IP]; found {
			// Update existing host
			host = existing
			host.Hostname = scanHost.Hostname
			host.OS = scanHost.OS
			host.LastSeen = time.Now()
			host.Status = "online"

			// Delete old ports and add new ones
			tx.Where("host_id = ?", host.ID).Delete(&models.Port{})
		} else {
			// Create new host
			newHost := models.Host{
				IP:       scanHost.IP,
				Hostname: scanHost.Hostname,
				OS:       scanHost.OS,
				TeamID:   job.TID,
				LastSeen: time.Now(),
				Status:   "online",
			}
			if err := tx.Create(&newHost).Error; err != nil {
				tx.Rollback()
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
				return
			}
			host = &newHost
		}

		// Add ports
		for _, port := range scanHost.Ports {
			port.HostID = host.ID
			if err := tx.Create(&port).Error; err != nil {
				tx.Rollback()
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
				return
			}
			
			// Save script results if present
			for i := range port.Scripts {
				port.Scripts[i].PortID = port.ID
				if err := tx.Create(&port.Scripts[i]).Error; err != nil {
					// Log but don't fail on script save errors
					fmt.Printf("Warning: failed to save script result: %v\n", err)
				}
			}
			
			portsProcessed++
		}

		if err := tx.Save(host).Error; err != nil {
			tx.Rollback()
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}

		hostsProcessed++
		delete(existingHostMap, scanHost.IP)
	}

	// Mark hosts not seen in this scan as potentially offline
	for _, host := range existingHostMap {
		host.Status = "offline"
		tx.Save(host)
	}

	// Update job status
	job.Status = "complete"
	job.CompletedAt = time.Now()
	job.HostsFound = hostsProcessed
	job.PortsFound = portsProcessed
	tx.Save(&job)

	// Record scan history
	history := models.ScanHistory{
		TeamID:    job.TID,
		ScanTime:  time.Now(),
		HostCount: hostsProcessed,
		PortCount: portsProcessed,
	}
	tx.Create(&history)

	if err := tx.Commit().Error; err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"status":          "success",
		"hosts_processed": hostsProcessed,
		"ports_processed": portsProcessed,
	})
}

// CancelJob godoc
// @Summary Cancel a job
// @Description Cancel a running or queued job
// @Tags jobs
// @Accept json
// @Produce json
// @Param jid path string true "Job ID"
// @Success 200 {object} map[string]string
// @Router /jobs/{jid}/cancel [post]
func (j JobController) CancelJob(c *gin.Context) {
	db := models.GetDB()
	var job models.Job

	if err := db.First(&job, "j_id = ?", c.Param("jid")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"status": "error", "message": "job not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	if job.Status == "complete" || job.Status == "failed" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "cannot cancel completed job"})
		return
	}

	job.Status = "cancelled"
	job.CompletedAt = time.Now()
	db.Save(&job)

	c.IndentedJSON(http.StatusOK, gin.H{"status": "success", "message": "job cancelled"})
}
