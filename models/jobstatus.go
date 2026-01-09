package models

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

type JobStatus struct {
	gorm.Model
	Name     string `json:"name" gorm:"uniqueIndex"`
	JobIndex int    `json:"job_index"`
}

var jobMutex sync.Mutex

func (js *JobStatus) GetNextJob() (Job, error) {
	jobMutex.Lock()
	defer jobMutex.Unlock()

	db := GetDB()
	var teams []Team
	result := db.Order("name ASC").Find(&teams)
	if result.Error != nil {
		return Job{}, errors.New("database error loading teams")
	}

	if len(teams) == 0 {
		return Job{}, errors.New("no teams configured")
	}

	// Ensure index is within bounds
	if js.JobIndex >= len(teams) {
		js.JobIndex = 0
	}

	team := teams[js.JobIndex]
	job := MakeJob(js.Name, team.IPRange, team.TID, team.Name)

	// Increment and wrap index
	js.JobIndex = (js.JobIndex + 1) % len(teams)
	db.Save(&js)

	return job, nil
}
