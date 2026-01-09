package models

import "gorm.io/gorm"

type Port struct {
	gorm.Model `json:"-"`
	Number     uint16 `json:"number"`
	State      string `json:"state"`
	Protocol   string `json:"protocol"`
	Service    string `json:"service"`
	Version    string `json:"version"`
	HostID     uint   `json:"host_id" gorm:"index"`
	IsBaseline bool   `json:"is_baseline"` // Whether this port is in the baseline
	IsNew      bool   `json:"is_new"`      // Whether this is a newly discovered port
}

// Common dangerous ports for highlighting
var DangerousPorts = map[uint16]string{
	21:    "FTP",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
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
	8080:  "HTTP-Proxy",
	27017: "MongoDB",
}

func (p *Port) IsDangerous() bool {
	_, exists := DangerousPorts[p.Number]
	return exists
}
