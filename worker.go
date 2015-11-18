package worker

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/admin"
)

func New(db *gorm.DB) *Worker {
	return &Worker{DB: db}
}

type Worker struct {
	Queue Queue
	DB    *gorm.DB
	Jobs  []*Job
}

func (worker *Worker) ConfigureQorResource(res *admin.Resource) {
	router := res.GetAdmin().GetRouter()
	controller := workerController{Worker: worker}

	router.Get("/"+res.ToParam(), controller.index)
	router.Get("/"+res.ToParam()+"/.*$", controller.show)
	router.Post("/"+res.ToParam()+"/.*/run$", controller.runJob)
	router.Post("/"+res.ToParam()+"/.*/kill$", controller.killJob)
}

func (worker *Worker) SetQueue(queue Queue) {
	worker.Queue = queue
}

func (worker *Worker) RegisterJob(job Job) error {
	worker.Jobs = append(worker.Jobs, &job)
	return nil
}

func (worker *Worker) GetJob(jobID uint) (*QorJob, error) {
	var qorJob QorJob

	if err := worker.DB.First(&qorJob, jobID).Error; err == nil {
		for _, job := range worker.Jobs {
			if job.Name == qorJob.Name {
				qorJob.Job = job
				return &qorJob, nil
			}
		}
	}
	return nil, fmt.Errorf("failed to find job: %v", jobID)
}

func (worker *Worker) AddJob(QorJob) error {
	return nil
}

func (worker *Worker) RunJob(jobID uint) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		return qorJob.Job.Run(qorJob.Argument)
	} else {
		return err
	}
}

func (worker *Worker) KillJob(jobID uint) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		return qorJob.GetQueue().Kill(qorJob)
	} else {
		return err
	}
}

func (worker *Worker) DeleteJob(jobID uint) error {
	if qorJob, err := worker.GetJob(jobID); err == nil {
		return qorJob.GetQueue().Delete(qorJob)
	} else {
		return err
	}
}
