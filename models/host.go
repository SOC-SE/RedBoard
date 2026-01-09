package models

import (
	"time"

	"gorm.io/gorm"
)

type Host struct {
	gorm.Model `json:"-"`
	IP         string `json:"ip" gorm:"index"`
	Hostname   string `json:"hostname"`
	OS         string `json:"os"`
	Ports      []Port `json:"ports" gorm:"foreignKey:HostID;constraint:OnDelete:CASCADE"`
	TeamID     string `json:"team_id" gorm:"index"`
	LastSeen   time.Time `json:"last_seen"`
	Status     string `json:"status"` // online, offline, unknown
}

// PortBaseline defines expected ports for monitoring
type PortBaseline struct {
	gorm.Model `json:"-"`
	TeamID     string `json:"team_id" gorm:"index"`
	HostIP     string `json:"host_ip"` // Can be "*" for all hosts in team
	Port       uint16 `json:"port"`
	Protocol   string `json:"protocol"`
	Service    string `json:"service"`
	Expected   bool   `json:"expected"` // true = expected, false = should alert if found
}

// ScanHistory tracks scan results over time
type ScanHistory struct {
	gorm.Model  `json:"-"`
	TeamID      string    `json:"team_id" gorm:"index"`
	ScanTime    time.Time `json:"scan_time"`
	HostCount   int       `json:"host_count"`
	PortCount   int       `json:"port_count"`
	NewPorts    int       `json:"new_ports"`    // Ports not in baseline
	MissingPorts int      `json:"missing_ports"` // Expected ports not found
}
