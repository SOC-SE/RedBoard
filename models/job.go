package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model  `json:"-"`
	JID         string    `json:"jid" gorm:"uniqueIndex"`
	Type        string    `json:"type"`
	IPRange     string    `json:"iprange"`
	Status      string    `json:"status"` // queued, running, complete, failed
	Scanner     string    `json:"scanner"`
	TID         string    `json:"tid" gorm:"column:t_id;index"`
	TeamName    string    `json:"team_name"` // Denormalized for convenience
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	HostsFound  int       `json:"hosts_found"`
	PortsFound  int       `json:"ports_found"`
	ErrorMsg    string    `json:"error_msg"`
}

func MakeJob(jobtype string, iprange string, tid string, teamName string) Job {
	var job Job
	job.Type = jobtype
	job.IPRange = iprange
	job.JID = uuid.New().String()
	job.TID = tid
	job.TeamName = teamName
	job.Status = "queued"
	return job
}
