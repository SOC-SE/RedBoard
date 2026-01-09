package models

import (
	"errors"
	"net"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model  `json:"-"`
	Name        string `json:"name" gorm:"uniqueIndex"`
	IPRange     string `json:"iprange"`
	TID         string `json:"tid" gorm:"uniqueIndex;column:t_id"`
	Description string `json:"description"`
	Color       string `json:"color"` // Hex color for UI display
	Hosts       []Host `json:"hosts,omitempty" gorm:"foreignKey:TeamID;references:TID;constraint:OnDelete:CASCADE"`
}

// TeamRequest for creating/updating teams via API
type TeamRequest struct {
	Name        string `json:"name" binding:"required"`
	IPRange     string `json:"iprange" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func MakeTeam(name string, iprange string) Team {
	var team Team
	team.Name = name
	team.IPRange = iprange
	team.TID = uuid.New().String()
	team.Color = generateTeamColor(name)
	return team
}

// ValidateIPRange checks if the IP range is valid for nmap
func ValidateIPRange(iprange string) error {
	iprange = strings.TrimSpace(iprange)
	
	if iprange == "" {
		return errors.New("IP range cannot be empty")
	}

	// Check for CIDR notation (e.g., 192.168.1.0/24)
	if strings.Contains(iprange, "/") {
		_, _, err := net.ParseCIDR(iprange)
		if err == nil {
			return nil
		}
	}

	// Check for range notation (e.g., 192.168.1.1-254)
	rangePattern := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}-\d{1,3}$`)
	if rangePattern.MatchString(iprange) {
		return nil
	}

	// Check for single IP
	if net.ParseIP(iprange) != nil {
		return nil
	}

	// Check for comma-separated IPs
	ips := strings.Split(iprange, ",")
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if net.ParseIP(ip) == nil {
			// Check if it's a valid CIDR
			_, _, err := net.ParseCIDR(ip)
			if err != nil {
				return errors.New("invalid IP address or range: " + ip)
			}
		}
	}

	return nil
}

// generateTeamColor creates a consistent color based on team name
func generateTeamColor(name string) string {
	colors := []string{
		"#3B82F6", // Blue
		"#10B981", // Green
		"#F59E0B", // Amber
		"#EF4444", // Red
		"#8B5CF6", // Purple
		"#EC4899", // Pink
		"#06B6D4", // Cyan
		"#F97316", // Orange
	}
	
	// Simple hash based on name
	hash := 0
	for _, c := range name {
		hash = (hash*31 + int(c)) % len(colors)
	}
	return colors[hash]
}
