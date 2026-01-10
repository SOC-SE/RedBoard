package models

import "time"

type LoginReq struct {
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterReq struct {
	Name     string `json:"name" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

// Scan request structs - separate from GORM models for proper JSON binding
type ScanScriptResult struct {
	Name   string `json:"name"`
	Output string `json:"output"`
}

type ScanPort struct {
	Number   uint16             `json:"number"`
	State    string             `json:"state"`
	Protocol string             `json:"protocol"`
	Service  string             `json:"service"`
	Version  string             `json:"version"`
	Scripts  []ScanScriptResult `json:"scripts,omitempty"`
}

type ScanHost struct {
	IP       string     `json:"ip"`
	Hostname string     `json:"hostname"`
	OS       string     `json:"os"`
	Status   string     `json:"status"`
	Ports    []ScanPort `json:"ports"`
}

type Scan struct {
	Status    string     `json:"status"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	Hosts     []ScanHost `json:"hosts"`
}
