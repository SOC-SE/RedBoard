package controllers

import (
	"net/http"
	"strings"

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

// GetVulnerabilities godoc
// @Summary Get vulnerability findings
// @Description Get all NSE script findings that indicate vulnerabilities
// @Tags hosts
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /vulnerabilities [get]
func (h HostController) GetVulnerabilities(c *gin.Context) {
	db := models.GetDB()

	var teams []models.Team
	db.Preload("Hosts").Preload("Hosts.Ports").Preload("Hosts.Ports.Scripts").Order("name ASC").Find(&teams)

	type VulnFinding struct {
		TeamName   string `json:"team_name"`
		TeamID     string `json:"team_id"`
		HostIP     string `json:"host_ip"`
		Hostname   string `json:"hostname"`
		Port       uint16 `json:"port"`
		Protocol   string `json:"protocol"`
		Service    string `json:"service"`
		ScriptName string `json:"script_name"`
		Output     string `json:"output"`
		Severity   string `json:"severity"`
	}

	var findings []VulnFinding

	// Keywords that indicate vulnerabilities
	vulnKeywords := []string{"vulnerable", "anonymous", "empty password", "authentication disabled", "no authentication", "unrestricted"}
	
	for _, team := range teams {
		for _, host := range team.Hosts {
			for _, port := range host.Ports {
				// Check script results for vulnerability indicators
				for _, script := range port.Scripts {
					outputLower := strings.ToLower(script.Output)
					for _, keyword := range vulnKeywords {
						if strings.Contains(outputLower, keyword) {
							severity := "medium"
							if strings.Contains(outputLower, "vulnerable") {
								severity = "critical"
							} else if strings.Contains(outputLower, "anonymous") || strings.Contains(outputLower, "empty password") {
								severity = "high"
							}
							
							findings = append(findings, VulnFinding{
								TeamName:   team.Name,
								TeamID:     team.TID,
								HostIP:     host.IP,
								Hostname:   host.Hostname,
								Port:       port.Number,
								Protocol:   port.Protocol,
								Service:    port.Service,
								ScriptName: script.Name,
								Output:     script.Output,
								Severity:   severity,
							})
							break // Only add once per script
						}
					}
				}
			}
		}
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"findings":    findings,
		"total_count": len(findings),
	})
}
