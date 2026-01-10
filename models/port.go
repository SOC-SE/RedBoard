package models

import (
	"gorm.io/gorm"
)

type ScriptResult struct {
	gorm.Model `json:"-"`
	PortID     uint   `json:"port_id" gorm:"index"`
	Name       string `json:"name"`
	Output     string `json:"output" gorm:"type:text"`
}

type Port struct {
	gorm.Model `json:"-"`
	Number     uint16         `json:"number"`
	State      string         `json:"state"`
	Protocol   string         `json:"protocol"`
	Service    string         `json:"service"`
	Version    string         `json:"version"`
	HostID     uint           `json:"host_id" gorm:"index"`
	IsBaseline bool           `json:"is_baseline"`
	IsNew      bool           `json:"is_new"`
	Scripts    []ScriptResult `json:"scripts,omitempty" gorm:"foreignKey:PortID;constraint:OnDelete:CASCADE"`
}

// Common dangerous ports for highlighting (removed DNS, SSH, HTTP-Proxy as they're expected)
var DangerousPorts = map[uint16]string{
	21:    "FTP",
	23:    "Telnet",
	25:    "SMTP",
	110:   "POP3",
	135:   "MSRPC",
	139:   "NetBIOS",
	143:   "IMAP",
	445:   "SMB",
	1433:  "MSSQL",
	1521:  "Oracle",
	3306:  "MySQL",
	3389:  "RDP",
	5432:  "PostgreSQL",
	5900:  "VNC",
	6379:  "Redis",
	27017: "MongoDB",
}

func (p *Port) IsDangerous() bool {
	_, exists := DangerousPorts[p.Number]
	return exists
}
