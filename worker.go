package worker

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func New(db *gorm.DB) *Worker {
	return &Worker{DB: db}
}

type Worker struct {
	DB   *gorm.DB
	Jobs []*Job
}

func (worker *Worker) AddJob(job Job) error {
	worker.Jobs = append(worker.Jobs, &job)
	return nil
}

type QorJob struct {
	gorm.Model
	Name     string
	Status   string
	Argument interface{}
}

func (worker *Worker) RunJob(id uint) error {
	var qorJob QorJob
	if !worker.DB.First(&qorJob, id).RecordNotFound() {
		for _, job := range worker.Jobs {
			if job.Name == qorJob.Name {
				return job.Run(qorJob.Argument)
			}
		}
	}

	return fmt.Errorf("failed to find job: %v", id)
}
